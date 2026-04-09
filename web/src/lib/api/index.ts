// API 模块导出
// 对应后端 /api/v1/* 路径

// ==================== Core ====================
export {api, getAccessToken, setAuth, clearAuth} from "./request";
export type {Token, ApiError} from "./request";

// ==================== Auth ====================
export {signIn, signUp, signOut, refreshToken, getCurrentUser, isAuthenticated} from "./auth";
export type {CurrentUser} from "./auth";

// ==================== Users ====================
export {userApi} from "./user";
export type {User, UserListResponse, CreateUserRequest, UpdateUserRequest} from "./user";

// ==================== Media ====================
export {mediaApi, encodingApi, legacyMediaApi} from "./media";
export type {
    Media,
    MediaListResponse,
    CreateMediaRequest,
    UpdateMediaRequest,
    EncodeProfile,
    EncodingTask,
    TranscodingStatusResponse,
    MediaVariantSummary,
} from "./media";

// ==================== Content ====================
export {contentApi} from "./content";
export type {Content, ContentListResponse, CreateContentRequest, UpdateContentRequest} from "./content";

// ==================== Categories ====================
export {categoryApi} from "./category";
export type {Category} from "./category";

// ==================== Tags ====================
export {tagApi} from "./tag";
export type {Tag} from "./tag";

// ==================== Comments ====================
export {commentApi} from "./comment";
export type {Comment} from "./comment";

// ==================== Playlists ====================
export {playlistApi} from "./playlist";
export type {Playlist, PlaylistListResponse, CreatePlaylistRequest, UpdatePlaylistRequest} from "./playlist";

// ==================== Search ====================
export {searchApi} from "./search";
export type {SearchResponse} from "./search";

// ==================== Interactions (New Module) ====================
export {
    interactionApi,
    likeApi,
    favoriteApi,
    subscriptionApi,
    shareApi,
} from "./interaction";
export type {
    LikeResponse,
    LikeStatusBatchResponse,
    Favorite,
    ToggleFavoriteResponse,
    FavoriteListResponse,
    SubscriptionStatus,
    SubscriptionListResponse,
    ShareResponse,
} from "./interaction";

// ==================== System (New Module) ====================
export {systemApi, statsApi, settingsApi} from "./system";
export type {
    DashboardStats,
    MediaStats,
    UserStats,
    TrafficStatsItem,
    TrafficStatsResponse,
    SystemSettings,
    UpdateSettingsRequest,
} from "./system";

// ==================== Legacy Compatibility ====================
// 为了保持向后兼容，保留旧的导出
export {statsApi as legacyStatsApi} from "./stats";
export type {Stats} from "./stats";
