/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package entity is the data access object for SYS.
package entity

//go:generate ent generate --feature intercept --feature schema/snapshot --feature sql/versioned-migration --feature sql/lock --feature sql/modifier ./schema
