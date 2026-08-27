/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 *
 * BUG-131 regression: every filter declared on ListMediasRequest must reach
 * MediaQueryOption. The original defect was not a bad filter -- it was a
 * *missing mapping*: the live gRPC handler hand-rolled its own request->option
 * conversion and quietly skipped tags/category_ids/state/featured, so the DAL
 * (which implements all of them) never saw the values and search returned the
 * full library no matter what the user clicked.
 */

package dto

import (
	"testing"

	media "origadmin/application/origstudio/api/gen/v1/media"
)

func TestListMediasRequestToQueryOptionMapsEveryFilter(t *testing.T) {
	state := "active"
	featured := true
	mtype := int32(1)
	status := int32(2)
	userID := "user-1"
	categoryID := int64(4)

	req := &media.ListMediasRequest{
		Page:         2,
		PageSize:     30,
		Keyword:      "golang",
		Type:         &mtype,
		Status:       &status,
		UserId:       &userID,
		CategoryId:   &categoryID,
		OrderBy:      "view_count",
		Descending:   true,
		Tags:         []string{"golang", "video"},
		CategoryIds:  []int64{4, 5},
		CreatedAfter: "2026-01-01T00:00:00Z",
		State:        &state,
		Featured:     &featured,
	}

	opts, err := ListMediasRequestToQueryOption(req)
	if err != nil {
		t.Fatalf("unexpected error mapping request: %v", err)
	}

	if opts.Page != 2 || opts.PageSize != 30 {
		t.Fatalf("pagination not mapped: page=%d size=%d", opts.Page, opts.PageSize)
	}
	if opts.Keyword != "golang" {
		t.Errorf("keyword not mapped: %q", opts.Keyword)
	}
	if opts.Type == nil || *opts.Type != mtype {
		t.Errorf("type not mapped: %v", opts.Type)
	}
	if opts.Status == nil || *opts.Status != status {
		t.Errorf("status not mapped: %v", opts.Status)
	}
	if opts.UserID == nil || *opts.UserID != userID {
		t.Errorf("user_id not mapped: %v", opts.UserID)
	}
	if opts.CategoryID == nil || *opts.CategoryID != categoryID {
		t.Errorf("category_id not mapped: %v", opts.CategoryID)
	}
	if opts.OrderBy != "view_count" || !opts.Descending {
		t.Errorf("ordering not mapped: order_by=%q desc=%v", opts.OrderBy, opts.Descending)
	}
	if len(opts.Tags) != 2 {
		t.Errorf("tags not mapped: %v", opts.Tags)
	}
	if len(opts.CategoryIDs) != 2 {
		t.Errorf("category_ids not mapped: %v", opts.CategoryIDs)
	}
	if opts.CreatedAfter != "2026-01-01T00:00:00Z" {
		t.Errorf("created_after not mapped: %q", opts.CreatedAfter)
	}
	if opts.State != state {
		t.Errorf("state not mapped: %q", opts.State)
	}
	if opts.Featured == nil || *opts.Featured != featured {
		t.Errorf("featured not mapped: %v", opts.Featured)
	}
}

// TestListMediasRequestFieldCountGuard fails whenever a field is added to the
// proto message. That is the point: BUG-131 was caused by a filter being
// declared but never mapped. When this guard trips, wire the new field into
// ListMediasRequestToQueryOption (and the DAL), extend the mapping test above,
// then bump the expected count here.
func TestListMediasRequestFieldCountGuard(t *testing.T) {
	const mappedFieldCount = 16 // page..seed, see media_service.proto

	got := (&media.ListMediasRequest{}).ProtoReflect().Descriptor().Fields().Len()
	if got != mappedFieldCount {
		t.Fatalf("ListMediasRequest now has %d fields, guard expects %d: "+
			"a new filter was declared -- map it in "+
			"ListMediasRequestToQueryOption before bumping this number",
			got, mappedFieldCount)
	}
}

// TestListMediasRequestToQueryOptionRandomSeed covers BUG-226: order_by=random
// requires a positive uint32 seed, which becomes opts.RandomSeed (a bound
// parameter for the ORDER BY expression in the DAL). The seed is never used as
// a raw string in SQL, so the same seed is reproducible and injection-safe.
func TestListMediasRequestToQueryOptionRandomSeed(t *testing.T) {
	seed := uint32(12345)

	// happy path: positive seed -> RandomSeed set
	opts, err := ListMediasRequestToQueryOption(&media.ListMediasRequest{
		OrderBy: "random",
		Seed:    &seed,
	})
	if err != nil {
		t.Fatalf("unexpected error for valid seed: %v", err)
	}
	if opts.RandomSeed == nil || *opts.RandomSeed != seed {
		t.Fatalf("RandomSeed not mapped: %v", opts.RandomSeed)
	}

	// missing seed -> rejected
	if _, err := ListMediasRequestToQueryOption(&media.ListMediasRequest{OrderBy: "random"}); err == nil {
		t.Error("expected error when order_by=random and seed is missing")
	}

	// zero seed -> rejected (seed must be positive)
	zero := uint32(0)
	if _, err := ListMediasRequestToQueryOption(&media.ListMediasRequest{OrderBy: "random", Seed: &zero}); err == nil {
		t.Error("expected error when order_by=random and seed is 0")
	}

	// seed ignored unless order_by=random
	opts, err = ListMediasRequestToQueryOption(&media.ListMediasRequest{OrderBy: "create_time", Seed: &seed})
	if err != nil {
		t.Fatalf("unexpected error when seed present but order_by != random: %v", err)
	}
	if opts.RandomSeed != nil {
		t.Error("RandomSeed should be nil when order_by != random")
	}
}
