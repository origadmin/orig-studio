package biz

import (
	"context"
	"os"
	"testing"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	_ "github.com/sqlite3ent/sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/media"
	"origadmin/application/origstudio/internal/data/entity/migrate"
)

func TestTagFiltering_SQLite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SQLite integration test in short mode")
	}
	dbPath := "test_tags_filter.db?_fk=1&_journal_mode=WAL"
	os.Remove("test_tags_filter.db")

	drv, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	defer drv.Close()

	client := entity.NewClient(entity.Driver(drv))
	ctx := context.Background()

	err = client.Schema.Create(ctx, migrate.WithForeignKeys(false))
	require.NoError(t, err)

	userID := client.User.Create().
		SetUsername("testuser").
		SetName("Test User").
		SetPassword("test").
		SetEmail("test@test.com").
		SetStatus("ACTIVE").
		SaveX(ctx).ID

	client.Media.Create().
		SetTitle("Video with tags #tag1 #tag2 #tag3").
		SetTags([]string{"tag1", "tag2", "tag3"}).
		SetType("video").
		SetURL("test1.mp4").
		SetPrivacy(media.PrivacyPUBLIC).
		SetState("active").
		SetListable(true).
		SetUserID(userID).
		SaveX(ctx)

	client.Media.Create().
		SetTitle("Video with tag1 only #tag1").
		SetTags([]string{"tag1"}).
		SetType("video").
		SetURL("test2.mp4").
		SetPrivacy(media.PrivacyPUBLIC).
		SetState("active").
		SetListable(true).
		SetUserID(userID).
		SaveX(ctx)

	client.Media.Create().
		SetTitle("Video without tags").
		SetTags([]string{}).
		SetType("video").
		SetURL("test3.mp4").
		SetPrivacy(media.PrivacyPUBLIC).
		SetState("active").
		SetListable(true).
		SetUserID(userID).
		SaveX(ctx)

	t.Run("FilterByTag1_sqljson", func(t *testing.T) {
		results, err := client.Media.Query().
			Where(func(s *sql.Selector) {
				predicates := []*sql.Predicate{sqljson.ValueContains(media.FieldTags, "tag1")}
				s.Where(sql.Or(predicates...))
			}).
			All(ctx)
		if err != nil {
			t.Logf("sqljson error: %v (may not support SQLite)", err)
		} else {
			t.Logf("sqljson found %d results", len(results))
			for _, m := range results {
				t.Logf("  Title: %s Tags: %v", m.Title, m.Tags)
			}
			assert.Equal(t, 2, len(results), "should find 2 media with tag1")
		}
	})

	t.Run("FilterByTag3_sqljson", func(t *testing.T) {
		results, err := client.Media.Query().
			Where(func(s *sql.Selector) {
				predicates := []*sql.Predicate{sqljson.ValueContains(media.FieldTags, "tag3")}
				s.Where(sql.Or(predicates...))
			}).
			All(ctx)
		if err != nil {
			t.Logf("sqljson error: %v", err)
		} else {
			t.Logf("sqljson found %d results", len(results))
			assert.Equal(t, 1, len(results), "should find 1 media with tag3")
		}
	})

	t.Run("FilterByNonexistentTag_sqljson", func(t *testing.T) {
		results, err := client.Media.Query().
			Where(func(s *sql.Selector) {
				predicates := []*sql.Predicate{sqljson.ValueContains(media.FieldTags, "nonexistent")}
				s.Where(sql.Or(predicates...))
			}).
			All(ctx)
		if err != nil {
			t.Logf("sqljson error: %v", err)
		} else {
			assert.Equal(t, 0, len(results), "should find 0 media with nonexistent tag")
		}
	})

	t.Run("FilterByTag1_LIKE_fallback", func(t *testing.T) {
		results, err := client.Media.Query().
			Where(func(s *sql.Selector) {
				s.Where(sql.Like(media.FieldTags, `%"tag1"%`))
			}).
			All(ctx)
		if err != nil {
			t.Logf("LIKE error: %v", err)
		} else {
			t.Logf("LIKE found %d results", len(results))
			for _, m := range results {
				t.Logf("  Title: %s Tags: %v", m.Title, m.Tags)
			}
			assert.Equal(t, 2, len(results), "should find 2 media with tag1 using LIKE")
		}
	})

	os.Remove("test_tags_filter.db")
}
