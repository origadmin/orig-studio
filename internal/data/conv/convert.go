/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package conv implements the functions, types, and interfaces for the module.
package conv

//go:generate abgen .

//go:abgen:package:path=origadmin/application/origcms/internal/data/entity,alias=entity
//go:abgen:package:path=origadmin/application/origcms/api/gen/v1/types,alias=types
//go:abgen:pair:packages="entity,types"
//go:abgen:convert:direction="both"
//go:abgen:convert:source:suffix=""
//go:abgen:convert:target:suffix="PB"
