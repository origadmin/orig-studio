package migrate

import (
	"context"
	"io"
	"time"
)

type Phase string

const (
	PhaseDiscover   Phase = "discover"
	PhaseUsers      Phase = "users"
	PhaseCategories Phase = "categories"
	PhaseTags       Phase = "tags"
	PhaseMedia      Phase = "media"
	PhaseComments   Phase = "comments"
	PhaseLikes      Phase = "likes"
	PhasePlaylists  Phase = "playlists"
	PhaseChannels   Phase = "channels"
	PhaseFiles      Phase = "files"
	PhaseVerify     Phase = "verify"
	PhaseComplete   Phase = "complete"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusRunning    Status = "running"
	StatusPaused     Status = "paused"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
)

type SourceConfig struct {
	Type     string            `json:"type" yaml:"type"`
	DSN      string            `json:"dsn" yaml:"dsn"`
	Options  map[string]string `json:"options" yaml:"options"`
	MediaDir string            `json:"media_dir" yaml:"media_dir"`
}

type TargetConfig struct {
	DSN       string `json:"dsn" yaml:"dsn"`
	Dialect   string `json:"dialect" yaml:"dialect"`
	MediaDir  string `json:"media_dir" yaml:"media_dir"`
	Overwrite bool   `json:"overwrite" yaml:"overwrite"`
	DryRun    bool   `json:"dry_run" yaml:"dry_run"`
}

type Progress struct {
	Phase       Phase     `json:"phase"`
	Status      Status    `json:"status"`
	TotalItems  int64     `json:"total_items"`
	DoneItems   int64     `json:"done_items"`
	FailedItems int64     `json:"failed_items"`
	CurrentItem string    `json:"current_item"`
	StartedAt   time.Time `json:"started_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Error       string    `json:"error,omitempty"`
}

type SourceStats struct {
	Users      int `json:"users"`
	Media      int `json:"media"`
	Categories int `json:"categories"`
	Tags       int `json:"tags"`
	Comments   int `json:"comments"`
	Channels   int `json:"channels"`
	Playlists  int `json:"playlists"`
	Likes      int `json:"likes"`
}

type UserRecord struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar"`
	Bio         string `json:"bio"`
	Role        string `json:"role"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
}

type CategoryRecord struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	ParentID    string `json:"parent_id,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
}

type TagRecord struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type MediaRecord struct {
	ID          string            `json:"id"`
	UserID      string            `json:"user_id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Type        string            `json:"type"`
	FilePath    string            `json:"file_path"`
	FileName    string            `json:"file_name"`
	FileSize    int64             `json:"file_size"`
	Duration    float64           `json:"duration"`
	Width       int               `json:"width"`
	Height      int               `json:"height"`
	MimeType    string            `json:"mime_type"`
	Checksum    string            `json:"checksum"`
	Thumbnail   string            `json:"thumbnail"`
	Tags        []string          `json:"tags"`
	CategoryID  string            `json:"category_id,omitempty"`
	ChannelID   string            `json:"channel_id,omitempty"`
	Privacy     string            `json:"privacy"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   string            `json:"created_at"`
}

type CommentRecord struct {
	ID        string `json:"id"`
	MediaID   string `json:"media_id"`
	UserID    string `json:"user_id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

type ChannelRecord struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Banner      string `json:"banner"`
	CreatedAt   string `json:"created_at"`
}

type PlaylistRecord struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Privacy     string `json:"privacy"`
	MediaIDs    []string `json:"media_ids"`
	CreatedAt   string `json:"created_at"`
}

type FileRef struct {
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	Size       int64  `json:"size"`
	Checksum   string `json:"checksum"`
}

type SourceAdapter interface {
	Name() string
	Connect(ctx context.Context, cfg *SourceConfig) error
	Discover(ctx context.Context) (*SourceStats, error)
	Close() error

	Users(ctx context.Context) (Iterator[*UserRecord], error)
	Categories(ctx context.Context) (Iterator[*CategoryRecord], error)
	Tags(ctx context.Context) (Iterator[*TagRecord], error)
	Media(ctx context.Context) (Iterator[*MediaRecord], error)
	Comments(ctx context.Context) (Iterator[*CommentRecord], error)
	Channels(ctx context.Context) (Iterator[*ChannelRecord], error)
	Playlists(ctx context.Context) (Iterator[*PlaylistRecord], error)
	FileRefs(ctx context.Context, media *MediaRecord) ([]*FileRef, error)
	OpenFile(ctx context.Context, ref *FileRef) (io.ReadCloser, error)
}

type Iterator[T any] interface {
	Next(ctx context.Context) bool
	Item() T
	Err() error
	Close() error
}

type ProgressReporter interface {
	UpdateProgress(p *Progress)
	ReportError(phase Phase, item string, err error)
	ReportWarning(phase Phase, item string, msg string)
}

type IDMapper interface {
	Map(sourceType, sourceID string) (targetID string, ok bool)
	Set(sourceType, sourceID, targetID string)
}

type Checksummer interface {
	Checksum(ctx context.Context, ref *FileRef) (string, error)
}
