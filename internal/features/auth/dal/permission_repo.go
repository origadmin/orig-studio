package dal

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/groupmember"
	"origadmin/application/origstudio/internal/dal/entity/permissiongroup"
	"origadmin/application/origstudio/internal/features/auth/biz"
)

type Data struct {
	db *entity.Client
}

func NewData(db *entity.Client) *Data {
	return &Data{db: db}
}

type permissionGroupRepo struct {
	data *Data
	log  *log.Helper
}

func NewPermissionGroupRepo(data *Data, logger log.Logger) biz.PermissionGroupRepo {
	return &permissionGroupRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "svc-auth/permission_group.data")),
	}
}

func (r *permissionGroupRepo) Create(ctx context.Context, name, description string, permissions, categoryScope []string, createdBy string) (*biz.PermissionGroupItem, error) {
	builder := r.data.db.PermissionGroup.Create().
		SetName(name).
		SetDescription(description).
		SetPermissions(permissions).
		SetIsActive(true).
		SetCreatedBy(createdBy)

	if len(categoryScope) > 0 {
		builder.SetCategoryScope(categoryScope)
	}

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission group: %w", err)
	}

	return mapPermissionGroupToItem(ent, 0), nil
}

func (r *permissionGroupRepo) Get(ctx context.Context, id string) (*biz.PermissionGroupItem, error) {
	ent, err := r.data.db.PermissionGroup.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission group: %w", err)
	}

	memberCount, err := r.data.db.GroupMember.Query().
		Where(groupmember.GroupIDEQ(id)).
		Count(ctx)
	if err != nil {
		r.log.Warnf("failed to count members for group %s: %v", id, err)
		memberCount = 0
	}

	return mapPermissionGroupToItem(ent, memberCount), nil
}

func (r *permissionGroupRepo) Update(ctx context.Context, id, name, description string, permissions, categoryScope []string) (*biz.PermissionGroupItem, error) {
	builder := r.data.db.PermissionGroup.UpdateOneID(id).
		SetName(name).
		SetDescription(description).
		SetPermissions(permissions)

	if len(categoryScope) > 0 {
		builder.SetCategoryScope(categoryScope)
	} else {
		builder.ClearCategoryScope()
	}

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update permission group: %w", err)
	}

	memberCount, err := r.data.db.GroupMember.Query().
		Where(groupmember.GroupIDEQ(id)).
		Count(ctx)
	if err != nil {
		r.log.Warnf("failed to count members for group %s: %v", id, err)
		memberCount = 0
	}

	return mapPermissionGroupToItem(ent, memberCount), nil
}

func (r *permissionGroupRepo) Delete(ctx context.Context, id string) error {
	_, err := r.data.db.GroupMember.Delete().
		Where(groupmember.GroupIDEQ(id)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete group members: %w", err)
	}

	err = r.data.db.PermissionGroup.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete permission group: %w", err)
	}

	return nil
}

func (r *permissionGroupRepo) Toggle(ctx context.Context, id string, isActive bool) error {
	_, err := r.data.db.PermissionGroup.UpdateOneID(id).
		SetIsActive(isActive).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to toggle permission group: %w", err)
	}
	return nil
}

func (r *permissionGroupRepo) List(ctx context.Context, isActive *bool, page, pageSize int) ([]*biz.PermissionGroupItem, int, error) {
	query := r.data.db.PermissionGroup.Query()

	if isActive != nil {
		query.Where(permissiongroup.IsActiveEQ(*isActive))
	}

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count permission groups: %w", err)
	}

	ents, err := query.
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order(entity.Desc(permissiongroup.FieldCreateTime)).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list permission groups: %w", err)
	}

	items := make([]*biz.PermissionGroupItem, len(ents))
	for i, ent := range ents {
		memberCount, cntErr := r.data.db.GroupMember.Query().
			Where(groupmember.GroupIDEQ(ent.ID)).
			Count(ctx)
		if cntErr != nil {
			r.log.Warnf("failed to count members for group %s: %v", ent.ID, cntErr)
			memberCount = 0
		}
		items[i] = mapPermissionGroupToItem(ent, memberCount)
	}

	return items, total, nil
}

func (r *permissionGroupRepo) GetMemberIDs(ctx context.Context, groupID string) ([]string, error) {
	members, err := r.data.db.GroupMember.Query().
		Where(groupmember.GroupIDEQ(groupID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get member IDs: %w", err)
	}

	userIDs := make([]string, len(members))
	for i, m := range members {
		userIDs[i] = m.UserID
	}
	return userIDs, nil
}

type groupMemberRepo struct {
	data *Data
	log  *log.Helper
}

func NewGroupMemberRepo(data *Data, logger log.Logger) biz.GroupMemberRepo {
	return &groupMemberRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "svc-auth/group_member.data")),
	}
}

func (r *groupMemberRepo) AddMembers(ctx context.Context, groupID string, userIDs []string) (added int, skipped int, err error) {
	for _, uid := range userIDs {
		_, createErr := r.data.db.GroupMember.Create().
			SetUserID(uid).
			SetGroupID(groupID).
			Save(ctx)
		if createErr != nil {
			if entity.IsConstraintError(createErr) {
				skipped++
				continue
			}
			return added, skipped, fmt.Errorf("failed to add member %s to group %s: %w", uid, groupID, createErr)
		}
		added++
	}
	return added, skipped, nil
}

func (r *groupMemberRepo) RemoveMember(ctx context.Context, groupID, userID string) error {
	member, err := r.data.db.GroupMember.Query().
		Where(
			groupmember.UserIDEQ(userID),
			groupmember.GroupIDEQ(groupID),
		).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("failed to find group member: %w", err)
	}

	err = r.data.db.GroupMember.DeleteOneID(member.ID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove group member: %w", err)
	}
	return nil
}

func (r *groupMemberRepo) ListByGroup(ctx context.Context, groupID string, page, pageSize int) ([]*biz.GroupMemberItem, int, error) {
	query := r.data.db.GroupMember.Query().
		Where(groupmember.GroupIDEQ(groupID))

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count group members: %w", err)
	}

	ents, err := query.
		WithUser().
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order(entity.Desc(groupmember.FieldJoinedAt)).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list group members: %w", err)
	}

	items := make([]*biz.GroupMemberItem, len(ents))
	for i, ent := range ents {
		items[i] = mapGroupMemberToItem(ent)
	}
	return items, total, nil
}

func (r *groupMemberRepo) ListByUser(ctx context.Context, userID string) ([]*biz.GroupMemberItem, error) {
	ents, err := r.data.db.GroupMember.Query().
		Where(groupmember.UserIDEQ(userID)).
		WithGroup().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list user group memberships: %w", err)
	}

	items := make([]*biz.GroupMemberItem, len(ents))
	for i, ent := range ents {
		items[i] = mapGroupMemberToItem(ent)
	}
	return items, nil
}

type userPermRepo struct {
	data *Data
	log  *log.Helper
}

func NewUserPermRepo(data *Data, logger log.Logger) biz.UserPermRepo {
	return &userPermRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "svc-auth/user_perm.data")),
	}
}

func (r *userPermRepo) GetUserRole(ctx context.Context, userID string) (string, error) {
	u, err := r.data.db.User.Get(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user role: %w", err)
	}
	return string(u.Role), nil
}

func mapPermissionGroupToItem(ent *entity.PermissionGroup, memberCount int) *biz.PermissionGroupItem {
	item := &biz.PermissionGroupItem{
		ID:          ent.ID,
		Name:        ent.Name,
		Description: ent.Description,
		Permissions: ent.Permissions,
		IsActive:    ent.IsActive,
		CreatedBy:   ent.CreatedBy,
		MemberCount: memberCount,
		CreateTime:   ent.CreateTime,
		UpdateTime:   ent.UpdateTime,
	}

	if len(ent.CategoryScope) > 0 {
		item.CategoryScope = ent.CategoryScope
	}

	return item
}

func mapGroupMemberToItem(ent *entity.GroupMember) *biz.GroupMemberItem {
	item := &biz.GroupMemberItem{
		ID:       ent.ID,
		UserID:   ent.UserID,
		GroupID:  ent.GroupID,
		JoinedAt: ent.JoinedAt,
	}

	if ent.Edges.User != nil {
		item.Username = ent.Edges.User.Username
	}

	return item
}
