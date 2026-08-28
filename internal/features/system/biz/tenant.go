package biz

import (
	"context"
	"errors"
	"fmt"

	"origadmin/application/origstudio/internal/data/entity"
	tenantpred "origadmin/application/origstudio/internal/data/entity/tenant"
	tenantctx "origadmin/application/origstudio/internal/infra/tenant"
)

type TenantUseCase struct {
	client *entity.Client
}

func NewTenantUseCase(client *entity.Client) *TenantUseCase {
	return &TenantUseCase{client: client}
}

type TenantDTO struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Domain      string                 `json:"domain,omitempty"`
	Logo        string                 `json:"logo,omitempty"`
	Description string                 `json:"description,omitempty"`
	Status      string                 `json:"status"`
	Plan        string                 `json:"plan"`
	MaxUsers    int                    `json:"max_users"`
	MaxStorage  int                    `json:"max_storage_mb"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

func (uc *TenantUseCase) Create(ctx context.Context, dto *TenantDTO) (*TenantDTO, error) {
	if dto.Slug == "" {
		return nil, errors.New("slug is required")
	}
	if dto.Name == "" {
		return nil, errors.New("name is required")
	}

	builder := uc.client.Tenant.Create().
		SetName(dto.Name).
		SetSlug(dto.Slug)

	if dto.Domain != "" {
		builder.SetDomain(dto.Domain)
	}
	if dto.Logo != "" {
		builder.SetLogo(dto.Logo)
	}
	if dto.Description != "" {
		builder.SetDescription(dto.Description)
	}
	if dto.Status != "" {
		builder.SetStatus(tenantpred.Status(dto.Status))
	}
	if dto.Plan != "" {
		builder.SetPlan(tenantpred.Plan(dto.Plan))
	}
	if dto.MaxUsers > 0 {
		builder.SetMaxUsers(dto.MaxUsers)
	}
	if dto.MaxStorage > 0 {
		builder.SetMaxStorageMB(dto.MaxStorage)
	}
	if dto.Config != nil {
		builder.SetConfig(dto.Config)
	}

	t, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create tenant: %w", err)
	}

	return entityToDTO(t), nil
}

func (uc *TenantUseCase) Get(ctx context.Context, id string) (*TenantDTO, error) {
	t, err := uc.client.Tenant.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get tenant: %w", err)
	}
	return entityToDTO(t), nil
}

func (uc *TenantUseCase) GetBySlug(ctx context.Context, slug string) (*TenantDTO, error) {
	t, err := uc.client.Tenant.Query().
		Where(tenantpred.Slug(slug)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tenant by slug: %w", err)
	}
	return entityToDTO(t), nil
}

func (uc *TenantUseCase) GetByDomain(ctx context.Context, domain string) (*TenantDTO, error) {
	t, err := uc.client.Tenant.Query().
		Where(tenantpred.DomainEQ(domain)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tenant by domain: %w", err)
	}
	return entityToDTO(t), nil
}

func (uc *TenantUseCase) List(ctx context.Context, page, pageSize int) ([]*TenantDTO, int, error) {
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

	dtos := make([]*TenantDTO, len(tenants))
	for i, t := range tenants {
		dtos[i] = entityToDTO(t)
	}

	return dtos, total, nil
}

func (uc *TenantUseCase) Update(ctx context.Context, id string, dto *TenantDTO) (*TenantDTO, error) {
	builder := uc.client.Tenant.UpdateOneID(id)

	if dto.Name != "" {
		builder.SetName(dto.Name)
	}
	if dto.Domain != "" {
		builder.SetDomain(dto.Domain)
	}
	if dto.Logo != "" {
		builder.SetLogo(dto.Logo)
	}
	if dto.Description != "" {
		builder.SetDescription(dto.Description)
	}
	if dto.Status != "" {
		builder.SetStatus(tenantpred.Status(dto.Status))
	}
	if dto.Plan != "" {
		builder.SetPlan(tenantpred.Plan(dto.Plan))
	}
	if dto.MaxUsers > 0 {
		builder.SetMaxUsers(dto.MaxUsers)
	}
	if dto.MaxStorage > 0 {
		builder.SetMaxStorageMB(dto.MaxStorage)
	}
	if dto.Config != nil {
		builder.SetConfig(dto.Config)
	}

	t, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update tenant: %w", err)
	}

	return entityToDTO(t), nil
}

func (uc *TenantUseCase) Delete(ctx context.Context, id string) error {
	return uc.client.Tenant.DeleteOneID(id).Exec(ctx)
}

func (uc *TenantUseCase) ResolveFromContext(ctx context.Context) (*TenantDTO, error) {
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

func entityToDTO(t *entity.Tenant) *TenantDTO {
	return &TenantDTO{
		ID:          t.ID,
		Name:        t.Name,
		Slug:        t.Slug,
		Domain:      t.Domain,
		Logo:        t.Logo,
		Description: t.Description,
		Status:      string(t.Status),
		Plan:        string(t.Plan),
		MaxUsers:    t.MaxUsers,
		MaxStorage:  t.MaxStorageMB,
		Config:      t.Config,
	}
}
