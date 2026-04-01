// 认证工具函数
import {api, setAuth, clearAuth, getAccessToken, type Token} from "./request";

// 登录
export async function signIn(credentials: {
    username: string;
    password: string;
    captcha_id?: string;
    captcha_code?: string;
}): Promise<Token> {
    const response = await api.post<Token>("/auth/signin", credentials);
    setAuth(response);
    return response;
}

// 注册
export async function signUp(data: {
    username: string;
    email: string;
    password: string;
}): Promise<Token> {
    const response = await api.post<Token>("/auth/signup", data);
    setAuth(response);
    return response;
}

// 登出
export async function signOut() {
    try {
        await api.post("/auth/signout", {});
    } finally {
        clearAuth();
    }
}

// 刷新 Token
export async function refreshToken(): Promise<Token> {
    const response = await api.post<Token>("/auth/refresh", {});
    setAuth(response);
    return response;
}

// 获取当前用户信息
export interface CurrentUser {
    id: string;
    username: string;
    email: string;
    avatar?: string;
    role: string;
    status: string;
    created_at: string;
}

export async function getCurrentUser(): Promise<CurrentUser> {
    return api.get<CurrentUser>("/auth/me");
}

// 检查是否已登录
export function isAuthenticated(): boolean {
    return !!getAccessToken();
}

// 导出 request 中的工具
export {getAccessToken, setAuth, clearAuth};
