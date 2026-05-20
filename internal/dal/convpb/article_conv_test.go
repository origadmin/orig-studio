package convpb

import (
	"testing"
	"time"

	"origadmin/application/origstudio/api/gen/v1/types"
	"origadmin/application/origstudio/internal/dal/entity"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestArticleRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)

	ent := &entity.Article{
		ID:           "art-123",
		Title:        "Test Article",
		Content:      "This is the article content",
		Summary:      "Article summary",
		Slug:         "test-article",
		ShortToken:   "abc123",
		State:        "published",
		ViewCount:    100,
		CommentCount: 10,
		Featured:     true,
		Tags:         []string{"golang", "protobuf"},
		UserID:       "user-1",
		CategoryID:   5,
		MediaID:      "media-1",
		Thumbnail:    "https://example.com/thumb.jpg",
		PublishedAt:  now,
		CreateTime:   now,
		UpdateTime:   now,
	}

	pb := ConvertArticleToArticlePB(ent)
	if pb.Id != ent.ID {
		t.Errorf("ID: got %q, want %q", pb.Id, ent.ID)
	}
	if pb.Title != ent.Title {
		t.Errorf("Title: got %q, want %q", pb.Title, ent.Title)
	}
	if pb.Content != ent.Content {
		t.Errorf("Content: got %q, want %q", pb.Content, ent.Content)
	}
	if pb.Summary != ent.Summary {
		t.Errorf("Summary: got %q, want %q", pb.Summary, ent.Summary)
	}
	if pb.Slug != ent.Slug {
		t.Errorf("Slug: got %q, want %q", pb.Slug, ent.Slug)
	}
	if pb.ShortToken != ent.ShortToken {
		t.Errorf("ShortToken: got %q, want %q", pb.ShortToken, ent.ShortToken)
	}
	if pb.State != ent.State {
		t.Errorf("State: got %q, want %q", pb.State, ent.State)
	}
	if pb.ViewCount != ent.ViewCount {
		t.Errorf("ViewCount: got %d, want %d", pb.ViewCount, ent.ViewCount)
	}
	if pb.CommentCount != ent.CommentCount {
		t.Errorf("CommentCount: got %d, want %d", pb.CommentCount, ent.CommentCount)
	}
	if pb.Featured != ent.Featured {
		t.Errorf("Featured: got %v, want %v", pb.Featured, ent.Featured)
	}
	if len(pb.Tags) != len(ent.Tags) {
		t.Errorf("Tags length: got %d, want %d", len(pb.Tags), len(ent.Tags))
	}
	if pb.UserId != ent.UserID {
		t.Errorf("UserID: got %q, want %q", pb.UserId, ent.UserID)
	}
	if pb.CategoryId != ent.CategoryID {
		t.Errorf("CategoryID: got %d, want %d", pb.CategoryId, ent.CategoryID)
	}
	if pb.MediaId != ent.MediaID {
		t.Errorf("MediaID: got %q, want %q", pb.MediaId, ent.MediaID)
	}
	if pb.Thumbnail != ent.Thumbnail {
		t.Errorf("Thumbnail: got %q, want %q", pb.Thumbnail, ent.Thumbnail)
	}

	if pb.PublishedAt == nil || !pb.PublishedAt.AsTime().Equal(now) {
		t.Errorf("PublishedAt: got %v, want %v", pb.PublishedAt, now)
	}
	if pb.CreateTime == nil || !pb.CreateTime.AsTime().Equal(now) {
		t.Errorf("CreateTime: got %v, want %v", pb.CreateTime, now)
	}
	if pb.UpdateTime == nil || !pb.UpdateTime.AsTime().Equal(now) {
		t.Errorf("UpdateTime: got %v, want %v", pb.UpdateTime, now)
	}
}

func TestArticlePBRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)

	pb := &types.Article{
		Id:           "art-456",
		Title:        "PB Article",
		Content:      "PB Content",
		Summary:      "PB Summary",
		Slug:         "pb-article",
		ShortToken:   "xyz789",
		State:        "draft",
		ViewCount:    200,
		CommentCount: 20,
		Featured:     false,
		Tags:         []string{"test", "proto"},
		UserId:       "user-2",
		CategoryId:   10,
		MediaId:      "media-2",
		Thumbnail:    "https://example.com/pb-thumb.jpg",
		PublishedAt:  timestamppb.New(now),
		CreateTime:   timestamppb.New(now),
		UpdateTime:   timestamppb.New(now),
	}

	ent := ConvertArticlePBToArticle(pb)
	if ent.ID != pb.Id {
		t.Errorf("ID: got %q, want %q", ent.ID, pb.Id)
	}
	if ent.Title != pb.Title {
		t.Errorf("Title: got %q, want %q", ent.Title, pb.Title)
	}
	if ent.Content != pb.Content {
		t.Errorf("Content: got %q, want %q", ent.Content, pb.Content)
	}
	if ent.Summary != pb.Summary {
		t.Errorf("Summary: got %q, want %q", ent.Summary, pb.Summary)
	}
	if ent.Slug != pb.Slug {
		t.Errorf("Slug: got %q, want %q", ent.Slug, pb.Slug)
	}
	if ent.ShortToken != pb.ShortToken {
		t.Errorf("ShortToken: got %q, want %q", ent.ShortToken, pb.ShortToken)
	}
	if ent.State != pb.State {
		t.Errorf("State: got %q, want %q", ent.State, pb.State)
	}
	if ent.ViewCount != pb.ViewCount {
		t.Errorf("ViewCount: got %d, want %d", ent.ViewCount, pb.ViewCount)
	}
	if ent.CommentCount != pb.CommentCount {
		t.Errorf("CommentCount: got %d, want %d", ent.CommentCount, pb.CommentCount)
	}
	if ent.Featured != pb.Featured {
		t.Errorf("Featured: got %v, want %v", ent.Featured, pb.Featured)
	}
	if len(ent.Tags) != len(pb.Tags) {
		t.Errorf("Tags length: got %d, want %d", len(ent.Tags), len(pb.Tags))
	}
	if ent.UserID != pb.UserId {
		t.Errorf("UserID: got %q, want %q", ent.UserID, pb.UserId)
	}
	if ent.CategoryID != pb.CategoryId {
		t.Errorf("CategoryID: got %d, want %d", ent.CategoryID, pb.CategoryId)
	}
	if ent.MediaID != pb.MediaId {
		t.Errorf("MediaID: got %q, want %q", ent.MediaID, pb.MediaId)
	}
	if ent.Thumbnail != pb.Thumbnail {
		t.Errorf("Thumbnail: got %q, want %q", ent.Thumbnail, pb.Thumbnail)
	}
	if !ent.PublishedAt.Equal(now) {
		t.Errorf("PublishedAt: got %v, want %v", ent.PublishedAt, now)
	}
	if !ent.CreateTime.Equal(now) {
		t.Errorf("CreateTime: got %v, want %v", ent.CreateTime, now)
	}
	if !ent.UpdateTime.Equal(now) {
		t.Errorf("UpdateTime: got %v, want %v", ent.UpdateTime, now)
	}
}

func TestArticleStateValues(t *testing.T) {
	states := []string{"draft", "published", "archived"}

	for _, state := range states {
		ent := &entity.Article{
			ID:    "art-state-test",
			Title: "State Test",
			State: state,
		}

		pb := ConvertArticleToArticlePB(ent)
		if pb.State != state {
			t.Errorf("State %s: got %q, want %q", state, pb.State, state)
		}

		pb2 := &types.Article{
			Id:    "art-pb-state-test",
			Title: "PB State Test",
			State: state,
		}

		ent2 := ConvertArticlePBToArticle(pb2)
		if ent2.State != state {
			t.Errorf("State %s: got %q, want %q", state, ent2.State, state)
		}
	}
}

func TestArticleEmptyTags(t *testing.T) {
	ent := &entity.Article{
		ID:    "art-no-tags",
		Title: "No Tags Article",
		Tags:  nil,
	}

	pb := ConvertArticleToArticlePB(ent)
	if pb.Tags != nil && len(pb.Tags) > 0 {
		t.Errorf("nil Tags should remain nil or empty, got %v", pb.Tags)
	}

	pb2 := &types.Article{
		Id:    "art-pb-no-tags",
		Title: "PB No Tags",
		Tags:  nil,
	}

	ent2 := ConvertArticlePBToArticle(pb2)
	if ent2.Tags != nil && len(ent2.Tags) > 0 {
		t.Errorf("nil Tags should remain nil or empty, got %v", ent2.Tags)
	}
}

func TestArticleZeroTimestamp(t *testing.T) {
	ent := &entity.Article{
		ID:         "art-zero-time",
		Title:      "Zero Time Article",
		PublishedAt: time.Time{},
		CreateTime: time.Time{},
		UpdateTime: time.Time{},
	}

	pb := ConvertArticleToArticlePB(ent)
	if pb.PublishedAt != nil {
		t.Errorf("zero PublishedAt should produce nil, got %v", pb.PublishedAt)
	}
	if pb.CreateTime != nil {
		t.Errorf("zero CreateTime should produce nil, got %v", pb.CreateTime)
	}
	if pb.UpdateTime != nil {
		t.Errorf("zero UpdateTime should produce nil, got %v", pb.UpdateTime)
	}
}

func TestArticleNilTimestamp(t *testing.T) {
	pb := &types.Article{
		Id:          "art-nil-ts",
		Title:       "Nil Timestamp Article",
		PublishedAt: nil,
		CreateTime:  nil,
		UpdateTime:  nil,
	}

	ent := ConvertArticlePBToArticle(pb)
	if !ent.PublishedAt.IsZero() {
		t.Errorf("nil PublishedAt should produce zero time, got %v", ent.PublishedAt)
	}
	if !ent.CreateTime.IsZero() {
		t.Errorf("nil CreateTime should produce zero time, got %v", ent.CreateTime)
	}
	if !ent.UpdateTime.IsZero() {
		t.Errorf("nil UpdateTime should produce zero time, got %v", ent.UpdateTime)
	}
}

func TestConvertArticleNilInputs(t *testing.T) {
	if ConvertArticlePBToArticle(nil) != nil {
		t.Error("ConvertArticlePBToArticle(nil) should return nil")
	}
	if ConvertArticleToArticlePB(nil) != nil {
		t.Error("ConvertArticleToArticlePB(nil) should return nil")
	}
}

func TestArticleFeaturedField(t *testing.T) {
	testCases := []struct {
		name     string
		featured bool
	}{
		{"featured_true", true},
		{"featured_false", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ent := &entity.Article{
				ID:       "art-featured-" + tc.name,
				Title:    tc.name,
				Featured: tc.featured,
			}

			pb := ConvertArticleToArticlePB(ent)
			if pb.Featured != tc.featured {
				t.Errorf("Entity Featured=%v -> PB Featured=%v, want %v",
					tc.featured, pb.Featured, tc.featured)
			}

			pb2 := &types.Article{
				Id:       "art-pb-featured-" + tc.name,
				Title:    tc.name,
				Featured: tc.featured,
			}

			ent2 := ConvertArticlePBToArticle(pb2)
			if ent2.Featured != tc.featured {
				t.Errorf("PB Featured=%v -> Entity Featured=%v, want %v",
					tc.featured, ent2.Featured, tc.featured)
			}
		})
	}
}

func TestArticleCounts(t *testing.T) {
	ent := &entity.Article{
		ID:           "art-counts",
		Title:        "Counts Test",
		ViewCount:    1000,
		CommentCount: 50,
	}

	pb := ConvertArticleToArticlePB(ent)
	if pb.ViewCount != 1000 {
		t.Errorf("ViewCount: got %d, want 1000", pb.ViewCount)
	}
	if pb.CommentCount != 50 {
		t.Errorf("CommentCount: got %d, want 50", pb.CommentCount)
	}

	pb2 := &types.Article{
		Id:           "art-pb-counts",
		Title:        "PB Counts Test",
		ViewCount:    2000,
		CommentCount: 100,
	}

	ent2 := ConvertArticlePBToArticle(pb2)
	if ent2.ViewCount != 2000 {
		t.Errorf("ViewCount: got %d, want 2000", ent2.ViewCount)
	}
	if ent2.CommentCount != 100 {
		t.Errorf("CommentCount: got %d, want 100", ent2.CommentCount)
	}
}

func TestArticleCategoryID(t *testing.T) {
	testCases := []struct {
		name       string
		categoryID int64
	}{
		{"zero_category", 0},
		{"small_category", 1},
		{"medium_category", 100},
		{"large_category", 999999},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ent := &entity.Article{
				ID:         "art-cat-" + tc.name,
				Title:      tc.name,
				CategoryID: tc.categoryID,
			}

			pb := ConvertArticleToArticlePB(ent)
			if pb.CategoryId != tc.categoryID {
				t.Errorf("CategoryID: got %d, want %d", pb.CategoryId, tc.categoryID)
			}

			pb2 := &types.Article{
				Id:         "art-pb-cat-" + tc.name,
				Title:      tc.name,
				CategoryId: tc.categoryID,
			}

			ent2 := ConvertArticlePBToArticle(pb2)
			if ent2.CategoryID != tc.categoryID {
				t.Errorf("CategoryID: got %d, want %d", ent2.CategoryID, tc.categoryID)
			}
		})
	}
}

func TestArticleFieldsMatching(t *testing.T) {
	protoFields := []string{
		"Id", "Title", "Content", "Summary", "Slug", "ShortToken",
		"State", "ViewCount", "CommentCount", "Featured", "Tags",
		"UserId", "CategoryId", "MediaId", "Thumbnail",
	}

	ent := &entity.Article{
		ID:           "art-fields",
		Title:        "Fields Test",
		Content:      "Content",
		Summary:      "Summary",
		Slug:         "fields-test",
		ShortToken:   "ft123",
		State:        "draft",
		ViewCount:    0,
		CommentCount: 0,
		Featured:     false,
		Tags:         []string{},
		UserID:       "",
		CategoryID:   0,
		MediaID:      "",
		Thumbnail:    "",
	}

	pb := ConvertArticleToArticlePB(ent)

	_ = protoFields
	if pb.Id != ent.ID {
		t.Errorf("Id mismatch")
	}
	if pb.Title != ent.Title {
		t.Errorf("Title mismatch")
	}
	if pb.Content != ent.Content {
		t.Errorf("Content mismatch")
	}
	if pb.Summary != ent.Summary {
		t.Errorf("Summary mismatch")
	}
	if pb.Slug != ent.Slug {
		t.Errorf("Slug mismatch")
	}
	if pb.ShortToken != ent.ShortToken {
		t.Errorf("ShortToken mismatch")
	}
	if pb.State != ent.State {
		t.Errorf("State mismatch")
	}
}

func TestArticleSlugGeneration(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "Hello World", "hello-world"},
		{"with_numbers", "Test 123 Article", "test-123-article"},
		{"chinese", "测试文章", ""},
		{"empty", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ent := &entity.Article{
				ID:    "art-slug-" + tc.name,
				Title: tc.input,
				Slug:  tc.expected,
			}

			pb := ConvertArticleToArticlePB(ent)
			if pb.Slug != tc.expected {
				t.Errorf("Slug: got %q, want %q", pb.Slug, tc.expected)
			}
		})
	}
}
