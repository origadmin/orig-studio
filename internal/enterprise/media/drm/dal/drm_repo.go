package dal

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/drmkey"
	"origadmin/application/origstudio/internal/dal/entity/drmpolicy"
	"origadmin/application/origstudio/internal/dal/entity/drmlicense"
	"origadmin/application/origstudio/internal/enterprise/media/drm/biz"
	"origadmin/application/origstudio/internal/enterprise/media/drm/dto"
)

type repo struct {
	db  *entity.Client
	log *log.Helper
}

func NewRepo(db *entity.Client, logger log.Logger) biz.Repo {
	return &repo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "enterprise/drm.repo")),
	}
}

func entityToDrmPolicyDTO(e *entity.DrmPolicy) *dto.DrmPolicyDTO {
	if e == nil {
		return nil
	}
	return &dto.DrmPolicyDTO{
		ID:              e.ID,
		Name:            e.Name,
		Type:            string(e.Type),
		HlsKeyURL:       e.HlsKeyURL,
		WidevinePssh:    e.WidevinePssh,
		FairplayCertURL: e.FairplayCertURL,
		IsDefault:       e.IsDefault,
		Description:     e.Description,
		CreateTime:      e.CreateTime,
		UpdateTime:      e.UpdateTime,
	}
}

func entityToDrmKeyDTO(e *entity.DrmKey) *dto.DrmKeyDTO {
	if e == nil {
		return nil
	}
	return &dto.DrmKeyDTO{
		ID:        e.ID,
		PolicyID:  e.PolicyID,
		ContentID: e.ContentID,
		KeyID:     e.KeyID,
		Iv:        e.Iv,
		CreatedAt: e.CreatedAt,
		ExpiresAt: e.ExpiresAt,
	}
}

func entityToDrmLicenseDTO(e *entity.DrmLicense) *dto.DrmLicenseDTO {
	if e == nil {
		return nil
	}
	return &dto.DrmLicenseDTO{
		ID:        e.ID,
		KeyID:     e.KeyID,
		UserID:    e.UserID,
		DeviceID:  e.DeviceID,
		Status:    string(e.Status),
		IssuedAt:  e.IssuedAt,
		ExpiresAt: e.ExpiresAt,
	}
}

func (r *repo) ListPolicies(ctx context.Context) ([]*dto.DrmPolicyDTO, error) {
	items, err := r.db.DrmPolicy.Query().
		Order(entity.Desc(drmpolicy.FieldCreateTime)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list drm policies: %w", err)
	}
	result := make([]*dto.DrmPolicyDTO, len(items))
	for i, item := range items {
		result[i] = entityToDrmPolicyDTO(item)
	}
	return result, nil
}

func (r *repo) GetPolicyByID(ctx context.Context, id string) (*dto.DrmPolicyDTO, error) {
	item, err := r.db.DrmPolicy.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get drm policy: %w", err)
	}
	return entityToDrmPolicyDTO(item), nil
}

func (r *repo) CreatePolicy(ctx context.Context, p *dto.DrmPolicyDTO) (*dto.DrmPolicyDTO, error) {
	builder := r.db.DrmPolicy.Create().
		SetName(p.Name).
		SetType(drmpolicy.Type(p.Type)).
		SetIsDefault(p.IsDefault)
	if p.HlsKeyURL != "" {
		builder.SetHlsKeyURL(p.HlsKeyURL)
	}
	if p.WidevinePssh != "" {
		builder.SetWidevinePssh(p.WidevinePssh)
	}
	if p.FairplayCertURL != "" {
		builder.SetFairplayCertURL(p.FairplayCertURL)
	}
	if p.Description != "" {
		builder.SetDescription(p.Description)
	}
	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create drm policy: %w", err)
	}
	return entityToDrmPolicyDTO(ent), nil
}

func (r *repo) UpdatePolicy(ctx context.Context, p *dto.DrmPolicyDTO) (*dto.DrmPolicyDTO, error) {
	builder := r.db.DrmPolicy.UpdateOneID(p.ID).
		SetName(p.Name).
		SetType(drmpolicy.Type(p.Type)).
		SetIsDefault(p.IsDefault)
	builder.SetHlsKeyURL(p.HlsKeyURL)
	builder.SetWidevinePssh(p.WidevinePssh)
	builder.SetFairplayCertURL(p.FairplayCertURL)
	builder.SetDescription(p.Description)
	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update drm policy: %w", err)
	}
	return entityToDrmPolicyDTO(ent), nil
}

func (r *repo) DeletePolicy(ctx context.Context, id string) error {
	return r.db.DrmPolicy.DeleteOneID(id).Exec(ctx)
}

func (r *repo) ListKeysByPolicy(ctx context.Context, policyID string) ([]*dto.DrmKeyDTO, error) {
	items, err := r.db.DrmKey.Query().
		Where(drmkey.PolicyIDEQ(policyID)).
		Order(entity.Desc(drmkey.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list drm keys: %w", err)
	}
	result := make([]*dto.DrmKeyDTO, len(items))
	for i, item := range items {
		result[i] = entityToDrmKeyDTO(item)
	}
	return result, nil
}

func (r *repo) CreateKey(ctx context.Context, k *dto.DrmKeyDTO) (*dto.DrmKeyDTO, error) {
	builder := r.db.DrmKey.Create().
		SetPolicyID(k.PolicyID).
		SetContentID(k.ContentID).
		SetKeyID(k.KeyID).
		SetKeyValue(k.KeyID)
	if k.Iv != "" {
		builder.SetIv(k.Iv)
	}
	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create drm key: %w", err)
	}
	return entityToDrmKeyDTO(ent), nil
}

func (r *repo) DeleteKey(ctx context.Context, id string) error {
	return r.db.DrmKey.DeleteOneID(id).Exec(ctx)
}

func (r *repo) ListLicenses(ctx context.Context, page, pageSize int) ([]*dto.DrmLicenseDTO, int, error) {
	query := r.db.DrmLicense.Query()
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count drm licenses: %w", err)
	}
	items, err := query.
		Order(entity.Desc(drmlicense.FieldIssuedAt)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list drm licenses: %w", err)
	}
	result := make([]*dto.DrmLicenseDTO, len(items))
	for i, item := range items {
		result[i] = entityToDrmLicenseDTO(item)
	}
	return result, total, nil
}