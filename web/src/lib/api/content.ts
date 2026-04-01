// API 客户端 - 内容模块
import {api} from "../request";

export interface Content {
    id: string;
    title: string;
    slug: string;
    content: string;
    excerpt?: string;
    type: "article" | "page" | "post";
    status: string;
    author_id: string;
    category_id?: string;
    tags?: string[];
    featured_image?: string;
    views: number;
    published_at?: string;
    created_at: string;
    updated_at?: string;
}

export interface ContentListResponse {
    list: Content[];
    total: number;
    page: number;
    page_size: number;
}

export interface CreateContentRequest {
    title: string;
    slug?: string;
    content: string;
    excerpt?: string;
    type: Content["type"];
    category_id?: string;
    tags?: string[];
    featured_image?: string;
    published_at?: string;
}

export interface UpdateContentRequest {
    title?: string;
    slug?: string;
    content?: string;
    excerpt?: string;
    category_id?: string;
    tags?: string[];
    featured_image?: string;
    status?: string;
    published_at?: string;
}

export const contentApi = {
    // 获取内容列表（公开）
    list: (params?: {
        page?: number;
        page_size?: number;
        type?: string;
        category_id?: string;
        keyword?: string;
        status?: string;
    }) => api.get<ContentListResponse>("/content", {...params, status: params?.status || "published"}),

    // 获取内容详情（公开）
    get: (id: string) => api.get<Content>(`/content/${id}`),

    // 根据 slug 获取内容
    getBySlug: (slug: string) => api.get<Content>(`/content/slug/${slug}`),

    // 获取推荐内容
    featured: (limit?: number) => api.get<Content[]>("/content/featured", {limit}),

    // 获取最新内容
    latest: (limit?: number) => api.get<Content[]>("/content/latest", {limit}),

    // 管理端：获取所有内容（包括未发布的）
    adminList: (params?: {
        page?: number;
        page_size?: number;
        type?: string;
        status?: string;
        keyword?: string;
    }) => api.get<ContentListResponse>("/content", params),

    // 管理端：创建内容
    create: (data: CreateContentRequest) => api.post<Content>("/content", data),

    // 管理端：更新内容
    update: (id: string, data: UpdateContentRequest) => api.put<Content>(`/content/${id}`, data),

    // 管理端：删除内容
    delete: (id: string) => api.del<void>(`/content/${id}`),

    // 管理端：更新内容状态
    updateStatus: (id: string, status: string) =>
        api.patch<Content>(`/content/${id}/status`, {status}),
};
