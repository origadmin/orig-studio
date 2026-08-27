/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 *
 * BUG-143 root cause ① regression: `listable` is a DERIVED visibility flag
 * (encoding_status=success AND review_status=reviewed AND state=active).
 * It must NEVER be taken from the client — admin and portal update paths must
 * recompute it server-side, otherwise any edit (even just adding a tag)
 * silently demotes reviewed content to invisible on every portal list.
 */

package service

import (
	"testing"

	types "origadmin/application/origstudio/api/gen/v1/types"
)

func TestPortalOwnerEditableFieldsDoesNotOverwriteListable(t *testing.T) {
	// Simulate a media that was properly reviewed/encoded/active (listable=true)
	// being edited by its owner. The incoming (client) payload carries
	// Listable=false (proto3 default when the UI does not submit it).
	existing := &types.Media{
		Title:          "original title",
		Tags:           []string{"golang"},
		State:          "active",
		Listable:       true,
		EncodingStatus: "success",
		ReviewStatus:   "reviewed",
	}
	src := &types.Media{
		Title:    "new title",
		Tags:     []string{"golang", "video"},
		Listable: false, // client default — must be ignored
	}

	// nil paths => legacy full-merge branch, which is exactly what a client
	// that submits no update_mask hits.
	portalOwnerEditableFields(existing, src, nil)

	if existing.Title != "new title" {
		t.Fatalf("title was not applied: got %q", existing.Title)
	}
	if len(existing.Tags) != 2 {
		t.Fatalf("tags were not applied: got %v", existing.Tags)
	}
	// Critical: the merge whitelist must NOT include Listable. Since the
	// derived flag is recomputed in UpdateMedia via ShouldBeListable, the merge
	// itself must leave the existing value untouched.
	if existing.Listable != true {
		t.Fatalf("portalOwnerEditableFields overwrote listable: got %v, want true (must stay derived server-side)", existing.Listable)
	}
}

func TestPortalOwnerEditableFieldsWhitelist(t *testing.T) {
	// Ensure the whitelist never grows to include owner/system-only fields.
	existing := &types.Media{
		UserId: "user-1",
	}
	src := &types.Media{
		UserId: "attacker",
	}

	portalOwnerEditableFields(existing, src, nil)

	if existing.UserId != "user-1" {
		t.Fatalf("portal owner edited UserId: got %q, want %q", existing.UserId, "user-1")
	}
}
