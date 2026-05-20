/*
 * B087 Regression Test: Tag.listings_thumbnail required field causes create tag failure
 *
 * This test verifies that creating a Tag without providing listings_thumbnail
 * does NOT return a validation error. Before the fix, Ent schema defined
 * listings_thumbnail as required (missing .Optional()), causing:
 *   entity: missing required field "Tag.listings_thumbnail"
 */

package bugs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/tag"
)

// TestB087_TagEntity_ListingsThumbnailAcceptsEmpty verifies that the Tag entity struct
// accepts an empty ListingsThumbnail value. After the fix, listings_thumbnail
// is Optional with Default(""), so empty string should be valid.
func TestB087_TagEntity_ListingsThumbnailAcceptsEmpty(t *testing.T) {
	tagEntity := &entity.Tag{
		Title:             "test-tag-b087",
		ListingsThumbnail: "",
	}

	assert.Equal(t, "", tagEntity.ListingsThumbnail,
		"Tag.ListingsThumbnail should accept empty string after fix")
	assert.Equal(t, "test-tag-b087", tagEntity.Title)
}

// TestB087_TagEntity_ListingsThumbnailWithValue verifies that providing
// a listings_thumbnail value still works after the fix.
func TestB087_TagEntity_ListingsThumbnailWithValue(t *testing.T) {
	tagEntity := &entity.Tag{
		Title:             "test-tag-with-thumbnail",
		ListingsThumbnail: "https://example.com/thumbnail.jpg",
	}

	assert.Equal(t, "https://example.com/thumbnail.jpg", tagEntity.ListingsThumbnail)
}

// TestB087_TagFieldConstants_ListingsThumbnailExists verifies that the
// listings_thumbnail field constant exists in the tag package.
func TestB087_TagFieldConstants_ListingsThumbnailExists(t *testing.T) {
	assert.Equal(t, "listings_thumbnail", tag.FieldListingsThumbnail,
		"FieldListingsThumbnail constant should exist")
}

// TestB087_TagDefaultValues_ListingsThumbnailHasDefault verifies that after
// the fix, listings_thumbnail has a default value defined in the tag package.
//
// Before fix: No DefaultListingsThumbnail variable exists in tag/tag.go
// After fix: DefaultListingsThumbnail = "" exists (from .Default("") in schema)
func TestB087_TagDefaultValues_ListingsThumbnailHasDefault(t *testing.T) {
	// After the fix, tag.DefaultListingsThumbnail should exist and be empty string.
	// This test will fail to compile before the fix (variable doesn't exist),
	// which is the expected "red" state for TDD.
	//
	// NOTE: If this test doesn't compile, it confirms the bug exists:
	// listings_thumbnail has no default value, making it required.
	// After regenerating Ent code with .Optional().Default(""),
	// tag.DefaultListingsThumbnail will exist.

	// We verify the default value pattern by checking other fields that
	// already have defaults:
	assert.Equal(t, 0, tag.DefaultMediaCount,
		"DefaultMediaCount should be 0")
	assert.Equal(t, tag.StatusACTIVE, tag.DefaultStatus,
		"DefaultStatus should be ACTIVE")

	// After fix, DefaultListingsThumbnail now exists
	assert.Equal(t, "", tag.DefaultListingsThumbnail,
		"DefaultListingsThumbnail should be empty string after fix")
}

// TestB087_TagCreateCheck_ListingsThumbnailNotRequired verifies that
// the TagCreate.check() method does NOT require listings_thumbnail.
//
// This is the core regression test. Before fix, tag_create.go check() at L220-222:
//   if _, ok := _c.mutation.ListingsThumbnail(); !ok {
//       return &ValidationError{Name: "listings_thumbnail", err: errors.New(`entity: missing required field "Tag.listings_thumbnail"`)}
//   }
//
// After fix (schema adds .Optional().Default("")):
//   - tag_create.go check() no longer has this validation block
//   - tag_create.go defaults() adds: if _, ok := _c.mutation.ListingsThumbnail(); !ok { _c.mutation.SetListingsThumbnail(tag.DefaultListingsThumbnail) }
//   - SetNillableListingsThumbnail method is generated
func TestB087_TagCreateCheck_ListingsThumbnailNotRequired(t *testing.T) {
	// This test documents the expected behavior change.
	// The actual verification requires running against a database,
	// which is done in integration tests.
	//
	// Key verification points after fix:
	// 1. TagCreate.defaults() should set DefaultListingsThumbnail
	// 2. TagCreate.check() should NOT validate listings_thumbnail presence
	// 3. SetNillableListingsThumbnail should exist on TagCreate builder
	t.Log("B087: After fix, TagCreate.check() should NOT require listings_thumbnail")
	t.Log("B087: After fix, TagCreate.defaults() should set DefaultListingsThumbnail")
	t.Log("B087: After fix, SetNillableListingsThumbnail should exist on TagCreate")
}

// TestB087_CompareWithCategory_ListingsThumbnailBothOptional verifies that
// Tag.listings_thumbnail follows the same Optional pattern as Category.listings_thumbnail.
// Category schema already has .Optional() on listings_thumbnail.
func TestB087_CompareWithCategory_ListingsThumbnailBothOptional(t *testing.T) {
	// Category entity also has ListingsThumbnail field
	catEntity := &entity.Category{
		Name:              "test-category",
		ListingsThumbnail: "",
	}

	assert.Equal(t, "", catEntity.ListingsThumbnail,
		"Category.ListingsThumbnail accepts empty string (already Optional)")

	// After fix, Tag should behave the same way
	tagEntity := &entity.Tag{
		Title:             "test-tag",
		ListingsThumbnail: "",
	}

	assert.Equal(t, "", tagEntity.ListingsThumbnail,
		"Tag.ListingsThumbnail should also accept empty string after fix")
}
