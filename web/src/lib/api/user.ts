// API 客户端 - 用户模块
import {api} from "../request";

export interface User {
    id: string;
    username: string;
    nickname?: string;
    email: string;
    avatar?: string;
    phone?: string;
    role: string;
    status: number | string; // Protojson serializes enum as string name (e.g. "USER_STATUS_ACTIVE") by default
    create_time: string;
    update_time?: string;
}

// Map proto UserStatus numeric enum values to frontend string labels
export const USER_STATUS_MAP: Record<number, string> = {
    0: 'unspecified',
    1: 'pending',
    2: 'active',
    3: 'inactive',
    4: 'suspended',
    5: 'rejected',
};

// Map protojson enum string names to frontend string labels
// protojson (without UseEnumNumbers) serializes enums as their proto name, e.g. "USER_STATUS_ACTIVE"
export const PROTO_STATUS_MAP: Record<string, string> = {
    'USER_STATUS_UNSPECIFIED': 'unspecified',
    'USER_STATUS_PENDING': 'pending',
    'USER_STATUS_ACTIVE': 'active',
    'USER_STATUS_INACTIVE': 'inactive',
    'USER_STATUS_SUSPENDED': 'suspended',
    'USER_STATUS_REJECTED': 'rejected',
};

// Map frontend string labels to proto UserStatus enum values
export const USER_STATUS_REVERSE_MAP: Record<string, number> = {
    'unspecified': 0,
    'pending': 1,
    'active': 2,
    'inactive': 3,
    'suspended': 4,
    'rejected': 5,
};

// Helper to normalize status to string label for display
// Handles three input formats:
// 1. Protojson enum name string: "USER_STATUS_ACTIVE" -> "active"
// 2. Numeric enum value: 2 -> "active"
// 3. Already a display label: "active" -> "active"
export function getUserStatusLabel(status: number | string): string {
    if (typeof status === 'string') {
        // Check if it's a proto enum name like "USER_STATUS_ACTIVE"
        if (PROTO_STATUS_MAP[status]) {
            return PROTO_STATUS_MAP[status];
        }
        // Already a display label like "active"
        return status;
    }
    return USER_STATUS_MAP[status] || 'unknown';
}

// Helper to normalize status to numeric value for API calls
export function getUserStatusCode(status: number | string): number {
    if (typeof status === 'number') return status;
    // Check if it's a proto enum name like "USER_STATUS_ACTIVE"
    if (PROTO_STATUS_MAP[status]) {
        return USER_STATUS_REVERSE_MAP[PROTO_STATUS_MAP[status]] ?? 2;
    }
    return USER_STATUS_REVERSE_MAP[status] ?? 2; // default to active
}

export interface UserListResponse {
    items: User[];
    total: number;
    page: number;
    page_size: number;
}

export interface CreateUserRequest {
    username: string;
    email: string;
    password: string;
    role?: string;
}

export interface UpdateUserRequest {
    username?: string;
    nickname?: string;
    email?: string;
    avatar?: string;
    phone?: string;
    role?: string;
    status?: string;
}

export interface AdminCreateUserRequest {
    username: string;
    email: string;
    password?: string;
    nickname?: string;
    role?: string;
}

export interface UpdateProfileRequest {
    nickname?: string;
    email?: string;
    avatar?: string;
    bio?: string;
}

export interface ChangePasswordRequest {
    old_password: string;
    new_password: string;
}

export interface SubscriptionStatusResponse {
    is_subscribed: boolean;
    subscriber_count: number;
}

export interface SubscriptionListResponse {
    items: User[];
    total: number;
    page: number;
    page_size: number;
}

export const userApi = {
    // 获取用户列表
    list: (params?: { page?: number; page_size?: number; keyword?: string; status?: string; role?: string }) =>
        api.get<UserListResponse>("/users", params),

    // 获取用户详情
    get: (id: string) => api.get<User>(`/users/${id}`),

    // 通过 username 获取用户
    getByUsername: (username: string) => api.get<User>(`/users/username/${username}`),

    // 创建用户
    create: (data: CreateUserRequest) => api.post<User>("/users", data),

    // 更新用户
    update: (id: string, data: UpdateUserRequest) => api.put<User>(`/users/${id}`, data),

    // 删除用户
    delete: (id: string) => api.del<void>(`/users/${id}`),

    // 更新用户状态
    updateStatus: (id: string, status: string) =>
        api.patch<User>(`/users/${id}/status`, {status}),

    // ==================== 当前用户 APIs (使用 /me 路径) ====================

    // 获取当前用户信息 - 使用 /me 路径
    getMe: () => api.get<User>("/me"),

    // 更新当前用户信息 - 使用 /me 路径
    updateMe: (data: UpdateProfileRequest) => api.put<User>("/me", data),

    // 修改密码 - 使用 /me/password 路径 (后端为 PUT 方法)
    changePassword: (data: ChangePasswordRequest) =>
        api.put<void>("/me/password", data),

    // ==================== Subscription APIs ====================

    // 获取订阅状态
    getSubscriptionStatus: (userId: string) =>
        api.get<SubscriptionStatusResponse>(`/users/${userId}/subscription`),

    // 订阅用户/频道
    subscribe: (userId: string) =>
        api.post<{ success: boolean }>(`/users/${userId}/subscribe`),

    // 取消订阅
    unsubscribe: (userId: string) =>
        api.del<{ success: boolean }>(`/users/${userId}/subscribe`),

    // 获取我的订阅列表
    getSubscriptions: (params?: { page?: number; page_size?: number }) =>
        api.get<SubscriptionListResponse>("/me/subscriptions", params),

    // Get my followers list
    getFollowers: (params?: { page?: number; page_size?: number }) =>
        api.get<SubscriptionListResponse>("/me/followers", params),
};

// ==================== Admin User API (UUID based, requires JWT + Admin) ====================
export const adminUserApi = {
    // Create user (Admin)
    create: (data: AdminCreateUserRequest) =>
        api.post<User>("/admin/users", data),

    // List all users (Admin)
    list: (params?: { page?: number; page_size?: number; keyword?: string; status?: string; role?: string }) =>
        api.get<UserListResponse>("/admin/users", params),

    // Get user detail by ID (Admin)
    get: (id: string) =>
        api.get<User>(`/admin/users/${id}`),

    // Update user by ID (Admin)
    update: (id: string, data: UpdateUserRequest) =>
        api.put<User>(`/admin/users/${id}`, data),

    // Delete user by ID (Admin)
    delete: (id: string) =>
        api.del<void>(`/admin/users/${id}`),

    // Update user status (Admin) - sends numeric status code
    updateStatus: (id: string, status: number | string) =>
        api.patch<void>(`/admin/users/${id}/status`, {status: typeof status === 'number' ? status : getUserStatusCode(status)}),

    // Update user role (Admin)
    updateRole: (id: string, role: string) =>
        api.patch<void>(`/admin/users/${id}/role`, {role}),
};

export default userApi;
