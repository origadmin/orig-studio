/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dto

import (
	"context"
)

// EncodeProfileRepo defines the storage operations for encode profiles.
type EncodeProfileRepo interface {
	ListActive(ctx context.Context) ([]*EncodeProfile, error)
	ListAll(ctx context.Context) ([]*EncodeProfile, error)
	Get(ctx context.Context, id int) (*EncodeProfile, error)
	GetByName(ctx context.Context, name string) (*EncodeProfile, error)
	Create(ctx context.Context, profile *EncodeProfile) (*EncodeProfile, error)
	Update(ctx context.Context, profile *EncodeProfile) (*EncodeProfile, error)
	Delete(ctx context.Context, id int) error
}
