package enums

import (
	"testing"
)

// --- ReviewStatus tests ---

func TestParseReviewStatus(t *testing.T) {
	tests := []struct {
		input string
		want  ReviewStatus
	}{
		{"pending_review", ReviewStatusPendingReview},
		{"reviewed", ReviewStatusReviewed},
		{"rejected", ReviewStatusRejected},
		{"PENDING_REVIEW", ReviewStatusPendingReview},
		{"Reviewed", ReviewStatusReviewed},
		{"REJECTED", ReviewStatusRejected},
		{"", ReviewStatusUnknown},
		{"invalid_value", ReviewStatusUnknown},
	}
	for _, tt := range tests {
		got := ParseReviewStatus(tt.input)
		if got != tt.want {
			t.Errorf("ParseReviewStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestReviewStatusValues(t *testing.T) {
	vals := ReviewStatus("").Values()
	expected := []string{"unknown", "pending_review", "reviewed", "rejected"}
	if len(vals) != len(expected) {
		t.Fatalf("ReviewStatus.Values() returned %d values, want %d", len(vals), len(expected))
	}
	for i, v := range expected {
		if vals[i] != v {
			t.Errorf("ReviewStatus.Values()[%d] = %q, want %q", i, vals[i], v)
		}
	}
}

func TestReviewStatusString(t *testing.T) {
	s := ReviewStatusPendingReview
	if s.String() != "pending_review" {
		t.Errorf("ReviewStatusPendingReview.String() = %q, want %q", s.String(), "pending_review")
	}
}

// --- MediaState tests ---

func TestParseMediaState(t *testing.T) {
	tests := []struct {
		input string
		want  MediaState
	}{
		{"draft", MediaStateDraft},
		{"active", MediaStateActive},
		{"deleted", MediaStateDeleted},
		{"DRAFT", MediaStateDraft},
		{"Active", MediaStateActive},
		{"DELETED", MediaStateDeleted},
		{"", MediaStateUnknown},
		{"invalid_value", MediaStateUnknown},
	}
	for _, tt := range tests {
		got := ParseMediaState(tt.input)
		if got != tt.want {
			t.Errorf("ParseMediaState(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMediaStateValues(t *testing.T) {
	vals := MediaState("").Values()
	expected := []string{"unknown", "draft", "active", "deleted"}
	if len(vals) != len(expected) {
		t.Fatalf("MediaState.Values() returned %d values, want %d", len(vals), len(expected))
	}
	for i, v := range expected {
		if vals[i] != v {
			t.Errorf("MediaState.Values()[%d] = %q, want %q", i, vals[i], v)
		}
	}
}

func TestMediaStateString(t *testing.T) {
	s := MediaStateActive
	if s.String() != "active" {
		t.Errorf("MediaStateActive.String() = %q, want %q", s.String(), "active")
	}
}

// --- CommentStatus tests ---

func TestParseCommentStatus(t *testing.T) {
	tests := []struct {
		input string
		want  CommentStatus
	}{
		{"pending", CommentStatusPending},
		{"approved", CommentStatusApproved},
		{"rejected", CommentStatusRejected},
		{"reported", CommentStatusReported},
		{"PENDING", CommentStatusPending},
		{"Approved", CommentStatusApproved},
		{"", CommentStatusUnknown},
		{"invalid_value", CommentStatusUnknown},
	}
	for _, tt := range tests {
		got := ParseCommentStatus(tt.input)
		if got != tt.want {
			t.Errorf("ParseCommentStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCommentStatusValues(t *testing.T) {
	vals := CommentStatus("").Values()
	expected := []string{"unknown", "pending", "approved", "rejected", "reported"}
	if len(vals) != len(expected) {
		t.Fatalf("CommentStatus.Values() returned %d values, want %d", len(vals), len(expected))
	}
	for i, v := range expected {
		if vals[i] != v {
			t.Errorf("CommentStatus.Values()[%d] = %q, want %q", i, vals[i], v)
		}
	}
}

func TestCommentStatusString(t *testing.T) {
	s := CommentStatusApproved
	if s.String() != "approved" {
		t.Errorf("CommentStatusApproved.String() = %q, want %q", s.String(), "approved")
	}
}

// --- SpriteStatus tests ---

func TestParseSpriteStatus(t *testing.T) {
	tests := []struct {
		input string
		want  SpriteStatus
	}{
		{"pending", SpriteStatusPending},
		{"processing", SpriteStatusProcessing},
		{"completed", SpriteStatusCompleted},
		{"failed", SpriteStatusFailed},
		{"PENDING", SpriteStatusPending},
		{"Processing", SpriteStatusProcessing},
		{"", SpriteStatusUnknown},
		{"invalid_value", SpriteStatusUnknown},
	}
	for _, tt := range tests {
		got := ParseSpriteStatus(tt.input)
		if got != tt.want {
			t.Errorf("ParseSpriteStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSpriteStatusValues(t *testing.T) {
	vals := SpriteStatus("").Values()
	expected := []string{"unknown", "pending", "processing", "completed", "failed"}
	if len(vals) != len(expected) {
		t.Fatalf("SpriteStatus.Values() returned %d values, want %d", len(vals), len(expected))
	}
	for i, v := range expected {
		if vals[i] != v {
			t.Errorf("SpriteStatus.Values()[%d] = %q, want %q", i, vals[i], v)
		}
	}
}

func TestSpriteStatusString(t *testing.T) {
	s := SpriteStatusCompleted
	if s.String() != "completed" {
		t.Errorf("SpriteStatusCompleted.String() = %q, want %q", s.String(), "completed")
	}
}

// --- UserStatus tests ---

func TestParseUserStatus(t *testing.T) {
	tests := []struct {
		input string
		want  UserStatus
	}{
		{"active", UserStatusActive},
		{"inactive", UserStatusInactive},
		{"ACTIVE", UserStatusActive},
		{"Inactive", UserStatusInactive},
		{"", UserStatusUnknown},
		{"invalid_value", UserStatusUnknown},
	}
	for _, tt := range tests {
		got := ParseUserStatus(tt.input)
		if got != tt.want {
			t.Errorf("ParseUserStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestUserStatusValues(t *testing.T) {
	vals := UserStatus("").Values()
	expected := []string{"unknown", "active", "inactive"}
	if len(vals) != len(expected) {
		t.Fatalf("UserStatus.Values() returned %d values, want %d", len(vals), len(expected))
	}
	for i, v := range expected {
		if vals[i] != v {
			t.Errorf("UserStatus.Values()[%d] = %q, want %q", i, vals[i], v)
		}
	}
}

func TestUserStatusString(t *testing.T) {
	s := UserStatusActive
	if s.String() != "active" {
		t.Errorf("UserStatusActive.String() = %q, want %q", s.String(), "active")
	}
}

// --- TagStatus tests ---

func TestParseTagStatus(t *testing.T) {
	tests := []struct {
		input string
		want  TagStatus
	}{
		{"active", TagStatusActive},
		{"inactive", TagStatusInactive},
		{"ACTIVE", TagStatusActive},
		{"Inactive", TagStatusInactive},
		{"", TagStatusUnknown},
		{"invalid_value", TagStatusUnknown},
	}
	for _, tt := range tests {
		got := ParseTagStatus(tt.input)
		if got != tt.want {
			t.Errorf("ParseTagStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestTagStatusValues(t *testing.T) {
	vals := TagStatus("").Values()
	expected := []string{"unknown", "active", "inactive"}
	if len(vals) != len(expected) {
		t.Fatalf("TagStatus.Values() returned %d values, want %d", len(vals), len(expected))
	}
	for i, v := range expected {
		if vals[i] != v {
			t.Errorf("TagStatus.Values()[%d] = %q, want %q", i, vals[i], v)
		}
	}
}

func TestTagStatusString(t *testing.T) {
	s := TagStatusActive
	if s.String() != "active" {
		t.Errorf("TagStatusActive.String() = %q, want %q", s.String(), "active")
	}
}

// --- ChannelStatus tests ---

func TestParseChannelStatus(t *testing.T) {
	tests := []struct {
		input string
		want  ChannelStatus
	}{
		{"active", ChannelStatusActive},
		{"pending", ChannelStatusPending},
		{"inactive", ChannelStatusInactive},
		{"ACTIVE", ChannelStatusActive},
		{"Pending", ChannelStatusPending},
		{"", ChannelStatusUnknown},
		{"invalid_value", ChannelStatusUnknown},
	}
	for _, tt := range tests {
		got := ParseChannelStatus(tt.input)
		if got != tt.want {
			t.Errorf("ParseChannelStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestChannelStatusValues(t *testing.T) {
	vals := ChannelStatus("").Values()
	expected := []string{"unknown", "active", "pending", "inactive"}
	if len(vals) != len(expected) {
		t.Fatalf("ChannelStatus.Values() returned %d values, want %d", len(vals), len(expected))
	}
	for i, v := range expected {
		if vals[i] != v {
			t.Errorf("ChannelStatus.Values()[%d] = %q, want %q", i, vals[i], v)
		}
	}
}

func TestChannelStatusString(t *testing.T) {
	s := ChannelStatusActive
	if s.String() != "active" {
		t.Errorf("ChannelStatusActive.String() = %q, want %q", s.String(), "active")
	}
}

// --- Cross-cutting: Invalid sentinel equals Unknown ---

func TestInvalidSentinels(t *testing.T) {
	if ReviewStatusInvalid != ReviewStatusUnknown {
		t.Error("ReviewStatusInvalid should equal ReviewStatusUnknown")
	}
	if MediaStateInvalid != MediaStateUnknown {
		t.Error("MediaStateInvalid should equal MediaStateUnknown")
	}
	if CommentStatusInvalid != CommentStatusUnknown {
		t.Error("CommentStatusInvalid should equal CommentStatusUnknown")
	}
	if SpriteStatusInvalid != SpriteStatusUnknown {
		t.Error("SpriteStatusInvalid should equal SpriteStatusUnknown")
	}
	if UserStatusInvalid != UserStatusUnknown {
		t.Error("UserStatusInvalid should equal UserStatusUnknown")
	}
	if TagStatusInvalid != TagStatusUnknown {
		t.Error("TagStatusInvalid should equal TagStatusUnknown")
	}
	if ChannelStatusInvalid != ChannelStatusUnknown {
		t.Error("ChannelStatusInvalid should equal ChannelStatusUnknown")
	}
}
