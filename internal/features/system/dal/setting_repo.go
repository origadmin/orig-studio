package dal

import (
	"context"
	"fmt"

	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/setting"
)

type SettingRepo struct {
	db *entity.Client
}

func NewSettingRepo(db *entity.Client) *SettingRepo {
	return &SettingRepo{db: db}
}

func (r *SettingRepo) GetByKey(ctx context.Context, key string) (*entity.Setting, error) {
	s, err := r.db.Setting.Query().
		Where(setting.KeyEQ(key)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get setting by key %s: %w", key, err)
	}
	return s, nil
}

func (r *SettingRepo) ListByCategory(ctx context.Context, category string) ([]*entity.Setting, error) {
	items, err := r.db.Setting.Query().
		Where(setting.CategoryEQ(setting.Category(category))).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list settings by category %s: %w", category, err)
	}
	return items, nil
}

func (r *SettingRepo) ListAll(ctx context.Context) ([]*entity.Setting, error) {
	items, err := r.db.Setting.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all settings: %w", err)
	}
	return items, nil
}

func (r *SettingRepo) Upsert(ctx context.Context, s *entity.Setting) (*entity.Setting, error) {
	existing, err := r.db.Setting.Query().
		Where(setting.KeyEQ(s.Key)).
		Only(ctx)
	if err != nil {
		if entity.IsNotFound(err) {
			created, err := r.db.Setting.Create().
				SetKey(s.Key).
				SetValue(s.Value).
				SetType(s.Type).
				SetCategory(s.Category).
				SetDescription(s.Description).
				SetIsSensitive(s.IsSensitive).
				SetFallbackValue(s.FallbackValue).
				SetIsBuiltin(s.IsBuiltin).
				Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("create setting %s: %w", s.Key, err)
			}
			return created, nil
		}
		return nil, fmt.Errorf("query setting %s: %w", s.Key, err)
	}

	updated, err := existing.Update().
		SetValue(s.Value).
		SetType(s.Type).
		SetCategory(s.Category).
		SetNillableDescription(&s.Description).
		SetIsSensitive(s.IsSensitive).
		SetNillableFallbackValue(&s.FallbackValue).
		SetIsBuiltin(s.IsBuiltin).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update setting %s: %w", s.Key, err)
	}
	return updated, nil
}

func (r *SettingRepo) Delete(ctx context.Context, id string) error {
	err := r.db.Setting.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete setting %s: %w", id, err)
	}
	return nil
}

func (r *SettingRepo) SeedDefaults(ctx context.Context, defaults []*entity.Setting) error {
	for _, s := range defaults {
		existing, err := r.db.Setting.Query().
			Where(setting.KeyEQ(s.Key)).
			Only(ctx)
		if err != nil {
			if entity.IsNotFound(err) {
				_, err = r.db.Setting.Create().
					SetKey(s.Key).
					SetValue(s.Value).
					SetType(s.Type).
					SetCategory(s.Category).
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
				SetType(s.Type).
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
