package dal

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/groupmember"
	"origadmin/application/origstudio/internal/data/entity/permissiongroup"
	"origadmin/application/origstudio/internal/enterprise/auth/permission/biz"
	"origadmin/application/origstudio/internal/enterprise/auth/permission/dto"
)

type groupRepo struct {
	db  *entity.Client
	log *log.Helper
}

func NewGroupRepo(db *entity.Client, logger log.Logger) biz.GroupRepo {
	return &groupRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "enterprise/permission.group_repo")),
	}
}

func (r *groupRepo) Create(ctx context.Context, name, description string, permissions, categoryScope []string, createdBy string) (*dto.GroupItem, error) {
	builder := r.db.PermissionGroup.Create().
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
	return mapGroupToItem(ent, 0), nil
}

func (r *groupRepo) Get(ctx context.Context, id string) (*dto.GroupItem, error) {
	ent, err := r.db.PermissionGroup.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission group: %w", err)
	}
	memberCount, err := r.db.GroupMember.Query().
		Where(groupmember.GroupIDEQ(id)).
		Count(ctx)
	if err != nil {
		r.log.Warnf("failed to count members for group %s: %v", id, err)
		memberCount = 0
	}
	return mapGroupToItem(ent, memberCount), nil
}

func (r *groupRepo) Update(ctx context.Context, id, name, description string, permissions, categoryScope []string) (*dto.GroupItem, error) {
	builder := r.db.PermissionGroup.UpdateOneID(id).
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
	memberCount, err := r.db.GroupMember.Query().
		Where(groupmember.GroupIDEQ(id)).
		Count(ctx)
	if err != nil {
		r.log.Warnf("failed to count members for group %s: %v", id, err)
		memberCount = 0
	}
	return mapGroupToItem(ent, memberCount), nil
}

func (r *groupRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.GroupMember.Delete().
		Where(groupmember.GroupIDEQ(id)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete group members: %w", err)
	}
	err = r.db.PermissionGroup.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete permission group: %w", err)
	}
	return nil
}

func (r *groupRepo) Toggle(ctx context.Context, id string, isActive bool) error {
	_, err := r.db.PermissionGroup.UpdateOneID(id).
		SetIsActive(isActive).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to toggle permission group: %w", err)
	}
	return nil
}

func (r *groupRepo) List(ctx context.Context, isActive *bool, page, pageSize int) ([]*dto.GroupItem, int, error) {
	query := r.db.PermissionGroup.Query()
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
	items := make([]*dto.GroupItem, len(ents))
	for i, ent := range ents {
		memberCount, cntErr := r.db.GroupMember.Query().
			Where(groupmember.GroupIDEQ(ent.ID)).
			Count(ctx)
		if cntErr != nil {
			r.log.Warnf("failed to count members for group %s: %v", ent.ID, cntErr)
			memberCount = 0
		}
		items[i] = mapGroupToItem(ent, memberCount)
	}
	return items, total, nil
}

func (r *groupRepo) GetMemberIDs(ctx context.Context, groupID string) ([]string, error) {
	members, err := r.db.GroupMember.Query().
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

type memberRepo struct {
	db  *entity.Client
	log *log.Helper
}

func NewMemberRepo(db *entity.Client, logger log.Logger) biz.MemberRepo {
	return &memberRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "enterprise/permission.member_repo")),
	}
}

func (r *memberRepo) AddMembers(ctx context.Context, groupID string, userIDs []string) (added int, skipped int, err error) {
	for _, uid := range userIDs {
		_, createErr := r.db.GroupMember.Create().
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

func (r *memberRepo) RemoveMember(ctx context.Context, groupID, userID string) error {
	member, err := r.db.GroupMember.Query().
		Where(
			groupmember.UserIDEQ(userID),
			groupmember.GroupIDEQ(groupID),
		).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("failed to find group member: %w", err)
	}
	err = r.db.GroupMember.DeleteOneID(member.ID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove group member: %w", err)
	}
	return nil
}

func (r *memberRepo) ListByGroup(ctx context.Context, groupID string, page, pageSize int) ([]*dto.MemberItem, int, error) {
	query := r.db.GroupMember.Query().
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
	items := make([]*dto.MemberItem, len(ents))
	for i, ent := range ents {
		items[i] = mapMemberToItem(ent)
	}
	return items, total, nil
}

func (r *memberRepo) ListByUser(ctx context.Context, userID string) ([]*dto.MemberItem, error) {
	ents, err := r.db.GroupMember.Query().
		Where(groupmember.UserIDEQ(userID)).
		WithGroup().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list user group memberships: %w", err)
	}
	items := make([]*dto.MemberItem, len(ents))
	for i, ent := range ents {
		items[i] = mapMemberToItem(ent)
	}
	return items, nil
}

type userPermRepo struct {
	db  *entity.Client
	log *log.Helper
}

func NewUserPermRepo(db *entity.Client, logger log.Logger) biz.UserPermRepo {
	return &userPermRepo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "enterprise/permission.user_perm_repo")),
	}
}

func (r *userPermRepo) GetUserRole(ctx context.Context, userID string) (string, error) {
	u, err := r.db.User.Get(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user role: %w", err)
	}
	return string(u.Role), nil
}

func mapGroupToItem(ent *entity.PermissionGroup, memberCount int) *dto.GroupItem {
	item := &dto.GroupItem{
		ID:          ent.ID,
		Name:        ent.Name,
		Description: ent.Description,
		Permissions: ent.Permissions,
		IsActive:    ent.IsActive,
		CreatedBy:   ent.CreatedBy,
		MemberCount: memberCount,
		CreateTime:  ent.CreateTime,
		UpdateTime:  ent.UpdateTime,
	}
	if len(ent.CategoryScope) > 0 {
		item.CategoryScope = ent.CategoryScope
	}
	return item
}

func mapMemberToItem(ent *entity.GroupMember) *dto.MemberItem {
	item := &dto.MemberItem{
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