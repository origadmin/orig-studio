import {api} from "../request";

export interface PermissionGroup {
    id: string;
    name: string;
    description?: string;
    permissions: string[];
    category_scope?: string[];
    is_active: boolean;
    member_count?: number;
    created_at: string;
    updated_at: string;
}

export interface PermissionGroupListResponse {
    items: PermissionGroup[];
    total: number;
    page: number;
    page_size: number;
}

export interface CreatePermissionGroupRequest {
    name: string;
    description?: string;
    permissions: string[];
    category_scope?: string[];
}

export interface UpdatePermissionGroupRequest {
    name?: string;
    description?: string;
    permissions?: string[];
    category_scope?: string[];
}

export interface GroupMember {
    id: string;
    user_id: string;
    username: string;
    group_id: string;
    joined_at: string;
}

export interface GroupMemberListResponse {
    items: GroupMember[];
    total: number;
    page: number;
    page_size: number;
}

export interface AddMembersRequest {
    user_ids: string[];
}

export interface AddMembersResponse {
    added: number;
    skipped: number;
}

export interface UserPermissionsResponse {
    user_id: string;
    role: string;
    effective_permissions: Record<string, {
        sources: string[];
        scope?: string[];
    }>;
    groups: {
        id: string;
        name: string;
        is_active: boolean;
        joined_at: string;
    }[];
}

export interface PermissionItem {
    key: string;
    label: string;
    resource_type: string;
    action: string;
}

export interface PermissionEnumResponse {
    permissions: PermissionItem[];
    role_defaults: Record<string, string[]>;
}

export const permissionApi = {
    getEnum: () =>
        api.get<PermissionEnumResponse>('/permissions'),
};

export const adminPermissionApi = {
    list: (params?: { page?: number; page_size?: number; is_active?: boolean }) =>
        api.get<PermissionGroupListResponse>('/admin/permission-groups', params),

    get: (id: string) =>
        api.get<PermissionGroup>(`/admin/permission-groups/${id}`),

    create: (data: CreatePermissionGroupRequest) =>
        api.post<PermissionGroup>('/admin/permission-groups', data),

    update: (id: string, data: UpdatePermissionGroupRequest) =>
        api.put<PermissionGroup>(`/admin/permission-groups/${id}`, data),

    delete: (id: string) =>
        api.del<void>(`/admin/permission-groups/${id}`),

    toggle: (id: string, data: { is_active: boolean }) =>
        api.post<{ id: string; is_active: boolean }>(`/admin/permission-groups/${id}/toggle`, data),

    getMembers: (id: string, params?: { page?: number; page_size?: number }) =>
        api.get<GroupMemberListResponse>(`/admin/permission-groups/${id}/members`, params),

    addMembers: (id: string, data: AddMembersRequest) =>
        api.post<AddMembersResponse>(`/admin/permission-groups/${id}/members`, data),

    removeMember: (groupId: string, userId: string) =>
        api.del<void>(`/admin/permission-groups/${groupId}/members/${userId}`),

    getUserPermissions: (userId: string) =>
        api.get<UserPermissionsResponse>(`/admin/users/${userId}/permissions`),
};
