// API 客户端 - 文章模块
import {api} from "../request";

export interface Article {
    id: string;
    title: string;
    slug: string;
    content: string;
    summary?: string;
    state: string;
    user_id: string;
    category_id?: number;
    tags?: string[];
    view_count: number;
    comment_count: number;
    featured: boolean;
    published_at?: string;
    created_at: string;
    updated_at?: string;
}

export interface ArticleListResponse {
    items: Article[];
    total: number;
    page: number;
    page_size: number;
}

export interface CreateArticleRequest {
    title: string;
    slug?: string;
    content: string;
    summary?: string;
    state?: string;
    category_id?: number;
    tags?: string[];
    featured?: boolean;
    published_at?: string;
}

export interface UpdateArticleRequest {
    title?: string;
    slug?: string;
    content?: string;
    summary?: string;
    state?: string;
    category_id?: number;
    tags?: string[];
    featured?: boolean;
    published_at?: string;
}

export const articleApi = {
    // Get article list (public, only published)
    list: (params?: {
        page?: number;
        page_size?: number;
        category_id?: number;
        keyword?: string;
        state?: string;
    }) => api.get<ArticleListResponse>("/articles", {...params, state: params?.state || "published"}),

    // Get article detail (public)
    get: (id: string) => api.get<Article>(`/articles/${id}`),

    // Get article by slug
    getBySlug: (slug: string) => api.get<Article>(`/articles/slug/${slug}`),

    // Get featured articles
    featured: (limit?: number) => api.get<Article[]>("/articles/featured", {limit}),

    // Get latest articles
    latest: (limit?: number) => api.get<Article[]>("/articles/latest", {limit}),
};

// ==================== Admin Article API (requires JWT + Admin) ====================
export const adminArticleApi = {
    // List all articles including unpublished (Admin)
    adminList: (params?: {
        page?: number;
        page_size?: number;
        state?: string;
        keyword?: string;
        category_id?: number;
    }) => api.get<ArticleListResponse>("/admin/articles", params),

    // Get article detail (Admin)
    get: (id: string) => api.get<Article>(`/admin/articles/${id}`),

    // Create article (Admin)
    create: (data: CreateArticleRequest) => api.post<Article>("/admin/articles", data),

    // Update article (Admin)
    update: (id: string, data: UpdateArticleRequest) => api.put<Article>(`/admin/articles/${id}`, data),

    // Delete article (Admin)
    delete: (id: string) => api.del<void>(`/admin/articles/${id}`),

    // Update article state (Admin)
    updateState: (id: string, state: string) =>
        api.patch<void>(`/admin/articles/${id}/state`, {state}),
};
