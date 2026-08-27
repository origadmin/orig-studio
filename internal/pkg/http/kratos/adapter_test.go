package kratos

import (
	"testing"
)

func TestTranslatePath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain path",
			input: "/api/v1/medias",
			want:  "/api/v1/medias",
		},
		{
			name:  "single param",
			input: "/api/v1/medias/:id",
			want:  "/api/v1/medias/{id}",
		},
		{
			name:  "multiple params",
			input: "/api/v1/medias/:id/tasks/:task_id/retry",
			want:  "/api/v1/medias/{id}/tasks/{task_id}/retry",
		},
		{
			name:  "catch-all named",
			input: "/static/*filepath",
			want:  "/static/{filepath:.*}",
		},
		{
			name:  "catch-all unnamed",
			input: "/files/*",
			want:  "/files/{path:.*}",
		},
		{
			name:  "root path",
			input: "/",
			want:  "/",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "param at start",
			input: "/:version/items",
			want:  "/{version}/items",
		},
		{
			name:  "no leading slash",
			input: "medias/:id",
			want:  "medias/{id}",
		},
		{
			name:  "mixed params and literals",
			input: "/users/:user_id/channels/:channel_id/videos/:video_id",
			want:  "/users/{user_id}/channels/{channel_id}/videos/{video_id}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translatePath(tt.input)
			if got != tt.want {
				t.Errorf("translatePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// Interface compliance is verified at compile time by `var _` assertions
// in adapter.go:
//   var _ http2.Router  = (*RouterAdapter)(nil)
//   var _ http2.Context = (*contextWrapper)(nil)
// No runtime test needed — if the adapter doesn't implement the interface,
// the package won't compile.
