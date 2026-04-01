// API 模块导出
export {api, getAccessToken, setAuth, clearAuth} from "./request";
export {signIn, signUp, signOut, refreshToken, getCurrentUser, isAuthenticated} from "./auth";
export {statsApi} from "./stats";
export type {Stats} from "./stats";

export type {Token, ApiError} from "./request";
export type {CurrentUser} from "./auth";

export {userApi} from "./user";
export type {User, UserListResponse, CreateUserRequest, UpdateUserRequest} from "./user";

export {mediaApi} from "./media";
export type {Media, MediaListResponse, CreateMediaRequest, UpdateMediaRequest} from "./media";

export {contentApi} from "./content";
export type {Content, ContentListResponse, CreateContentRequest, UpdateContentRequest} from "./content";

export {categoryApi} from "./category";
export type {Category} from "./category";

export {tagApi} from "./tag";
export type {Tag} from "./tag";

export {commentApi} from "./comment";
export type {Comment} from "./comment";

export {likeApi} from "./like";
export type {LikeResponse, ToggleLikeResponse} from "./like";

export {favoriteApi} from "./favorite";
export type {Favorite, ToggleFavoriteResponse} from "./favorite";

export {playlistApi} from "./playlist";
export type {Playlist, PlaylistDetail} from "./playlist";

export {searchApi} from "./search";
export type {SearchResponse} from "./search";
