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
	// Batch 3: M2M / derived entity phases
	PhasePlaylistMedia Phase = "playlist_media"
	PhaseSubtitles     Phase = "subtitles"
	PhaseSubscriptions Phase = "subscriptions"
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
	// MediaTypes restricts which source media types are migrated. Only
	// platform types the target supports end-to-end may be listed; the rest
	// stay in the source for a later import once support lands. Empty means
	// migrate all types.
	MediaTypes []string `json:"media_types" yaml:"media_types"`
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
	// Extended real-schema fields (users_user)
	Name          string `json:"name"`
	Title         string `json:"title"`
	Location      string `json:"location"`
	IsSuperuser   bool   `json:"is_superuser"`
	IsFeatured    bool   `json:"is_featured"`
	AdvancedUser  bool   `json:"advanced_user"`
	IsEditor      bool   `json:"is_editor"`
	IsManager     bool   `json:"is_manager"`
	MediaCount    int    `json:"media_count"`
	AllowContact  bool   `json:"allow_contact"`
	DateAdded     string `json:"date_added"`
	LastLogin     string `json:"last_login"`
	Notifications bool   `json:"notifications"`
}

type CategoryRecord struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	ParentID    string `json:"parent_id,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	// Extended real-schema fields (files_category)
	UID               string `json:"uid"`
	Thumbnail         string `json:"thumbnail"`
	ListingsThumbnail string `json:"listings_thumbnail"`
	UserID            string `json:"user_id"`
	IsGlobal          bool   `json:"is_global"`
	IsRBACCategory    bool   `json:"is_rbac_category"`
	IdentityProvider  string `json:"identity_provider"`
	MediaCount        int    `json:"media_count"`
	AddDate           string `json:"add_date"`
}

type TagRecord struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	// Extended real-schema fields (files_tag)
	MediaCount        int    `json:"media_count"`
	ListingsThumbnail string `json:"listings_thumbnail"`
	UserID            string `json:"user_id"`
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
	// Extended real-schema fields (files_media)
	Views          int     `json:"views"`
	Likes          int     `json:"likes"`
	Dislikes       int     `json:"dislikes"`
	State          string  `json:"state"`
	EncodingStatus string  `json:"encoding_status"`
	UID            string  `json:"uid"`
	HLSFile        string  `json:"hls_file"`
	Sprites        string  `json:"sprites"`
	Poster         string  `json:"poster"`
	PreviewFile    string  `json:"preview_file"`
	UploadedThumb  string  `json:"uploaded_thumb"`
	UploadedPoster string  `json:"uploaded_poster"`
	MediaInfo      string  `json:"media_info"`
	Md5sum         string  `json:"md5sum"`
	LicenseID      string  `json:"license_id"`
	FriendlyToken  string  `json:"friendly_token"`
	ThumbnailTime  float64 `json:"thumbnail_time"`
	AllowDownload  bool    `json:"allow_download"`
	EnableComments bool    `json:"enable_comments"`
	Featured       bool    `json:"featured"`
	Listable       bool    `json:"listable"`
	IsReviewed     bool    `json:"is_reviewed"`
	ReportedTimes  int     `json:"reported_times"`
	EditDate       string  `json:"edit_date"`
}

type CommentRecord struct {
	ID        string `json:"id"`
	MediaID   string `json:"media_id"`
	UserID    string `json:"user_id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
	// Extended real-schema fields (files_comment)
	ParentID string `json:"parent_id,omitempty"`
	UID      string `json:"uid"`
	Level    int    `json:"level"`
	TreeID   int    `json:"tree_id"`
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
	// Extended real-schema fields (users_channel)
	Title        string `json:"title"`
	FriendlyToken string `json:"friendly_token"`
	BannerLogo   string `json:"banner_logo"`
	AddDate      string `json:"add_date"`
}

type PlaylistRecord struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Privacy     string `json:"privacy"`
	MediaIDs    []string `json:"media_ids"`
	CreatedAt   string `json:"created_at"`
	// Extended real-schema fields (files_playlist)
	UID           string `json:"uid"`
	FriendlyToken string `json:"friendly_token"`
	AddDate       string `json:"add_date"`
}

// MediaTagRecord is a row of the M2M pivot files_media_tags.
type MediaTagRecord struct {
	MediaID string `json:"media_id"`
	TagID   string `json:"tag_id"`
}

// MediaCategoryRecord is a row of the M2M pivot files_media_category.
type MediaCategoryRecord struct {
	MediaID    string `json:"media_id"`
	CategoryID string `json:"category_id"`
}

// PlaylistMediaRecord is a row of the M2M pivot files_playlistmedia.
type PlaylistMediaRecord struct {
	PlaylistID string `json:"playlist_id"`
	MediaID    string `json:"media_id"`
	Ordering   int    `json:"ordering"`
	ActionDate string `json:"action_date"`
}

// SubtitleRecord joins files_subtitle with files_language (language flattened).
type SubtitleRecord struct {
	MediaID   string `json:"media_id"`
	Language  string `json:"language"`
	Label     string `json:"label"`
	FileURL   string `json:"file_url"`
	UserID    string `json:"user_id"`
}

// SubscriptionRecord is a row of the M2M pivot users_channel_subscribers.
type SubscriptionRecord struct {
	ChannelID    string `json:"channel_id"`
	SubscriberID string `json:"subscriber_id"`
}

// LicenseRecord is a row of files_license. A-side data with no B-side entity;
// it is folded into Media.metadata during phaseMedia.
type LicenseRecord struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// RatingRecord is a row of files_rating. Folded into Media.metadata.
type RatingRecord struct {
	ID               string `json:"id"`
	MediaID          string `json:"media_id"`
	UserID           string `json:"user_id"`
	RatingCategoryID string `json:"rating_category_id"`
	Score            int    `json:"score"`
	AddDate          string `json:"add_date"`
}

// EncodingRecord is a row of files_encoding. Folded into Media.metadata.
type EncodingRecord struct {
	ID            string `json:"id"`
	MediaID       string `json:"media_id"`
	ProfileID     string `json:"profile_id"`
	Status        string `json:"status"`
	Progress      int    `json:"progress"`
	MediaFile     string `json:"media_file"`
	Size          string `json:"size"`
	MD5Sum        string `json:"md5sum"`
	Chunk         bool   `json:"chunk"`
	ChunkFilePath string `json:"chunk_file_path"`
	ChunksInfo    string `json:"chunks_info"`
	Logs          string `json:"logs"`
	Commands      string `json:"commands"`
	TempFile      string `json:"temp_file"`
	TaskID        string `json:"task_id"`
	Worker        string `json:"worker"`
	TotalRunTime  int    `json:"total_run_time"`
	Retries       int    `json:"retries"`
	AddDate       string `json:"add_date"`
	UpdateDate    string `json:"update_date"`
}

// MediaPermissionRecord is a row of files_mediapermission. Folded into Media.metadata.
type MediaPermissionRecord struct {
	ID          string `json:"id"`
	MediaID     string `json:"media_id"`
	UserID      string `json:"user_id"`
	OwnerUserID string `json:"owner_user_id"`
	Permission  string `json:"permission"`
	CreatedAt   string `json:"created_at"`
}

type FileRef struct {
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	Size       int64  `json:"size"`
	Checksum   string `json:"checksum"`
	// Kind classifies the ref so phaseMedia can align DB path fields
	// (url/thumbnail/poster/vtt_path/sprite_path) with where phaseFiles lands.
	// Values: original / thumbnail / poster / sprite_vtt / sprite_jpg / subtitle.
	Kind string `json:"kind"`
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
	// M2M / derived entities
	MediaTags(ctx context.Context) (Iterator[*MediaTagRecord], error)
	MediaCategories(ctx context.Context) (Iterator[*MediaCategoryRecord], error)
	PlaylistMedia(ctx context.Context) (Iterator[*PlaylistMediaRecord], error)
	Subtitles(ctx context.Context) (Iterator[*SubtitleRecord], error)
	Subscriptions(ctx context.Context) (Iterator[*SubscriptionRecord], error)
	// Lookup helpers for A-side data with no B-side entity (folded into Media.metadata).
	Licenses(ctx context.Context) (map[string]*LicenseRecord, error)
	RatingsByMedia(ctx context.Context, mediaID string) ([]*RatingRecord, error)
	EncodingsByMedia(ctx context.Context, mediaID string) ([]*EncodingRecord, error)
	MediaPermissionsByMedia(ctx context.Context, mediaID string) ([]*MediaPermissionRecord, error)
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
