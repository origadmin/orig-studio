// Copyright (c) 2024 OrigAdmin. All rights reserved.

//go:build ent
// +build ent

package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"origadmin/application/origcms/internal/data/entity/schema"
)

func main() {
	ex, err := schemaExample()
	if err != nil {
		log.Fatal(err)
	}
	genGraph := gen.NewGraph(gen.NewConfig("ent", schema.Tables), schema.Nodes, schema.Edges, ex)
	if err := entc.Generate(".", &gen.Config{}, entc.WithGraph(genGraph, schema.Graph())); err != nil {
		log.Fatal("running entc generator:", err)
	}
}

func schemaExample() (map[string]any, error) {
	return make(map[string]any), nil
}
