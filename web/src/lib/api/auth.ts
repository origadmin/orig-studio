// API 客户端 - 认证模块
import {api, setAuth, clearAuth, getAccessToken, getRefreshToken, isTokenExpired, type Token} from "../request";

export interface CurrentUser {
    id: string;
    username: string;
    email: string;
    nickname?: string;
    avatar?: string;
    role: string;
    status: string;
    created_at: string;
}

interface SignInResponse extends Token {
    user: CurrentUser;
}

// 登录
export const signIn = async (username: string, password: string): Promise<SignInResponse> => {
    const response = await api.post<SignInResponse>("/auth/signin", {
        username,
        password,
    });
    setAuth(response);
    return response;
};

// 注册
export const signUp = async (
    username: string,
    email: string,
    password: string,
    nickname?: string
): Promise<CurrentUser> => {
    return api.post<CurrentUser>("/auth/signup", {
        username,
        email,
        password,
        nickname,
    });
};

// 登出
export const signOut = () => {
    clearAuth();
};

// 刷新 Token
export const refreshToken = async (): Promise<Token> => {
    const refresh_token = getRefreshToken();
    if (!refresh_token) {
        throw new Error("No refresh token");
    }
    const response = await api.post<Token>("/auth/refresh", {refresh_token});
    setAuth(response);
    return response;
};

// 获取当前用户 - 使用 /me 路径
export const getCurrentUser = async (): Promise<CurrentUser> => {
    const token = getAccessToken();
    if (!token || isTokenExpired()) {
        throw new Error("Not authenticated");
    }
    return api.get<CurrentUser>("/me");
};

// API 对象（用于测试兼容性）
export const authApi = {
    login: signIn,
    register: signUp,
    logout: signOut,
    getCurrentUser,
};

// 检查是否已认证
export const isAuthenticated = (): boolean => {
    const token = getAccessToken();
    return !!token && !isTokenExpired();
};
