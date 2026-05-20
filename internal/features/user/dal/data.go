/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package data provides the data access layer for svc-user.
package dal

import (
	"origadmin/application/origstudio/internal/dal/entity"
)

// CategoryEnt is the component category for Ent database.
const CategoryEnt = "ent"

// Data encapsulates the core data access components.
type Data struct {
	DB *entity.Client
}

// NewData creates a new Data instance.
func NewData(database *entity.Client) (*Data, error) {
	return &Data{DB: database}, nil
}
