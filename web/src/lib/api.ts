// API client - Main export file
// Re-exports from the API modules

export {api, getAccessToken, setAuth, clearAuth} from "./request";
export {isAuthenticated} from "./auth";
export {signIn, signUp, signOut, refreshToken, getCurrentUser, isAuthenticated as checkAuth} from "./auth";

export type {Token, ApiError} from "./request";
export type {CurrentUser} from "./auth";

export {userApi} from "./api/user";
export type {User, UserListResponse, CreateUserRequest, UpdateUserRequest} from "./api/user";

export {mediaApi} from "./api/media";
export type {Media, MediaListResponse, CreateMediaRequest, UpdateMediaRequest} from "./api/media";

export {contentApi} from "./api/content";
export type {Content, ContentListResponse, CreateContentRequest, UpdateContentRequest} from "./api/content";

export {categoryApi} from "./api/category";
export type {Category} from "./api/category";

export {tagApi} from "./api/tag";
export type {Tag} from "./api/tag";

export {commentApi} from "./api/comment";
export type {Comment} from "./api/comment";

export {likeApi} from "./api/like";
export type {LikeResponse, ToggleLikeResponse} from "./api/like";

export {favoriteApi} from "./api/favorite";
export type {Favorite, ToggleFavoriteResponse} from "./api/favorite";

export {playlistApi} from "./api/playlist";
export type {Playlist, PlaylistDetail} from "./api/playlist";

export {searchApi} from "./api/search";
export type {SearchResponse} from "./api/search";

// Stats API
export const statsApi = {
    get: () => api.get<{ users: number; media: number; content: number; storage: string; views: number }>("/stats"),
};
