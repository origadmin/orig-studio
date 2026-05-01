package convpb

import (
	"testing"
	"time"

	"origadmin/application/origcms/api/gen/v1/types"
	"origadmin/application/origcms/internal/data/entity"
	"origadmin/application/origcms/internal/data/entity/category"
	"origadmin/application/origcms/internal/data/entity/channel"
	"origadmin/application/origcms/internal/data/entity/comment"
	"origadmin/application/origcms/internal/data/entity/media"
	"origadmin/application/origcms/internal/data/entity/playlist"
	"origadmin/application/origcms/internal/data/entity/tag"
	"origadmin/application/origcms/internal/data/entity/user"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMediaPrivacyFromPB(t *testing.T) {
	tests := []struct {
		input    types.Privacy
		expected media.Privacy
	}{
		{types.Privacy_PRIVACY_PUBLIC, media.PrivacyPUBLIC},
		{types.Privacy_PRIVACY_PRIVATE, media.PrivacyPRIVATE},
		{types.Privacy_PRIVACY_UNLISTED, media.PrivacyUNLISTED},
		{types.Privacy_PRIVACY_PAID, media.PrivacyPAID},
		{types.Privacy_PRIVACY_UNSPECIFIED, media.PrivacyPUBLIC},
	}
	for _, tt := range tests {
		result := ConvertPrivacyPBToMediaPrivacy(tt.input)
		if result != tt.expected {
			t.Errorf("ConvertPrivacyPBToMediaPrivacy(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestMediaPrivacyToPB(t *testing.T) {
	tests := []struct {
		input    media.Privacy
		expected types.Privacy
	}{
		{media.PrivacyPUBLIC, types.Privacy_PRIVACY_PUBLIC},
		{media.PrivacyPRIVATE, types.Privacy_PRIVACY_PRIVATE},
		{media.PrivacyUNLISTED, types.Privacy_PRIVACY_UNLISTED},
		{media.PrivacyPAID, types.Privacy_PRIVACY_PAID},
		{media.Privacy("UNKNOWN"), types.Privacy_PRIVACY_PUBLIC},
	}
	for _, tt := range tests {
		result := ConvertMediaPrivacyToPrivacyPB(tt.input)
		if result != tt.expected {
			t.Errorf("ConvertMediaPrivacyToPrivacyPB(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestChannelPrivacyFromPB(t *testing.T) {
	tests := []struct {
		input    types.Privacy
		expected channel.Privacy
	}{
		{types.Privacy_PRIVACY_PUBLIC, channel.PrivacyPUBLIC},
		{types.Privacy_PRIVACY_PRIVATE, channel.PrivacyPRIVATE},
		{types.Privacy_PRIVACY_UNLISTED, channel.PrivacyUNLISTED},
		{types.Privacy_PRIVACY_PAID, channel.PrivacyPAID},
		{types.Privacy_PRIVACY_SUBSCRIBERS_ONLY, channel.PrivacySUBSCRIBERS_ONLY},
		{types.Privacy_PRIVACY_UNSPECIFIED, channel.PrivacyPUBLIC},
	}
	for _, tt := range tests {
		result := ConvertPrivacyPBToChannelPrivacy(tt.input)
		if result != tt.expected {
			t.Errorf("ConvertPrivacyPBToChannelPrivacy(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestChannelPrivacyToPB(t *testing.T) {
	tests := []struct {
		input    channel.Privacy
		expected types.Privacy
	}{
		{channel.PrivacyPUBLIC, types.Privacy_PRIVACY_PUBLIC},
		{channel.PrivacyPRIVATE, types.Privacy_PRIVACY_PRIVATE},
		{channel.PrivacyUNLISTED, types.Privacy_PRIVACY_UNLISTED},
		{channel.PrivacyPAID, types.Privacy_PRIVACY_PAID},
		{channel.PrivacySUBSCRIBERS_ONLY, types.Privacy_PRIVACY_SUBSCRIBERS_ONLY},
		{channel.Privacy("UNKNOWN"), types.Privacy_PRIVACY_PUBLIC},
	}
	for _, tt := range tests {
		result := ConvertChannelPrivacyToPrivacyPB(tt.input)
		if result != tt.expected {
			t.Errorf("ConvertChannelPrivacyToPrivacyPB(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestPlaylistPrivacyFromPB(t *testing.T) {
	tests := []struct {
		input    types.Privacy
		expected playlist.Privacy
	}{
		{types.Privacy_PRIVACY_PUBLIC, playlist.PrivacyPUBLIC},
		{types.Privacy_PRIVACY_PRIVATE, playlist.PrivacyPRIVATE},
		{types.Privacy_PRIVACY_UNLISTED, playlist.PrivacyUNLISTED},
		{types.Privacy_PRIVACY_PAID, playlist.PrivacyPAID},
		{types.Privacy_PRIVACY_UNSPECIFIED, playlist.PrivacyPUBLIC},
	}
	for _, tt := range tests {
		result := ConvertPrivacyPBToPlaylistPrivacy(tt.input)
		if result != tt.expected {
			t.Errorf("ConvertPrivacyPBToPlaylistPrivacy(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestPlaylistPrivacyToPB(t *testing.T) {
	tests := []struct {
		input    playlist.Privacy
		expected types.Privacy
	}{
		{playlist.PrivacyPUBLIC, types.Privacy_PRIVACY_PUBLIC},
		{playlist.PrivacyPRIVATE, types.Privacy_PRIVACY_PRIVATE},
		{playlist.PrivacyUNLISTED, types.Privacy_PRIVACY_UNLISTED},
		{playlist.PrivacyPAID, types.Privacy_PRIVACY_PAID},
		{playlist.Privacy("UNKNOWN"), types.Privacy_PRIVACY_PUBLIC},
	}
	for _, tt := range tests {
		result := ConvertPlaylistPrivacyToPrivacyPB(tt.input)
		if result != tt.expected {
			t.Errorf("ConvertPlaylistPrivacyToPrivacyPB(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestPlaylistStatusFromPB(t *testing.T) {
	tests := []struct {
		input    types.PlaylistStatus
		expected playlist.Status
	}{
		{types.PlaylistStatus_PLAYLIST_STATUS_ACTIVE, playlist.StatusACTIVE},
		{types.PlaylistStatus_PLAYLIST_STATUS_INACTIVE, playlist.StatusINACTIVE},
		{types.PlaylistStatus_PLAYLIST_STATUS_DRAFT, playlist.StatusDRAFT},
		{types.PlaylistStatus_PLAYLIST_STATUS_ARCHIVED, playlist.StatusARCHIVED},
		{types.PlaylistStatus_PLAYLIST_STATUS_UNSPECIFIED, playlist.StatusACTIVE},
	}
	for _, tt := range tests {
		result := ConvertPlaylistStatusPBToPlaylistStatus(tt.input)
		if result != tt.expected {
			t.Errorf("ConvertPlaylistStatusPBToPlaylistStatus(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestPlaylistStatusToPB(t *testing.T) {
	tests := []struct {
		input    playlist.Status
		expected types.PlaylistStatus
	}{
		{playlist.StatusACTIVE, types.PlaylistStatus_PLAYLIST_STATUS_ACTIVE},
		{playlist.StatusINACTIVE, types.PlaylistStatus_PLAYLIST_STATUS_INACTIVE},
		{playlist.StatusDRAFT, types.PlaylistStatus_PLAYLIST_STATUS_DRAFT},
		{playlist.StatusARCHIVED, types.PlaylistStatus_PLAYLIST_STATUS_ARCHIVED},
		{playlist.Status("UNKNOWN"), types.PlaylistStatus_PLAYLIST_STATUS_ACTIVE},
	}
	for _, tt := range tests {
		result := ConvertPlaylistStatusToPlaylistStatusPB(tt.input)
		if result != tt.expected {
			t.Errorf("ConvertPlaylistStatusToPlaylistStatusPB(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestTagStatusFromPB(t *testing.T) {
	tests := []struct {
		input    types.TagStatus
		expected tag.Status
	}{
		{types.TagStatus_TAG_STATUS_ACTIVE, tag.StatusACTIVE},
		{types.TagStatus_TAG_STATUS_INACTIVE, tag.StatusINACTIVE},
		{types.TagStatus_TAG_STATUS_UNSPECIFIED, tag.StatusACTIVE},
	}
	for _, tt := range tests {
		result := ConvertTagStatusPBToTagStatus(tt.input)
		if result != tt.expected {
			t.Errorf("ConvertTagStatusPBToTagStatus(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestTagStatusToPB(t *testing.T) {
	tests := []struct {
		input    tag.Status
		expected types.TagStatus
	}{
		{tag.StatusACTIVE, types.TagStatus_TAG_STATUS_ACTIVE},
		{tag.StatusINACTIVE, types.TagStatus_TAG_STATUS_INACTIVE},
		{tag.Status("UNKNOWN"), types.TagStatus_TAG_STATUS_ACTIVE},
	}
	for _, tt := range tests {
		result := ConvertTagStatusToTagStatusPB(tt.input)
		if result != tt.expected {
			t.Errorf("ConvertTagStatusToTagStatusPB(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestUserStatusFromPB(t *testing.T) {
	tests := []struct {
		input    types.UserStatus
		expected user.Status
	}{
		{types.UserStatus_USER_STATUS_ACTIVE, user.StatusACTIVE},
		{types.UserStatus_USER_STATUS_INACTIVE, user.StatusINACTIVE},
		{types.UserStatus_USER_STATUS_PENDING, user.StatusPENDING},
		{types.UserStatus_USER_STATUS_SUSPENDED, user.StatusSUSPENDED},
		{types.UserStatus_USER_STATUS_REJECTED, user.StatusREJECTED},
		{types.UserStatus_USER_STATUS_UNSPECIFIED, user.StatusACTIVE},
	}
	for _, tt := range tests {
		result := ConvertUserStatusPBToUserStatus(tt.input)
		if result != tt.expected {
			t.Errorf("ConvertUserStatusPBToUserStatus(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestUserStatusToPB(t *testing.T) {
	tests := []struct {
		input    user.Status
		expected types.UserStatus
	}{
		{user.StatusACTIVE, types.UserStatus_USER_STATUS_ACTIVE},
		{user.StatusINACTIVE, types.UserStatus_USER_STATUS_INACTIVE},
		{user.StatusPENDING, types.UserStatus_USER_STATUS_PENDING},
		{user.StatusSUSPENDED, types.UserStatus_USER_STATUS_SUSPENDED},
		{user.StatusREJECTED, types.UserStatus_USER_STATUS_REJECTED},
		{user.Status("UNKNOWN"), types.UserStatus_USER_STATUS_ACTIVE},
	}
	for _, tt := range tests {
		result := ConvertUserStatusToUserStatusPB(tt.input)
		if result != tt.expected {
			t.Errorf("ConvertUserStatusToUserStatusPB(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestUserRoleFromPB(t *testing.T) {
	tests := []struct {
		input    string
		expected user.Role
	}{
		{"admin", user.RoleAdmin},
		{"editor", user.RoleEditor},
		{"user", user.RoleUser},
		{"unknown", user.RoleUser},
	}
	for _, tt := range tests {
		result := ConvertStringToUserRole(tt.input)
		if result != tt.expected {
			t.Errorf("ConvertStringToUserRole(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestUserRoleToPB(t *testing.T) {
	tests := []struct {
		input    user.Role
		expected string
	}{
		{user.RoleAdmin, "admin"},
		{user.RoleEditor, "editor"},
		{user.RoleUser, "user"},
	}
	for _, tt := range tests {
		result := ConvertUserRoleToString(tt.input)
		if result != tt.expected {
			t.Errorf("ConvertUserRoleToString(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestCategoryStatusFromInt32(t *testing.T) {
	tests := []struct {
		input    int32
		expected category.Status
	}{
		{1, category.StatusACTIVE},
		{2, category.StatusINACTIVE},
		{99, category.StatusACTIVE},
	}
	for _, tt := range tests {
		result := ConvertInt32ToCategoryStatus(tt.input)
		if result != tt.expected {
			t.Errorf("ConvertInt32ToCategoryStatus(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestCategoryStatusToInt32(t *testing.T) {
	tests := []struct {
		input    category.Status
		expected int32
	}{
		{category.StatusACTIVE, 1},
		{category.StatusINACTIVE, 2},
	}
	for _, tt := range tests {
		result := ConvertCategoryStatusToInt32(tt.input)
		if result != tt.expected {
			t.Errorf("ConvertCategoryStatusToInt32(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestCommentTextContentMapping(t *testing.T) {
	commentEnt := &entity.Comment{
		ID:     "c-1",
		Text:   "hello world",
		UserID: "u-1",
		MediaID: "m-1",
		Status: comment.StatusPENDING,
	}
	pb := ConvertCommentToCommentPB(commentEnt)
	if pb.Content != "hello world" {
		t.Errorf("Comment.Text -> CommentPB.Content: got %q, want %q", pb.Content, "hello world")
	}

	pb2 := &types.Comment{
		Id:      "c-1",
		Content: "hello from pb",
		UserId:  "u-1",
		MediaId: "m-1",
	}
	ent := ConvertCommentPBToComment(pb2)
	if ent.Text != "hello from pb" {
		t.Errorf("CommentPB.Content -> Comment.Text: got %q, want %q", ent.Text, "hello from pb")
	}
}

func TestLikeTypeMapping(t *testing.T) {
	likeEnt := &entity.Like{
		ID:       "l-1",
		MediaID:  "m-1",
		UserID:   "u-1",
		LikeType: "like",
	}
	pb := ConvertLikeToLikePB(likeEnt)
	if pb.Type != "like" {
		t.Errorf("Like.LikeType -> LikePB.Type: got %q, want %q", pb.Type, "like")
	}

	pb2 := &types.Like{
		Id:     "l-1",
		UserId: "u-1",
		MediaId: "m-1",
		Type:   "dislike",
	}
	ent := ConvertLikePBToLike(pb2)
	if ent.LikeType != "dislike" {
		t.Errorf("LikePB.Type -> Like.LikeType: got %q, want %q", ent.LikeType, "dislike")
	}
}

func TestTagTitleNameMapping(t *testing.T) {
	tagEnt := &entity.Tag{
		ID:    1,
		Title: "golang",
		Slug:  "go-lang",
	}
	pb := ConvertTagToTagPB(tagEnt)
	if pb.Name != "golang" {
		t.Errorf("Tag.Title -> TagPB.Name: got %q, want %q", pb.Name, "golang")
	}

	pb2 := &types.Tag{
		Id:   2,
		Name: "rust",
		Slug: "rust-lang",
	}
	ent := ConvertTagPBToTag(pb2)
	if ent.Title != "rust" {
		t.Errorf("TagPB.Name -> Tag.Title: got %q, want %q", ent.Title, "rust")
	}
}

func TestTimestampMapping(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)

	ent := &entity.Category{
		ID:        1,
		Name:      "test",
		CreateTime: now,
		UpdateTime: now,
	}
	pb := ConvertCategoryToCategoryPB(ent)
	if pb.CreateTime == nil {
		t.Fatal("Category.CreateTime -> CategoryPB.CreateTime: got nil timestamp")
	}
	if pb.UpdateTime == nil {
		t.Fatal("Category.UpdateTime -> CategoryPB.UpdateTime: got nil timestamp")
	}
	if !pb.CreateTime.AsTime().Equal(now) {
		t.Errorf("Category.CreateTime -> CategoryPB.CreateTime: got %v, want %v", pb.CreateTime.AsTime(), now)
	}
	if !pb.UpdateTime.AsTime().Equal(now) {
		t.Errorf("Category.UpdateTime -> CategoryPB.UpdateTime: got %v, want %v", pb.UpdateTime.AsTime(), now)
	}

	pb2 := &types.Category{
		Id:         1,
		Name:       "test",
		CreateTime: timestamppb.New(now),
		UpdateTime: timestamppb.New(now),
	}
	ent2 := ConvertCategoryPBToCategory(pb2)
	if !ent2.CreateTime.Equal(now) {
		t.Errorf("CategoryPB.CreateTime -> Category.CreateTime: got %v, want %v", ent2.CreateTime, now)
	}
	if !ent2.UpdateTime.Equal(now) {
		t.Errorf("CategoryPB.UpdateTime -> Category.UpdateTime: got %v, want %v", ent2.UpdateTime, now)
	}
}

func TestTimestampMapping_ZeroTime(t *testing.T) {
	ent := &entity.Category{
		ID:        1,
		Name:      "test",
		CreateTime: time.Time{},
		UpdateTime: time.Time{},
	}
	pb := ConvertCategoryToCategoryPB(ent)
	if pb.CreateTime != nil {
		t.Errorf("zero CreateTime should produce nil CreateTime, got %v", pb.CreateTime)
	}
	if pb.UpdateTime != nil {
		t.Errorf("zero UpdateTime should produce nil UpdateTime, got %v", pb.UpdateTime)
	}
}

func TestTimestampMapping_NilTimestamp(t *testing.T) {
	pb := &types.Category{
		Id:         1,
		Name:       "test",
		CreateTime: nil,
		UpdateTime: nil,
	}
	ent := ConvertCategoryPBToCategory(pb)
	if !ent.CreateTime.IsZero() {
		t.Errorf("nil CreateTime should produce zero CreateTime, got %v", ent.CreateTime)
	}
	if !ent.UpdateTime.IsZero() {
		t.Errorf("nil UpdateTime should produce zero UpdateTime, got %v", ent.UpdateTime)
	}
}

func TestChannelPrivacyRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	ent := &entity.Channel{
		ID:              "ch-1",
		UserID:          "u-1",
		Title:           "Test Channel",
		Privacy:         channel.PrivacySUBSCRIBERS_ONLY,
		SubscriberCount: 100,
		MediaCount:      5,
		CreateTime:       now,
		UpdateTime:       now,
	}
	pb := ConvertChannelToChannelPB(ent)
	if pb.Privacy != types.Privacy_PRIVACY_SUBSCRIBERS_ONLY {
		t.Errorf("Channel privacy to PB: got %v, want %v", pb.Privacy, types.Privacy_PRIVACY_SUBSCRIBERS_ONLY)
	}
	ent2 := ConvertChannelPBToChannel(pb)
	if ent2.Privacy != channel.PrivacySUBSCRIBERS_ONLY {
		t.Errorf("Channel privacy from PB: got %v, want %v", ent2.Privacy, channel.PrivacySUBSCRIBERS_ONLY)
	}
}

func TestPlaylistPrivacyAndStatusRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	ent := &entity.Playlist{
		ID:          "pl-1",
		Title:       "Test Playlist",
		UserID:      "u-1",
		Privacy:     playlist.PrivacyPAID,
		Status:      playlist.StatusDRAFT,
		MediaCount:  3,
		CreateTime:   now,
		UpdateTime:   now,
	}
	pb := ConvertPlaylistToPlaylistPB(ent)
	if pb.Privacy != types.Privacy_PRIVACY_PAID {
		t.Errorf("Playlist privacy to PB: got %v, want %v", pb.Privacy, types.Privacy_PRIVACY_PAID)
	}
	if pb.Status != types.PlaylistStatus_PLAYLIST_STATUS_DRAFT {
		t.Errorf("Playlist status to PB: got %v, want %v", pb.Status, types.PlaylistStatus_PLAYLIST_STATUS_DRAFT)
	}
	ent2 := ConvertPlaylistPBToPlaylist(pb)
	if ent2.Privacy != playlist.PrivacyPAID {
		t.Errorf("Playlist privacy from PB: got %v, want %v", ent2.Privacy, playlist.PrivacyPAID)
	}
	if ent2.Status != playlist.StatusDRAFT {
		t.Errorf("Playlist status from PB: got %v, want %v", ent2.Status, playlist.StatusDRAFT)
	}
}

func TestMediaPrivacyRoundTrip(t *testing.T) {
	ent := &entity.Media{
		ID:      "m-1",
		Title:   "Test Media",
		Privacy: media.PrivacyPAID,
	}
	pb := ConvertMediaToMediaPB(ent)
	if pb.Privacy != types.Privacy_PRIVACY_PAID {
		t.Errorf("Media privacy to PB: got %v, want %v", pb.Privacy, types.Privacy_PRIVACY_PAID)
	}
	ent2 := ConvertMediaPBToMedia(pb)
	if ent2.Privacy != media.PrivacyPAID {
		t.Errorf("Media privacy from PB: got %v, want %v", ent2.Privacy, media.PrivacyPAID)
	}
}

func TestUserStatusSuspendedRejected(t *testing.T) {
	ent := &entity.User{
		ID:       "u-1",
		Username: "suspended_user",
		Status:   user.StatusSUSPENDED,
		Role:     user.RoleUser,
	}
	pb := ConvertUserToUserPB(ent)
	if pb.Status != types.UserStatus_USER_STATUS_SUSPENDED {
		t.Errorf("User SUSPENDED to PB: got %v, want %v", pb.Status, types.UserStatus_USER_STATUS_SUSPENDED)
	}

	ent2 := &entity.User{
		ID:       "u-2",
		Username: "rejected_user",
		Status:   user.StatusREJECTED,
		Role:     user.RoleUser,
	}
	pb2 := ConvertUserToUserPB(ent2)
	if pb2.Status != types.UserStatus_USER_STATUS_REJECTED {
		t.Errorf("User REJECTED to PB: got %v, want %v", pb2.Status, types.UserStatus_USER_STATUS_REJECTED)
	}

	pb3 := &types.User{
		Id:     "u-3",
		Status: types.UserStatus_USER_STATUS_SUSPENDED,
	}
	ent3 := ConvertUserPBToUser(pb3)
	if ent3.Status != user.StatusSUSPENDED {
		t.Errorf("User SUSPENDED from PB: got %v, want %v", ent3.Status, user.StatusSUSPENDED)
	}

	pb4 := &types.User{
		Id:     "u-4",
		Status: types.UserStatus_USER_STATUS_REJECTED,
	}
	ent4 := ConvertUserPBToUser(pb4)
	if ent4.Status != user.StatusREJECTED {
		t.Errorf("User REJECTED from PB: got %v, want %v", ent4.Status, user.StatusREJECTED)
	}
}

func TestConvertNilInputs(t *testing.T) {
	if ConvertChannelPBToChannel(nil) != nil {
		t.Error("ConvertChannelPBToChannel(nil) should return nil")
	}
	if ConvertChannelToChannelPB(nil) != nil {
		t.Error("ConvertChannelToChannelPB(nil) should return nil")
	}
	if ConvertCommentPBToComment(nil) != nil {
		t.Error("ConvertCommentPBToComment(nil) should return nil")
	}
	if ConvertCommentToCommentPB(nil) != nil {
		t.Error("ConvertCommentToCommentPB(nil) should return nil")
	}
	if ConvertLikePBToLike(nil) != nil {
		t.Error("ConvertLikePBToLike(nil) should return nil")
	}
	if ConvertLikeToLikePB(nil) != nil {
		t.Error("ConvertLikeToLikePB(nil) should return nil")
	}
	if ConvertTagPBToTag(nil) != nil {
		t.Error("ConvertTagPBToTag(nil) should return nil")
	}
	if ConvertTagToTagPB(nil) != nil {
		t.Error("ConvertTagToTagPB(nil) should return nil")
	}
	if ConvertUserPBToUser(nil) != nil {
		t.Error("ConvertUserPBToUser(nil) should return nil")
	}
	if ConvertUserToUserPB(nil) != nil {
		t.Error("ConvertUserToUserPB(nil) should return nil")
	}
	if ConvertPlaylistPBToPlaylist(nil) != nil {
		t.Error("ConvertPlaylistPBToPlaylist(nil) should return nil")
	}
	if ConvertPlaylistToPlaylistPB(nil) != nil {
		t.Error("ConvertPlaylistToPlaylistPB(nil) should return nil")
	}
	if ConvertCategoryPBToCategory(nil) != nil {
		t.Error("ConvertCategoryPBToCategory(nil) should return nil")
	}
	if ConvertCategoryToCategoryPB(nil) != nil {
		t.Error("ConvertCategoryToCategoryPB(nil) should return nil")
	}
	if ConvertMediaPBToMedia(nil) != nil {
		t.Error("ConvertMediaPBToMedia(nil) should return nil")
	}
}

func TestConvertTimeToTimestamp(t *testing.T) {
	if ConvertTimeToTimestamp(time.Time{}) != nil {
		t.Error("zero time should produce nil timestamp")
	}
	now := time.Now().Truncate(time.Millisecond)
	ts := ConvertTimeToTimestamp(now)
	if ts == nil {
		t.Fatal("non-zero time should produce non-nil timestamp")
	}
	if !ts.AsTime().Equal(now) {
		t.Errorf("timestamp round-trip: got %v, want %v", ts.AsTime(), now)
	}
}

func TestConvertTimestampToTime(t *testing.T) {
	if !ConvertTimestampToTime(nil).IsZero() {
		t.Error("nil timestamp should produce zero time")
	}
	now := time.Now().Truncate(time.Millisecond)
	result := ConvertTimestampToTime(timestamppb.New(now))
	if !result.Equal(now) {
		t.Errorf("timestamp to time: got %v, want %v", result, now)
	}
}

func TestUserRoleRoundTrip(t *testing.T) {
	ent := &entity.User{
		ID:       "u-1",
		Username: "admin_user",
		Role:     user.RoleAdmin,
	}
	pb := ConvertUserToUserPB(ent)
	if pb.Role != "admin" {
		t.Errorf("User role admin to PB: got %q, want %q", pb.Role, "admin")
	}

	pb2 := &types.User{
		Id:     "u-2",
		Role:   "editor",
	}
	ent2 := ConvertUserPBToUser(pb2)
	if ent2.Role != user.RoleEditor {
		t.Errorf("User role editor from PB: got %v, want %v", ent2.Role, user.RoleEditor)
	}
}

func TestTagStatusRoundTrip(t *testing.T) {
	ent := &entity.Tag{
		ID:     1,
		Title:  "test",
		Status: tag.StatusINACTIVE,
	}
	pb := ConvertTagToTagPB(ent)
	if pb.Status != types.TagStatus_TAG_STATUS_INACTIVE {
		t.Errorf("Tag status INACTIVE to PB: got %v, want %v", pb.Status, types.TagStatus_TAG_STATUS_INACTIVE)
	}
	ent2 := ConvertTagPBToTag(pb)
	if ent2.Status != tag.StatusINACTIVE {
		t.Errorf("Tag status INACTIVE from PB: got %v, want %v", ent2.Status, tag.StatusINACTIVE)
	}
}

func TestCommentTimestampMapping(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	ent := &entity.Comment{
		ID:        "c-1",
		Text:      "test",
		UserID:    "u-1",
		MediaID:   "m-1",
		Status:    comment.StatusPENDING,
		CreateTime: now,
		UpdateTime: now,
	}
	pb := ConvertCommentToCommentPB(ent)
	if pb.CreateTime == nil || !pb.CreateTime.AsTime().Equal(now) {
		t.Errorf("Comment.CreateTime -> CommentPB.CreateTime: got %v, want %v", pb.CreateTime, now)
	}
	if pb.UpdateTime == nil || !pb.UpdateTime.AsTime().Equal(now) {
		t.Errorf("Comment.UpdateTime -> CommentPB.UpdateTime: got %v, want %v", pb.UpdateTime, now)
	}
}

func TestLikeTimestampMapping(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	ent := &entity.Like{
		ID:        "l-1",
		MediaID:   "m-1",
		UserID:    "u-1",
		LikeType:  "like",
		CreateTime: now,
	}
	pb := ConvertLikeToLikePB(ent)
	if pb.CreateTime == nil || !pb.CreateTime.AsTime().Equal(now) {
		t.Errorf("Like.CreateTime -> LikePB.CreateTime: got %v, want %v", pb.CreateTime, now)
	}
}

func TestTagTimestampMapping(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	ent := &entity.Tag{
		ID:        1,
		Title:     "test",
		Status:    tag.StatusACTIVE,
		CreateTime: now,
		UpdateTime: now,
	}
	pb := ConvertTagToTagPB(ent)
	if pb.CreateTime == nil || !pb.CreateTime.AsTime().Equal(now) {
		t.Errorf("Tag.CreateTime -> TagPB.CreateTime: got %v, want %v", pb.CreateTime, now)
	}
	if pb.UpdateTime == nil || !pb.UpdateTime.AsTime().Equal(now) {
		t.Errorf("Tag.UpdateTime -> TagPB.UpdateTime: got %v, want %v", pb.UpdateTime, now)
	}
}

func TestUserTimestampMapping(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	ent := &entity.User{
		ID:        "u-1",
		Username:  "testuser",
		Status:    user.StatusACTIVE,
		Role:      user.RoleUser,
		CreateTime: now,
		UpdateTime: now,
	}
	pb := ConvertUserToUserPB(ent)
	if pb.CreateTime == nil || !pb.CreateTime.AsTime().Equal(now) {
		t.Errorf("User.CreateTime -> UserPB.CreateTime: got %v, want %v", pb.CreateTime, now)
	}
	if pb.UpdateTime == nil || !pb.UpdateTime.AsTime().Equal(now) {
		t.Errorf("User.UpdateTime -> UserPB.UpdateTime: got %v, want %v", pb.UpdateTime, now)
	}
}
