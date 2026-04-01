// API 请求库 - 参考 webui 模式
// 支持 Token 刷新、错误处理、请求拦截

import axios from "axios";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:9090";
const API_PREFIX = "/api/v1";
const REQUEST_TIMEOUT = 30000;

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

export const getAccessToken = () => accessToken;

/** Called by useAuth.login() after a successful signin/signup */
export const setAuth = (token: Token) => {
    accessToken = token.access_token;
    localStorage.setItem(TOKEN_KEY, token.access_token);
    localStorage.setItem("token_expires_at", String(Date.now() + token.expires_in * 1000));
};

export const clearAuth = () => {
    accessToken = null;
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    localStorage.removeItem("token_expires_at");
};

export const isTokenExpired = (): boolean => {
    const expiresAt = localStorage.getItem("token_expires_at");
    if (!expiresAt) return false; // no expiry tracked → assume valid
    return Date.now() > parseInt(expiresAt, 10);
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
            if (accessToken && !isTokenExpired()) {
                config.headers.Authorization = `Bearer ${accessToken}`;
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

            // 其余 401：清 token，跳登录
            if (!isRefreshing) {
                isRefreshing = true;
                clearAuth();
                processQueue(error, null);
                isRefreshing = false;
                window.location.href = "/auth/signin";
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

let requestPromise: ReturnType<typeof createRequest> | null = null;
const getRequest = () => {
    if (!requestPromise) {
        requestPromise = createRequest();
    }
    return requestPromise;
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
    const request = await getRequest();

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
