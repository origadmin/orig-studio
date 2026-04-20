/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package convpb implements the functions, types, and interfaces for the module.
package convpb

//go:generate abgen .

// Abgen can convert structure to structure, the name must be consistent,
// but case and type can be ignored,
// and if it encounters something that cannot be automatically converted,
// a conversion function will be generated
//go:abgen:package:path=origadmin/application/origcms/internal/data/entity,alias=entity
//go:abgen:package:path=origadmin/application/origcms/api/gen/v1/types,alias=types
//go:abgen:pair:packages="entity,types"
//go:abgen:convert:direction="both"
//go:abgen:convert:source:suffix=""
//go:abgen:convert:target:suffix="PB"
