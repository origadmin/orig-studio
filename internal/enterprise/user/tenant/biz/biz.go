package biz

import (
	"context"
	"errors"
	"fmt"

	"origadmin/application/origstudio/internal/dal/entity"
	tenantpred "origadmin/application/origstudio/internal/dal/entity/tenant"
	tenantctx "origadmin/application/origstudio/internal/infra/tenant"
	"origadmin/application/origstudio/internal/enterprise/user/tenant/dto"
)

type UseCase struct {
	client *entity.Client
}

func NewUseCase(client *entity.Client) *UseCase {
	return &UseCase{client: client}
}

func (uc *UseCase) Create(ctx context.Context, t *dto.TenantDTO) (*dto.TenantDTO, error) {
	if t.Slug == "" {
		return nil, errors.New("slug is required")
	}
	if t.Name == "" {
		return nil, errors.New("name is required")
	}

	builder := uc.client.Tenant.Create().
		SetName(t.Name).
		SetSlug(t.Slug)

	if t.Domain != "" {
		builder.SetDomain(t.Domain)
	}
	if t.Logo != "" {
		builder.SetLogo(t.Logo)
	}
	if t.Description != "" {
		builder.SetDescription(t.Description)
	}
	if t.Status != "" {
		builder.SetStatus(tenantpred.Status(t.Status))
	}
	if t.Plan != "" {
		builder.SetPlan(tenantpred.Plan(t.Plan))
	}
	if t.MaxUsers > 0 {
		builder.SetMaxUsers(t.MaxUsers)
	}
	if t.MaxStorage > 0 {
		builder.SetMaxStorageMB(t.MaxStorage)
	}
	if t.Config != nil {
		builder.SetConfig(t.Config)
	}

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create tenant: %w", err)
	}

	return EntityToDTO(ent), nil
}

func (uc *UseCase) Get(ctx context.Context, id string) (*dto.TenantDTO, error) {
	ent, err := uc.client.Tenant.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get tenant: %w", err)
	}
	return EntityToDTO(ent), nil
}

func (uc *UseCase) GetBySlug(ctx context.Context, slug string) (*dto.TenantDTO, error) {
	ent, err := uc.client.Tenant.Query().
		Where(tenantpred.Slug(slug)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tenant by slug: %w", err)
	}
	return EntityToDTO(ent), nil
}

func (uc *UseCase) GetByDomain(ctx context.Context, domain string) (*dto.TenantDTO, error) {
	ent, err := uc.client.Tenant.Query().
		Where(tenantpred.DomainEQ(domain)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tenant by domain: %w", err)
	}
	return EntityToDTO(ent), nil
}

func (uc *UseCase) List(ctx context.Context, page, pageSize int) ([]*dto.TenantDTO, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	total, err := uc.client.Tenant.Query().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count tenants: %w", err)
	}

	tenants, err := uc.client.Tenant.Query().
		Offset(offset).
		Limit(pageSize).
		Order(entity.Asc(tenantpred.FieldCreateTime)).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list tenants: %w", err)
	}

	dtos := make([]*dto.TenantDTO, len(tenants))
	for i, ent := range tenants {
		dtos[i] = EntityToDTO(ent)
	}

	return dtos, total, nil
}

func (uc *UseCase) Update(ctx context.Context, id string, t *dto.TenantDTO) (*dto.TenantDTO, error) {
	builder := uc.client.Tenant.UpdateOneID(id)

	if t.Name != "" {
		builder.SetName(t.Name)
	}
	if t.Domain != "" {
		builder.SetDomain(t.Domain)
	}
	if t.Logo != "" {
		builder.SetLogo(t.Logo)
	}
	if t.Description != "" {
		builder.SetDescription(t.Description)
	}
	if t.Status != "" {
		builder.SetStatus(tenantpred.Status(t.Status))
	}
	if t.Plan != "" {
		builder.SetPlan(tenantpred.Plan(t.Plan))
	}
	if t.MaxUsers > 0 {
		builder.SetMaxUsers(t.MaxUsers)
	}
	if t.MaxStorage > 0 {
		builder.SetMaxStorageMB(t.MaxStorage)
	}
	if t.Config != nil {
		builder.SetConfig(t.Config)
	}

	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update tenant: %w", err)
	}

	return EntityToDTO(ent), nil
}

func (uc *UseCase) Delete(ctx context.Context, id string) error {
	return uc.client.Tenant.DeleteOneID(id).Exec(ctx)
}

func (uc *UseCase) ResolveFromContext(ctx context.Context) (*dto.TenantDTO, error) {
	tc := tenantctx.FromContext(ctx)
	if tc == nil || tc.IsSystem {
		return nil, nil
	}

	if tc.ID != "" {
		return uc.Get(ctx, tc.ID)
	}

	if tc.Slug != "" {
		return uc.GetBySlug(ctx, tc.Slug)
	}

	return nil, nil
}

func EntityToDTO(ent *entity.Tenant) *dto.TenantDTO {
	return &dto.TenantDTO{
		ID:          ent.ID,
		Name:        ent.Name,
		Slug:        ent.Slug,
		Domain:      ent.Domain,
		Logo:        ent.Logo,
		Description: ent.Description,
		Status:      string(ent.Status),
		Plan:        string(ent.Plan),
		MaxUsers:    ent.MaxUsers,
		MaxStorage:  ent.MaxStorageMB,
		Config:      ent.Config,
	}
}