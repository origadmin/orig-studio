/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package mixin provides common ent schema mixins.
package mixin

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"entgo.io/ent/dialect"
)

// AuditMixin provides audit fields (create_time, update_time, create_author, update_author).
type AuditMixin struct {
	mixin.Schema
}

func (AuditMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("create_time").
			GoType(time.Time{}).
			Default(time.Now).
			Immutable().
			Comment("Create time"),
		field.Time("update_time").
			GoType(time.Time{}).
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Update time"),
		field.Int64("create_author").
			Default(0).
			Comment("Create author user ID"),
		field.Int64("update_author").
			Default(0).
			Comment("Update author user ID"),
	}
}

// SoftDeleteMixin provides soft delete fields (deleted_at).
type SoftDeleteMixin struct {
	mixin.Schema
}

func (SoftDeleteMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("deleted_at").
			GoType(time.Time{}).
			Optional().
			Nillable().
			Comment("Delete time"),
	}
}

// FieldUUID generates a UUID field with the given name and comment.
func FieldUUID(name, comment string) ent.Field {
	return field.String(name).
		MaxLen(36).
		Default("").
		Comment(comment)
}

// FieldUUIDFK generates a UUID field with the given name and comment.
func FieldUUIDFK(name, comment string) ent.Field {
	return field.String(name).
		MaxLen(36).
		Default("").
		Comment(comment)
}

// Time generates a time field with the given name and comment.
func Time(name, comment string) ent.Field {
	return field.Time(name).
		GoType(time.Time{}).
		Default(time.Now).
		SchemaType(map[string]string{
			dialect.SQLite:   "datetime",
			dialect.Postgres: "timestamptz",
		}).
		Comment(comment)
}

// TimeOptional generates an optional time field with the given name and comment.
func TimeOptional(name, comment string) ent.Field {
	return field.Time(name).
		GoType(time.Time{}).
		Optional().
		Nillable().
		SchemaType(map[string]string{
			dialect.SQLite:   "datetime",
			dialect.Postgres: "timestamptz",
		}).
		Comment(comment)
}
