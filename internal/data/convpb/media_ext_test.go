package convpb

import (
	"testing"

	"origadmin/application/origcms/internal/data/entity"
)

func TestConvertMediaWithEdges_NilInput(t *testing.T) {
	result := ConvertMediaWithEdges(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestConvertMediaWithEdges_NoEdges(t *testing.T) {
	from := &entity.Media{
		ID:    "test-id",
		Title: "Test Media",
	}

	result := ConvertMediaWithEdges(from)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Media == nil {
		t.Fatal("expected non-nil Media")
	}
	if result.Media.Id != "test-id" {
		t.Errorf("expected Id=test-id, got %s", result.Media.Id)
	}
	if result.Media.Title != "Test Media" {
		t.Errorf("expected Title=Test Media, got %s", result.Media.Title)
	}
	if result.User != nil {
		t.Errorf("expected nil User, got %v", result.User)
	}
	if result.Category != nil {
		t.Errorf("expected nil Category, got %v", result.Category)
	}
	if result.Channel != nil {
		t.Errorf("expected nil Channel, got %v", result.Channel)
	}
}

func TestConvertMediaWithEdges_WithUserEdge(t *testing.T) {
	from := &entity.Media{
		ID:     "media-1",
		Title:  "Test Media",
		UserID: "user-1",
	}
	from.Edges.User = &entity.User{
		ID:       "user-1",
		Username: "testuser",
		Name:     "Test Nickname",
		Logo:     "https://example.com/avatar.png",
	}

	result := ConvertMediaWithEdges(from)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.User == nil {
		t.Fatal("expected non-nil User edge")
	}
	if result.User.ID != "user-1" {
		t.Errorf("expected User.ID=user-1, got %s", result.User.ID)
	}
	if result.User.Username != "testuser" {
		t.Errorf("expected User.Username=testuser, got %s", result.User.Username)
	}
	if result.User.Nickname != "Test Nickname" {
		t.Errorf("expected User.Nickname=Test Nickname, got %s", result.User.Nickname)
	}
	if result.User.Avatar != "https://example.com/avatar.png" {
		t.Errorf("expected User.Avatar=https://example.com/avatar.png, got %s", result.User.Avatar)
	}
}

func TestConvertMediaWithEdges_WithCategoryEdge(t *testing.T) {
	from := &entity.Media{
		ID:         "media-1",
		Title:      "Test Media",
		CategoryID: 42,
	}
	from.Edges.Category = &entity.Category{
		ID:   42,
		Name: "Tech",
		Slug: "tech",
	}

	result := ConvertMediaWithEdges(from)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Category == nil {
		t.Fatal("expected non-nil Category edge")
	}
	if result.Category.ID != 42 {
		t.Errorf("expected Category.ID=42, got %d", result.Category.ID)
	}
	if result.Category.Name != "Tech" {
		t.Errorf("expected Category.Name=Tech, got %s", result.Category.Name)
	}
	if result.Category.Slug != "tech" {
		t.Errorf("expected Category.Slug=tech, got %s", result.Category.Slug)
	}
}

func TestConvertMediaWithEdges_WithChannelEdge(t *testing.T) {
	from := &entity.Media{
		ID:        "media-1",
		Title:     "Test Media",
		ChannelID: "ch-1",
	}
	from.Edges.Channel = &entity.Channel{
		ID:    "ch-1",
		Title: "My Channel",
	}

	result := ConvertMediaWithEdges(from)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Channel == nil {
		t.Fatal("expected non-nil Channel edge")
	}
	if result.Channel.ID != "ch-1" {
		t.Errorf("expected Channel.ID=ch-1, got %s", result.Channel.ID)
	}
	if result.Channel.Title != "My Channel" {
		t.Errorf("expected Channel.Title=My Channel, got %s", result.Channel.Title)
	}
}

func TestConvertMediaWithEdges_AllEdges(t *testing.T) {
	from := &entity.Media{
		ID:         "media-1",
		Title:      "Full Media",
		UserID:     "user-1",
		CategoryID: 5,
		ChannelID:  "ch-1",
	}
	from.Edges.User = &entity.User{
		ID:       "user-1",
		Username: "john",
		Name:     "John Doe",
		Logo:     "avatar.jpg",
	}
	from.Edges.Category = &entity.Category{
		ID:   5,
		Name: "Music",
		Slug: "music",
	}
	from.Edges.Channel = &entity.Channel{
		ID:    "ch-1",
		Title: "John's Channel",
	}

	result := ConvertMediaWithEdges(from)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.User == nil {
		t.Fatal("expected non-nil User edge")
	}
	if result.Category == nil {
		t.Fatal("expected non-nil Category edge")
	}
	if result.Channel == nil {
		t.Fatal("expected non-nil Channel edge")
	}
	// Verify base Media fields are still populated
	if result.Media.Id != "media-1" {
		t.Errorf("expected Media.Id=media-1, got %s", result.Media.Id)
	}
	if result.Media.UserId != "user-1" {
		t.Errorf("expected Media.UserId=user-1, got %s", result.Media.UserId)
	}
}
