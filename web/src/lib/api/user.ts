// API 客户端 - 用户模块
import {api} from "../request";

export interface User {
    id: string;
    username: string;
    email: string;
    avatar?: string;
    role: string;
    status: string;
    created_at: string;
    updated_at?: string;
}

export interface UserListResponse {
    list: User[];
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
    email?: string;
    avatar?: string;
    role?: string;
    status?: string;
}

export const userApi = {
    // 获取用户列表
    list: (params?: { page?: number; page_size?: number; keyword?: string; status?: string }) =>
        api.get<UserListResponse>("/users", params),

    // 获取用户详情
    get: (id: string) => api.get<User>(`/users/${id}`),

    // 创建用户
    create: (data: CreateUserRequest) => api.post<User>("/users", data),

    // 更新用户
    update: (id: string, data: UpdateUserRequest) => api.put<User>(`/users/${id}`, data),

    // 删除用户
    delete: (id: string) => api.del<void>(`/users/${id}`),

    // 更新用户状态
    updateStatus: (id: string, status: string) =>
        api.patch<User>(`/users/${id}/status`, {status}),
};
