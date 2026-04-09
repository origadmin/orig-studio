// API 请求库 - 参考 webui 模式
// 支持 Token 刷新、错误处理、请求拦截

import axios from "axios";

// Use relative path by default, let rsbuild proxy handle the actual URL
const getApiBaseUrl = (): string => {
    return '';
};

export const API_BASE_URL = getApiBaseUrl();
export const API_PREFIX = "/api/v1";
export const REQUEST_TIMEOUT = 30000;

interface Token {
    access_token: string;
    expires_in: number;
    token_type: string;
    refresh_token?: string; // optional, not issued in M1
}

interface ApiError {
    code: number;
    message: string;
    details?: unknown;
}

// Token 管理 - 与 useAuth.ts 共享同一套 localStorage key
const TOKEN_KEY = "origcms_token";
const USER_KEY = "origcms_user";

let accessToken: string | null = localStorage.getItem(TOKEN_KEY);
let isRefreshing = false;
let failedQueue: Array<{
    resolve: (value: unknown) => void;
    reject: (reason?: unknown) => void;
}> = [];

const processQueue = (error: ApiError | null, token: string | null = null) => {
    failedQueue.forEach((prom) => {
        if (error) {
            prom.reject(error);
        } else {
            prom.resolve(token);
        }
    });
    failedQueue = [];
};

// 确保每次获取token时都从localStorage读取最新值
export const getAccessToken = () => {
    accessToken = localStorage.getItem(TOKEN_KEY);
    return accessToken;
};

// 获取刷新token
export const getRefreshToken = () => {
    return localStorage.getItem('origcms_refresh_token');
};

/** Called by useAuth.login() after a successful signin/signup */
export const setAuth = (token: Token) => {
    accessToken = token.access_token;
    localStorage.setItem(TOKEN_KEY, token.access_token);
    if (token.refresh_token) {
        localStorage.setItem('origcms_refresh_token', token.refresh_token);
    }
    localStorage.setItem("token_expires_at", String(Date.now() + token.expires_in * 1000));
};

export const clearAuth = () => {
    accessToken = null;
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem('origcms_refresh_token');
    localStorage.removeItem(USER_KEY);
    localStorage.removeItem("token_expires_at");
};

export const isTokenExpired = (): boolean => {
    const token = localStorage.getItem(TOKEN_KEY);
    if (!token) return true;

    try {
        // 解析 JWT token 来获取 exp 字段
        const payload = JSON.parse(atob(token.split('.')[1]));
        if (!payload.exp) return true;
        return Date.now() > payload.exp * 1000;
    } catch {
        return true;
    }
};

// 创建 Axios 实例
function createRequest() {
    const request = axios.create({
        baseURL: API_BASE_URL + API_PREFIX,
        timeout: REQUEST_TIMEOUT,
        headers: {
            "Content-Type": "application/json",
        },
    });

    // 请求拦截器：带上 token（如果有）
    request.interceptors.request.use(
        (config) => {
            const token = localStorage.getItem(TOKEN_KEY);
            if (token && !isTokenExpired()) {
                config.headers.Authorization = `Bearer ${token}`;
            }
            return config;
        },
        (error) => Promise.reject(error)
    );

    // 响应拦截器：401 → 清 token，跳登录
    request.interceptors.response.use(
        (response) => response,
        async (error) => {
            const originalRequest = error.config;

            // 非 401 直接拒绝
            if (error.response?.status !== 401) {
                return Promise.reject(error);
            }

            // auth 接口返回 401 不做重定向（登录失败正常错误）
            const authUrls = ["/auth/signin", "/auth/signup"];
            if (authUrls.some((url) => originalRequest.url?.includes(url))) {
                return Promise.reject(error);
            }

            // 检查是否真的是认证失败（token 无效或过期）
            // 避免因为其他原因导致的 401 错误（如权限不足）而清除 token
            const errorMessage = error.response?.data?.message;
            if (errorMessage && (errorMessage.includes('token') || errorMessage.includes('Token') || errorMessage.includes('authentication'))) {
                // 其余 401：清 token，跳登录
                if (!isRefreshing) {
                    isRefreshing = true;
                    clearAuth();
                    processQueue(error, null);
                    isRefreshing = false;
                    window.location.href = "/auth/signin";
                }
            } else {
                // 其他 401 错误（如权限不足），不清除 token
                return Promise.reject(error);
            }

            return new Promise((_, reject) => {
                failedQueue.push({
                    resolve: () => {
                    }, reject
                });
            });
        }
    );

    return request;
}

let requestInstance: ReturnType<typeof createRequest> | null = null;
const getRequest = () => {
    if (!requestInstance) {
        requestInstance = createRequest();
    }
    return requestInstance;
};

// 请求方法
type Method = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";

interface RequestOptions {
    params?: Record<string, unknown>;
    body?: unknown;
    headers?: Record<string, string>;
}

async function fetchApi<T>(
    url: string,
    method: Method = "GET",
    options: RequestOptions = {}
): Promise<T> {
    const request = getRequest();

    // 构建 URL 参数
    const searchParams = new URLSearchParams();
    if (options.params) {
        Object.entries(options.params).forEach(([key, value]) => {
            if (value !== null && value !== undefined && value !== "") {
                if (Array.isArray(value)) {
                    value.forEach((v) => searchParams.append(key, String(v)));
                } else {
                    searchParams.set(key, String(value));
                }
            }
        });
    }

    const finalUrl = searchParams.toString() ? `${url}?${searchParams.toString()}` : url;

    try {
        const response = await request<T>({
            url: finalUrl,
            method,
            data: options.body,
            headers: options.headers,
        });
        return response.data;
    } catch (error: unknown) {
        const axiosError = error as { response?: { data?: ApiError }; message?: string };
        const errorData = axiosError.response?.data;

        if (errorData?.message) {
            throw new Error(errorData.message);
        }
        throw new Error(axiosError.message || "Request failed");
    }
}

export const api = {
    get: <T>(url: string, params?: Record<string, unknown>) => fetchApi<T>(url, "GET", {params}),
    post: <T, B = unknown>(url: string, body?: B) => fetchApi<T>(url, "POST", {body}),
    put: <T, B = unknown>(url: string, body?: B) => fetchApi<T>(url, "PUT", {body}),
    patch: <T, B = unknown>(url: string, body?: B) => fetchApi<T>(url, "PATCH", {body}),
    del: <T>(url: string, params?: Record<string, unknown>) => fetchApi<T>(url, "DELETE", {params}),
};

export type {Token, ApiError};
