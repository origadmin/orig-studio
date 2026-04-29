package convpb

import (
	"testing"

	"origadmin/application/origcms/internal/data/entity"
)

func TestConvertMediaToMediaPB_NilInput(t *testing.T) {
	result := ConvertMediaToMediaPB(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestConvertMediaToMediaPB_BasicFieldsOnly(t *testing.T) {
	from := &entity.Media{
		ID:       "media-1",
		Title:    "Test Video",
		Type:     "video",
		URL:      "https://example.com/video.mp4",
		Duration: 120,
		UserID:   "user-1",
	}

	result := ConvertMediaToMediaPB(from)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Id != "media-1" {
		t.Errorf("expected Id=media-1, got %s", result.Id)
	}
	if result.Title != "Test Video" {
		t.Errorf("expected Title=Test Video, got %s", result.Title)
	}
	if result.Type != "video" {
		t.Errorf("expected Type=video, got %s", result.Type)
	}
	if result.Url != "https://example.com/video.mp4" {
		t.Errorf("expected Url=https://example.com/video.mp4, got %s", result.Url)
	}
	if result.Duration != 120 {
		t.Errorf("expected Duration=120, got %d", result.Duration)
	}
	if result.UserId != "user-1" {
		t.Errorf("expected UserId=user-1, got %s", result.UserId)
	}
	// No edges set, so edge fields should be nil
	if result.User != nil {
		t.Errorf("expected nil User edge, got %v", result.User)
	}
	if result.Category != nil {
		t.Errorf("expected nil Category edge, got %v", result.Category)
	}
	if result.Channel != nil {
		t.Errorf("expected nil Channel edge, got %v", result.Channel)
	}
}

func TestConvertMediaToMediaPB_WithUserEdge(t *testing.T) {
	from := &entity.Media{
		ID:    "media-1",
		Title: "Test Video",
		Edges: entity.MediaEdges{
			User: &entity.User{
				ID:       "user-1",
				Username: "testuser",
				Email:    "test@example.com",
			},
		},
	}

	result := ConvertMediaToMediaPB(from)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.User == nil {
		t.Fatal("expected non-nil User edge")
	}
	if result.User.Id != "user-1" {
		t.Errorf("expected User.Id=user-1, got %s", result.User.Id)
	}
	if result.User.Username != "testuser" {
		t.Errorf("expected User.Username=testuser, got %s", result.User.Username)
	}
	if result.User.Email != "test@example.com" {
		t.Errorf("expected User.Email=test@example.com, got %s", result.User.Email)
	}
}

func TestConvertMediaToMediaPB_WithCategoryEdge(t *testing.T) {
	from := &entity.Media{
		ID:    "media-1",
		Title: "Test Video",
		Edges: entity.MediaEdges{
			Category: &entity.Category{
				ID:   42,
				Name: "Tech",
				Slug: "tech",
			},
		},
	}

	result := ConvertMediaToMediaPB(from)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Category == nil {
		t.Fatal("expected non-nil Category edge")
	}
	if result.Category.Id != 42 {
		t.Errorf("expected Category.Id=42, got %d", result.Category.Id)
	}
	if result.Category.Name != "Tech" {
		t.Errorf("expected Category.Name=Tech, got %s", result.Category.Name)
	}
	if result.Category.Slug != "tech" {
		t.Errorf("expected Category.Slug=tech, got %s", result.Category.Slug)
	}
}

func TestConvertMediaToMediaPB_WithChannelEdge(t *testing.T) {
	from := &entity.Media{
		ID:    "media-1",
		Title: "Test Video",
		Edges: entity.MediaEdges{
			Channel: &entity.Channel{
				ID:    "ch-1",
				Title: "My Channel",
			},
		},
	}

	result := ConvertMediaToMediaPB(from)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Channel == nil {
		t.Fatal("expected non-nil Channel edge")
	}
	if result.Channel.Id != "ch-1" {
		t.Errorf("expected Channel.Id=ch-1, got %s", result.Channel.Id)
	}
	if result.Channel.Title != "My Channel" {
		t.Errorf("expected Channel.Title=My Channel, got %s", result.Channel.Title)
	}
}

func TestConvertMediaToMediaPB_AllEdges(t *testing.T) {
	from := &entity.Media{
		ID:    "media-1",
		Title: "Test Video",
		Edges: entity.MediaEdges{
			User: &entity.User{
				ID:       "user-1",
				Username: "john",
				Email:    "john@example.com",
			},
			Category: &entity.Category{
				ID:   5,
				Name: "Music",
				Slug: "music",
			},
			Channel: &entity.Channel{
				ID:    "ch-1",
				Title: "John's Channel",
			},
		},
	}

	result := ConvertMediaToMediaPB(from)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Verify User edge
	if result.User == nil {
		t.Fatal("expected non-nil User edge")
	}
	if result.User.Id != "user-1" {
		t.Errorf("expected User.Id=user-1, got %s", result.User.Id)
	}
	if result.User.Username != "john" {
		t.Errorf("expected User.Username=john, got %s", result.User.Username)
	}
	// Verify Category edge
	if result.Category == nil {
		t.Fatal("expected non-nil Category edge")
	}
	if result.Category.Id != 5 {
		t.Errorf("expected Category.Id=5, got %d", result.Category.Id)
	}
	if result.Category.Name != "Music" {
		t.Errorf("expected Category.Name=Music, got %s", result.Category.Name)
	}
	// Verify Channel edge
	if result.Channel == nil {
		t.Fatal("expected non-nil Channel edge")
	}
	if result.Channel.Id != "ch-1" {
		t.Errorf("expected Channel.Id=ch-1, got %s", result.Channel.Id)
	}
	if result.Channel.Title != "John's Channel" {
		t.Errorf("expected Channel.Title=John's Channel, got %s", result.Channel.Title)
	}
}
