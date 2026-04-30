// API 模块导出
// 对应后端 /api/v1/* 路径

// ==================== Core ====================
export {api, getAccessToken, setAuth, clearAuth} from "../request";
export type {Token, ApiError, ApiResponse} from "../request";

// ==================== Auth ====================
export {signIn, signUp, signOut, refreshToken, getCurrentUser, isAuthenticated, authApi} from "./auth";
export type {CurrentUser} from "./auth";

// ==================== Users ====================
export {userApi, adminUserApi} from "./user";
export type {
    User,
    UserListResponse,
    CreateUserRequest,
    UpdateUserRequest,
    UpdateProfileRequest,
    ChangePasswordRequest,
    SubscriptionStatusResponse,
    SubscriptionListResponse
} from "./user";

// ==================== Media ====================
export {mediaApi, encodingApi, legacyMediaApi, normalizeMedia, normalizeMediaList} from "./media";
export type {
    Media,
    MediaListResponse,
    CreateMediaRequest,
    UpdateMediaRequest,
    EncodeProfile,
    EncodingTask,
    TranscodingStatusResponse,
    MediaVariantSummary,
    LikeResponse,
    FavoriteResponse,
    ShareResponse,
} from "./media";

// ==================== Articles ====================
export {articleApi, adminArticleApi} from "./article";
export type {Article, ArticleListResponse, CreateArticleRequest, UpdateArticleRequest} from "./article";

// ==================== Categories ====================
export {categoryApi, adminCategoryApi} from "./category";
export type {Category} from "./category";

// ==================== Tags ====================
export {tagApi} from "./tag";
export type {Tag} from "./tag";

// ==================== Comments ====================
export {commentApi} from "./comment";
export type {Comment} from "./comment";

// ==================== Playlists ====================
export {playlistApi, adminPlaylistApi} from "./playlist";
export type {Playlist, PlaylistListResponse, CreatePlaylistRequest, UpdatePlaylistRequest} from "./playlist";

// ==================== Search ====================
export {searchApi} from "./search";
export type {SearchResponse} from "./search";

// ==================== Notifications ====================
export {notificationApi} from "./notification";
export type {Notification, NotificationListResponse} from "./notification";

// ==================== Interactions (Legacy - use mediaApi instead) ====================
export {likeApi} from "./like";
export {favoriteApi} from "./favorite";
export {shareApi} from "./share";
export {subscriptionApi} from "./subscription";
export type {
    SubscriptionStatus,
    SubscriptionListResponse as SubscriptionList,
} from "./subscription";

// ==================== Stats ====================
export {statsApi} from "./stats";
export type {
    DashboardStats,
    MediaStats,
    UserStats,
} from "./stats";

// ==================== System ====================
export {systemApi, settingsApi} from "./system";
export type {
    GroupedSettings,
    SettingItem,
    UpdateSettingsRequest,
    UpdateSettingItem,
} from "./system";

// ==================== History ====================
export {historyApi} from "./history";
export type {HistoryItem, HistoryListResponse} from "./history";

// ==================== Explore ====================
export {exploreApi} from "./explore";
export type {TrendingItem, TrendingResponse} from "./explore";
