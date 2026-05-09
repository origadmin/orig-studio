import {api} from "../request";

export interface NavItem {
    id: string;
    type: "internal_link" | "external_link" | "category";
    label: string;
    label_i18n?: Record<string, string>;
    url: string;
    icon?: string;
    color?: string;
    sequence: number;
    is_visible?: boolean;
    open_new_tab: boolean;
}

export interface NavItemListResponse {
    items: NavItem[];
    total: number;
}

export interface CreateNavItemRequest {
    type: "internal_link" | "external_link" | "category";
    label: string;
    url: string;
    icon?: string;
    sequence?: number;
    open_new_tab?: boolean;
}

export interface UpdateNavItemRequest {
    type?: "internal_link" | "external_link" | "category";
    label?: string;
    url?: string;
    icon?: string;
    sequence?: number;
    open_new_tab?: boolean;
}

export interface Banner {
    id: string;
    title: string;
    title_i18n?: Record<string, string>;
    subtitle?: string;
    subtitle_i18n?: Record<string, string>;
    badge_text?: string;
    image_url?: string;
    image_mobile_url?: string;
    primary_btn_text?: string;
    primary_btn_url?: string;
    secondary_btn_text?: string;
    secondary_btn_url?: string;
    is_active: boolean;
    sequence: number;
    start_time?: string;
    end_time?: string;
    create_time: string;
    update_time: string;
}

export interface BannerListResponse {
    items: Banner[];
    total: number;
}

export interface CreateBannerRequest {
    title: string;
    subtitle?: string;
    badge_text?: string;
    image_url?: string;
    primary_btn_text?: string;
    primary_btn_url?: string;
    secondary_btn_text?: string;
    secondary_btn_url?: string;
    sequence?: number;
    start_time?: string;
    end_time?: string;
}

export interface UpdateBannerRequest {
    title?: string;
    subtitle?: string;
    badge_text?: string;
    image_url?: string;
    primary_btn_text?: string;
    primary_btn_url?: string;
    secondary_btn_text?: string;
    secondary_btn_url?: string;
    sequence?: number;
    start_time?: string;
    end_time?: string;
}

export interface CustomPage {
    id: string;
    title: string;
    slug: string;
    type: "static" | "markdown" | "rich_text";
    content_format: "markdown" | "html" | "plain";
    content: string;
    layout: "default" | "full_width" | "sidebar";
    is_published: boolean;
    published_at?: string;
    seo_title?: string;
    seo_description?: string;
    featured_image?: string;
    view_count: number;
    create_time: string;
    update_time: string;
}

export interface CustomPageListResponse {
    items: CustomPage[];
    total: number;
}

export interface CreateCustomPageRequest {
    title: string;
    slug: string;
    type?: "static" | "markdown" | "rich_text";
    content_format?: "markdown" | "html" | "plain";
    content: string;
    layout?: "default" | "full_width" | "sidebar";
    is_published?: boolean;
    seo_title?: string;
    seo_description?: string;
    featured_image?: string;
}

export interface UpdateCustomPageRequest {
    title?: string;
    slug?: string;
    type?: "static" | "markdown" | "rich_text";
    content_format?: "markdown" | "html" | "plain";
    content?: string;
    layout?: "default" | "full_width" | "sidebar";
    is_published?: boolean;
    seo_title?: string;
    seo_description?: string;
    featured_image?: string;
}

export interface FeaturedUser {
    id: string;
    username: string;
    avatar?: string;
    subscriber_count: number;
}

export interface PortalConfig {
    navigation: {
        items: NavItem[];
        visible_count: number;
    };
    banners: Banner[];
    featured_users: FeaturedUser[];
    site: {
        name: string;
        default_lang: string;
    };
}

export interface ModulePortalConfig {
    modules: {
        articles: boolean;
        videos: boolean;
        music: boolean;
    };
    layout: 'video' | 'article' | 'mixed' | 'welcome' | 'doc';
    site: {
        site_name: string;
        site_description: string;
        allow_registration: boolean;
        allow_upload: boolean;
    };
}

export const portalApi = {
    getConfig: () =>
        api.get<PortalConfig>('/portal/config'),

    getModuleConfig: () =>
        api.get<ModulePortalConfig>('/portal/config'),
};

export const adminPortalApi = {
    listNavItems: () =>
        api.get<NavItemListResponse>('/admin/nav-items'),

    createNavItem: (data: CreateNavItemRequest) =>
        api.post<NavItem>('/admin/nav-items', data),

    updateNavItem: (id: string, data: UpdateNavItemRequest) =>
        api.put<NavItem>(`/admin/nav-items/${id}`, data),

    deleteNavItem: (id: string) =>
        api.del<void>(`/admin/nav-items/${id}`),

    reorderNavItems: (data: { ids: string[] }) =>
        api.put<void>('/admin/nav-items/reorder', data),

    listBanners: () =>
        api.get<BannerListResponse>('/admin/banners'),

    createBanner: (data: CreateBannerRequest) =>
        api.post<Banner>('/admin/banners', data),

    updateBanner: (id: string, data: UpdateBannerRequest) =>
        api.put<Banner>(`/admin/banners/${id}`, data),

    toggleBanner: (id: string) =>
        api.post<Banner>(`/admin/banners/${id}/toggle`),

    listPages: () =>
        api.get<CustomPageListResponse>('/admin/pages'),

    createPage: (data: CreateCustomPageRequest) =>
        api.post<CustomPage>('/admin/pages', data),

    updatePage: (id: string, data: UpdateCustomPageRequest) =>
        api.put<CustomPage>(`/admin/pages/${id}`, data),

    deletePage: (id: string) =>
        api.del<void>(`/admin/pages/${id}`),

    getPage: (slug: string) =>
        api.get<CustomPage>(`/p/${slug}`),
};
