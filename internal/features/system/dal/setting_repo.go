package dal

import (
	"context"
	"fmt"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/setting"
	systemdto "origadmin/application/origstudio/internal/features/system/dto"
)

type SettingRepo struct {
	db *entity.Client
}

func NewSettingRepo(db *entity.Client) *SettingRepo {
	return &SettingRepo{db: db}
}

// EntityToSettingDTO converts an entity.Setting to a dto.SettingDTO.
func EntityToSettingDTO(s *entity.Setting) *systemdto.SettingDTO {
	if s == nil {
		return nil
	}
	return &systemdto.SettingDTO{
		ID:            s.ID,
		Key:           s.Key,
		Value:         s.Value,
		Type:          systemdto.SettingType(string(s.Type)),
		Category:      systemdto.SettingCategory(string(s.Category)),
		Description:   s.Description,
		IsSensitive:   s.IsSensitive,
		FallbackValue: s.FallbackValue,
		IsBuiltin:     s.IsBuiltin,
		CreateTime:    s.CreateTime,
		UpdateTime:    s.UpdateTime,
	}
}

// SettingDTOToEntityFields extracts fields from a SettingDTO for dal operations.
// Returns the field values needed for create/update operations.
func SettingDTOToEntityFields(s *systemdto.SettingDTO) (key, value string, typ setting.Type, category setting.Category, desc string, isSensitive bool, fallback string, isBuiltin bool) {
	if s == nil {
		return
	}
	return s.Key, s.Value, setting.Type(string(s.Type)), setting.Category(string(s.Category)),
		s.Description, s.IsSensitive, s.FallbackValue, s.IsBuiltin
}

func (r *SettingRepo) GetByKey(ctx context.Context, key string) (*systemdto.SettingDTO, error) {
	s, err := r.db.Setting.Query().
		Where(setting.KeyEQ(key)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get setting by key %s: %w", key, err)
	}
	return EntityToSettingDTO(s), nil
}

func (r *SettingRepo) ListByCategory(ctx context.Context, category string) ([]*systemdto.SettingDTO, error) {
	items, err := r.db.Setting.Query().
		Where(setting.CategoryEQ(setting.Category(category))).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list settings by category %s: %w", category, err)
	}
	result := make([]*systemdto.SettingDTO, len(items))
	for i, item := range items {
		result[i] = EntityToSettingDTO(item)
	}
	return result, nil
}

func (r *SettingRepo) ListAll(ctx context.Context) ([]*systemdto.SettingDTO, error) {
	items, err := r.db.Setting.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all settings: %w", err)
	}
	result := make([]*systemdto.SettingDTO, len(items))
	for i, item := range items {
		result[i] = EntityToSettingDTO(item)
	}
	return result, nil
}

func (r *SettingRepo) Upsert(ctx context.Context, s *systemdto.SettingDTO) (*systemdto.SettingDTO, error) {
	existing, err := r.db.Setting.Query().
		Where(setting.KeyEQ(s.Key)).
		Only(ctx)
	if err != nil {
		if entity.IsNotFound(err) {
			created, err := r.db.Setting.Create().
				SetKey(s.Key).
				SetValue(s.Value).
				SetType(setting.Type(string(s.Type))).
				SetCategory(setting.Category(string(s.Category))).
				SetDescription(s.Description).
				SetIsSensitive(s.IsSensitive).
				SetFallbackValue(s.FallbackValue).
				SetIsBuiltin(s.IsBuiltin).
				Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("create setting %s: %w", s.Key, err)
			}
			return EntityToSettingDTO(created), nil
		}
		return nil, fmt.Errorf("query setting %s: %w", s.Key, err)
	}

	updated, err := existing.Update().
		SetValue(s.Value).
		SetType(setting.Type(string(s.Type))).
		SetCategory(setting.Category(string(s.Category))).
		SetNillableDescription(&s.Description).
		SetIsSensitive(s.IsSensitive).
		SetNillableFallbackValue(&s.FallbackValue).
		SetIsBuiltin(s.IsBuiltin).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update setting %s: %w", s.Key, err)
	}
	return EntityToSettingDTO(updated), nil
}

func (r *SettingRepo) Delete(ctx context.Context, id string) error {
	err := r.db.Setting.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete setting %s: %w", id, err)
	}
	return nil
}

func (r *SettingRepo) ResetToDefault(ctx context.Context, key string) (*systemdto.SettingDTO, error) {
	s, err := r.db.Setting.Query().
		Where(setting.KeyEQ(key)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get setting for reset %s: %w", key, err)
	}
	if s.FallbackValue == "" {
		return EntityToSettingDTO(s), nil
	}
	updated, err := s.Update().
		SetValue(s.FallbackValue).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("reset setting %s: %w", key, err)
	}
	return EntityToSettingDTO(updated), nil
}

func (r *SettingRepo) SeedDefaults(ctx context.Context, defaults []*systemdto.SettingDTO) error {
	for _, s := range defaults {
		existing, err := r.db.Setting.Query().
			Where(setting.KeyEQ(s.Key)).
			Only(ctx)
		if err != nil {
			if entity.IsNotFound(err) {
				_, err = r.db.Setting.Create().
					SetKey(s.Key).
					SetValue(s.Value).
					SetType(setting.Type(string(s.Type))).
					SetCategory(setting.Category(string(s.Category))).
					SetDescription(s.Description).
					SetIsSensitive(s.IsSensitive).
					SetFallbackValue(s.FallbackValue).
					SetIsBuiltin(s.IsBuiltin).
					Save(ctx)
				if err != nil {
					return fmt.Errorf("seed setting %s: %w", s.Key, err)
				}
			} else {
				return fmt.Errorf("check setting %s: %w", s.Key, err)
			}
			continue
		}
		if existing.Value == "" && s.Value != "" {
			_, err = existing.Update().
				SetValue(s.Value).
				SetFallbackValue(s.FallbackValue).
				SetDescription(s.Description).
				SetType(setting.Type(string(s.Type))).
				SetIsSensitive(s.IsSensitive).
				SetIsBuiltin(s.IsBuiltin).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("update seed setting %s: %w", s.Key, err)
			}
		}
	}
	return nil
}
