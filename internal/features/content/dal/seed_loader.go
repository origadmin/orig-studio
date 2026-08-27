/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dal

import (
	_ "embed"
	"fmt"
	"os"
	"regexp"

	"github.com/goccy/go-yaml"
)

//go:embed seeddata/taxonomy.yaml
var defaultTaxonomyYAML []byte

// taxonomySpec is the top-level taxonomy.yaml document.
type taxonomySpec struct {
	Categories []*catSpecNode `yaml:"categories"`
	Tags       []*tagSpec     `yaml:"tags"`
}

// catSpecNode is a category spec with an optional children subtree. The tree is
// expanded at seed time (parent upserted first, then children with ParentID set).
type catSpecNode struct {
	Name        string         `yaml:"name"`
	Slug        string         `yaml:"slug"`
	Kind        string         `yaml:"kind"` // deprecated: form|genre axis dropped 2026-08-26; kept for backward-compat, no longer set by taxonomy.yaml
	Icon        string         `yaml:"icon"`
	Color       string         `yaml:"color"`
	Sequence    int            `yaml:"sequence"`
	Description string         `yaml:"description"`
	IsGlobal    *bool          `yaml:"is_global"` // nil => true (portal categories are shared/global)
	Children    []*catSpecNode `yaml:"children"`
}

// tagSpec describes a tag to seed. Upsert by slug (D4: count>0 skip is gone).
type tagSpec struct {
	Title       string `yaml:"title"`
	Slug        string `yaml:"slug"`
	Color       string `yaml:"color"`
	Description string `yaml:"description"`
}

// slugRe enforces the §3.7 stable-key rule: slugs must be ASCII
// (^[a-z0-9_-]+$). A non-ASCII slug fails at startup, not at runtime.
var slugRe = regexp.MustCompile(`^[a-z0-9_-]+$`)

// loadTaxonomy reads the runtime override resources/configs/content/taxonomy.yaml
// when present (Dockerfile.runtime copies resources/ to /app/resources/, cwd=/app),
// else falls back to the embedded default (§3.2).
func loadTaxonomy() ([]byte, error) {
	const runtimePath = "resources/configs/content/taxonomy.yaml"
	if b, err := os.ReadFile(runtimePath); err == nil {
		return b, nil
	}
	return defaultTaxonomyYAML, nil
}

// parseTaxonomy validates and parses taxonomy YAML into specs.
func parseTaxonomy(data []byte) (*taxonomySpec, error) {
	var spec taxonomySpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parse taxonomy.yaml: %w", err)
	}
	if len(spec.Categories) == 0 {
		return nil, fmt.Errorf("taxonomy.yaml: no categories defined")
	}
	if err := validateTaxonomySlugs(&spec); err != nil {
		return nil, err
	}
	return &spec, nil
}

func validateTaxonomySlugs(spec *taxonomySpec) error {
	var walk func(nodes []*catSpecNode) error
	walk = func(nodes []*catSpecNode) error {
		for _, n := range nodes {
			if !slugRe.MatchString(n.Slug) {
				return fmt.Errorf("taxonomy.yaml: category slug %q violates ^[a-z0-9_-]+$ (must be ASCII stable key, §3.7)", n.Slug)
			}
			if err := walk(n.Children); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(spec.Categories); err != nil {
		return err
	}
	for _, t := range spec.Tags {
		if !slugRe.MatchString(t.Slug) {
			return fmt.Errorf("taxonomy.yaml: tag slug %q violates ^[a-z0-9_-]+$ (must be ASCII stable key, §3.7)", t.Slug)
		}
	}
	return nil
}

// seedTaxonomy loads and parses taxonomy (runtime override else embedded default).
func seedTaxonomy() (*taxonomySpec, error) {
	data, err := loadTaxonomy()
	if err != nil {
		return nil, err
	}
	return parseTaxonomy(data)
}
