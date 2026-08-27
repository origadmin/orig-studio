/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// reservedAdminDatatypes is the single source of truth for the reserved
// datatype keywords declared in
// docs/modules/shared/engineering/dashboard-stats-endpoint-convention.md
// (§命名空间保留) and enforced by MODULE_CLASSIFICATION_v3.md 原则 5.
//
// A reserved word MAY ONLY appear as the first segment of
//   /api/v1/admin/{reserved}/{domain-name}      (datatype route)
// and MUST NEVER appear as
//   /api/v1/admin/{reserved}/{id}               (CRUD collection)
// nor as a bare
//   /api/v1/admin/{reserved}                    (collection list)
//
// Rationale (BUG-216 / BUG-224): the datatype segment and the domain/collection
// segment share the same URL level under /admin/. If a future business domain is
// named "stats" and given CRUD (/admin/stats/{id}), it collides with the
// existing datatype route /admin/stats/medias — same segment, two meanings.
//
// KEEP IN SYNC: when adding a reserved keyword, update this map, the ADR
// reserved-word list, and MODULE_CLASSIFICATION_v3.md 原则 5.
var reservedAdminDatatypes = map[string]bool{
	"stats":  true,
	"metric": true,
	"audit":  true,
	"report": true,
}

// adminRouteRegex captures the quoted path inside google.api.http annotations,
// e.g. option (google.api.http) = { get: "/api/v1/admin/stats/medias" };
var adminRouteRegex = regexp.MustCompile(`"((?:/api/v1)?/admin/[^"]+)"`)

// idPlaceholder matches a gRPC-Gateway path parameter, e.g. {id}, {user_id}, {review.id}.
var idPlaceholder = regexp.MustCompile(`^\{.*\}$`)

// findProtoRoot walks up from this test file until it finds api/proto/v1.
func findProtoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")
	dir := filepath.Dir(file)
	for i := 0; i < 12; i++ {
		candidate := filepath.Join(dir, "api", "proto", "v1")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("could not locate api/proto/v1 from %s", file)
	return ""
}

// collectAdminRoutes scans every .proto under the proto tree for admin HTTP
// route annotations and returns the unique set of path strings.
func collectAdminRoutes(t *testing.T) []string {
	t.Helper()
	root := findProtoRoot(t)
	var routes []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".proto") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for _, m := range adminRouteRegex.FindAllStringSubmatch(string(data), -1) {
			routes = append(routes, m[1])
		}
		return nil
	})
	require.NoError(t, err)
	require.NotEmpty(t, routes, "no /admin/* routes found; proto tree may have moved")
	return routes
}

// TestReservedDatatypesNotUsedAsCollections enforces the mutual-exclusion rule:
// a reserved datatype keyword must never be reused as a CRUD collection segment
// (/{id}) or as a bare collection list under /admin/.
//
// This is the automated half of the governance in
// dashboard-stats-endpoint-convention.md §命名空间保留. The reverse direction
// (forgetting to reserve a NEW datatype) is caught by human review at the entry
// indexes, because /admin/{collection}/{subresource} and /admin/{datatype}/{domain}
// are indistinguishable by shape alone.
func TestReservedDatatypesNotUsedAsCollections(t *testing.T) {
	routes := collectAdminRoutes(t)
	for _, route := range routes {
		segs := strings.Split(strings.Trim(route, "/"), "/")
		// segs = [api, v1, admin, seg1, seg2?, ...]
		if len(segs) < 4 || segs[2] != "admin" {
			continue
		}
		seg1 := segs[3]

		if !reservedAdminDatatypes[seg1] {
			// seg1 is a plain domain/collection name — by construction it is
			// not reserved. The reverse collision (a reserved word used here)
			// is impossible to reach in this branch, so nothing to assert.
			continue
		}

		// seg1 IS reserved: it must be a datatype route, i.e. it must have a
		// second segment that is NOT an id placeholder.
		if len(segs) < 5 {
			t.Errorf("reserved datatype %q used as a bare collection (no domain segment): %s", seg1, route)
			continue
		}
		seg2 := segs[4]
		if idPlaceholder.MatchString(seg2) {
			t.Errorf("reserved datatype %q reused as a CRUD collection (/admin/%s/{id}): %s", seg1, seg1, route)
		}
	}
}

// TestReservedDatatypesLiveKeywordExercised is a positive sanity check that the
// convention is actually in use and the reserved set is not pure fiction: the
// live datatype keyword ("stats") must be exercised by at least one
// /admin/stats/{domain} route. Future-reserved keywords (metric/audit/report)
// are allowed to have no route yet — they are placeholders, not violations.
func TestReservedDatatypesLiveKeywordExercised(t *testing.T) {
	routes := collectAdminRoutes(t)
	liveKeyword := "stats"
	exercised := false
	for _, route := range routes {
		segs := strings.Split(strings.Trim(route, "/"), "/")
		if len(segs) >= 5 && segs[2] == "admin" && segs[3] == liveKeyword && !idPlaceholder.MatchString(segs[4]) {
			exercised = true
			break
		}
	}
	require.True(t, exercised, "live reserved datatype %q has no /admin/%s/{domain} route; the convention is broken or the reserved set drifted", liveKeyword, liveKeyword)
}
