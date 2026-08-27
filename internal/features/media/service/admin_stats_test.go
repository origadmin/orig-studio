/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package service

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/origadmin/runtime/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/sqlite3ent/sqlite3"

	media "origadmin/application/origstudio/api/gen/v1/media"
	types "origadmin/application/origstudio/api/gen/v1/types"
	"origadmin/application/origstudio/internal/dal/entity"
	emedia "origadmin/application/origstudio/internal/dal/entity/media"
	"origadmin/application/origstudio/internal/dal/entity/order"
	"origadmin/application/origstudio/internal/features/system/dal"
)

// seedStatsDB opens an in-memory SQLite client, creates the full schema, and
// seeds deterministic data so the admin stats handlers return real (non-zero) values.
func seedStatsDB(t *testing.T) *entity.Client {
	t.Helper()
	client, err := entity.Open("sqlite3", "file:adminstats?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Schema.Create(context.Background()))

	ctx := context.Background()
	now := time.Now()

	u := client.User.Create().
		SetUsername("u1").SetEmail("u1@example.com").SetPassword("p").SetName("User One").
		SaveX(ctx)

	client.Media.Create().SetUser(u).SetTitle("m1").SetURL("https://example.com/m1.mp4").
		SetType("video").SetPrivacy(emedia.PrivacyPUBLIC).
		SetViewCount(100).SetCreateTime(now.AddDate(0, 0, -2)).SaveX(ctx)
	client.Media.Create().SetUser(u).SetTitle("m2").SetURL("https://example.com/m2.jpg").
		SetType("image").SetPrivacy(emedia.PrivacyPRIVATE).
		SetViewCount(50).SetCreateTime(now.AddDate(0, 0, -1)).SaveX(ctx)

	plan := client.SubscriptionPlan.Create().
		SetName("pro").SetPrice(9.99).SetDurationDays(30).SaveX(ctx)

	// o1: paid + linked to a plan => counts as subscription revenue
	client.Order.Create().SetOrderNo("O1").SetAmount(9.99).
		SetUser(u).SetStatus(order.StatusPaid).SetPlanID(plan.ID).
		SetCreateTime(now.AddDate(0, 0, -2)).SaveX(ctx)
	// o2: completed, no plan => counts as one-time revenue only
	client.Order.Create().SetOrderNo("O2").SetAmount(19.99).
		SetUser(u).SetStatus(order.StatusCompleted).
		SetCreateTime(now.AddDate(0, 0, -1)).SaveX(ctx)
	// o3: pending => excluded from revenue entirely
	client.Order.Create().SetOrderNo("O3").SetAmount(5.00).
		SetUser(u).SetStatus(order.StatusPending).
		SetCreateTime(now).SaveX(ctx)

	return client
}

func findStatPoint(pts []*types.StatPoint, day time.Time) int64 {
	key := day.Format("2006-01-02")
	for _, p := range pts {
		if p.Date == key {
			return p.Value
		}
	}
	return 0
}

// TestAdminServiceStats verifies the four admin stats gRPC handlers return
// real aggregated data instead of empty stubs.
func TestAdminServiceStats(t *testing.T) {
	client := seedStatsDB(t)
	repo := dal.NewStatsRepo(client)
	svc := NewAdminService(repo, nil, log.NewStdLogger(io.Discard))
	ctx := context.Background()

	t.Run("GetMediaStats", func(t *testing.T) {
		resp, err := svc.GetMediaStats(ctx, &media.GetMediaStatsRequest{Period: "month"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, int32(2), resp.TotalUploads)
		assert.Equal(t, int32(150), resp.TotalViews)
		assert.Len(t, resp.DailyStats, 30)
	})

	t.Run("GetUserStats", func(t *testing.T) {
		resp, err := svc.GetUserStats(ctx, &media.GetUserStatsRequest{Period: "month"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, int32(1), resp.TotalUsers)
		assert.Len(t, resp.DailyStats, 30)
	})

	t.Run("GetTrafficStats", func(t *testing.T) {
		resp, err := svc.GetTrafficStats(ctx, &media.GetTrafficStatsRequest{Period: "month"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		// No bandwidth/request log exists in schema; documented gap => 0.
		assert.Equal(t, int64(0), resp.TotalBandwidth)
		assert.Equal(t, int64(0), resp.TotalRequests)
		assert.Len(t, resp.DailyStats, 30)
		// Daily traffic is approximated by daily view counts.
		assert.Equal(t, int64(100), findStatPoint(resp.DailyStats, time.Now().AddDate(0, 0, -2)))
		assert.Equal(t, int64(50), findStatPoint(resp.DailyStats, time.Now().AddDate(0, 0, -1)))
	})

	t.Run("GetRevenueStats", func(t *testing.T) {
		resp, err := svc.GetRevenueStats(ctx, &media.GetRevenueStatsRequest{Period: "month"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		// Revenue is reported in minor units (cents).
		assert.Equal(t, int64(2998), resp.TotalRevenue)       // (9.99 + 19.99) * 100
		assert.Equal(t, int64(999), resp.SubscriptionRevenue)  // 9.99 * 100 (plan order)
		assert.Equal(t, int64(0), resp.AdRevenue)              // no ad source (documented gap)
		assert.Len(t, resp.DailyStats, 30)
		assert.Equal(t, int64(999), findStatPoint(resp.DailyStats, time.Now().AddDate(0, 0, -2)))
		assert.Equal(t, int64(1999), findStatPoint(resp.DailyStats, time.Now().AddDate(0, 0, -1)))
	})
}
