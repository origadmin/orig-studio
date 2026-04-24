// API 请求库 - 按照 webui 项目模式重写

import axios from "axios";

// Use relative path by default, let rsbuild proxy handle the actual URL
const getApiBaseUrl = (): string => {
    return '';
};

export const API_BASE_URL = getApiBaseUrl();
export const API_PREFIX = "/api/v1";
export const REQUEST_TIMEOUT = 30000;

// 统一响应格式接口
export interface ApiResponse<T> {
    code: number;
    message: string;
    data: T;
}

interface Token {
    access_token: string;
    expires_in: number;
    token_type: string;
    refresh_token?: string;
    user?: {
        id: string;
        username: string;
        nickname?: string;
        email?: string;
        is_staff: boolean;
    };
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
    // 确保 expires_in 是数字
    const expiresIn = typeof token.expires_in === 'string' 
        ? parseInt(token.expires_in, 10) 
        : token.expires_in;
    localStorage.setItem("token_expires_at", String(Date.now() + expiresIn * 1000));
    
    // 如果响应中包含 user 信息，保存它
    if (token.user) {
        const user = {
            id: token.user.id,
            username: token.user.username,
            displayName: token.user.nickname || token.user.username,
            avatarUrl: undefined,
            roles: token.user.is_staff ? ['admin', 'user'] : ['user']
        };
        localStorage.setItem(USER_KEY, JSON.stringify(user));
        
        // 触发事件通知 useAuth 更新
        window.dispatchEvent(new StorageEvent('storage', {
            key: USER_KEY,
            newValue: JSON.stringify(user),
        }));
    }
    
    // 触发事件通知 useAuth 更新
    window.dispatchEvent(new StorageEvent('storage', {
        key: TOKEN_KEY,
        newValue: token.access_token,
    }));
};

export const clearAuth = () => {
    accessToken = null;
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem('origcms_refresh_token');
    localStorage.removeItem(USER_KEY);
    localStorage.removeItem("token_expires_at");
};

export const isTokenExpired = (bufferSeconds: number = 60): boolean => {
    const token = localStorage.getItem(TOKEN_KEY);
    if (!token) return true;

    try {
        // 解析 JWT token 来获取 exp 字段
        const payload = JSON.parse(atob(token.split('.')[1]));
        if (!payload.exp) return true;
        // 提前 bufferSeconds 认为过期，避免边界情况
        return Date.now() > (payload.exp - bufferSeconds) * 1000;
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

    // Request Interceptor
    request.interceptors.request.use(
        (config) => {
            const token = localStorage.getItem(TOKEN_KEY);
            if (token && config.headers) {
                config.headers.Authorization = `Bearer ${token}`;
            }
            return config;
        },
        (error) => Promise.reject(error)
    );

    // 响应拦截器：处理 401 和 token 刷新 - 按照 webui 项目模式
    let isRefreshing = false;
    let failedQueue: { resolve: (value: unknown) => void; reject: (reason?: unknown) => void }[] = [];

    const processQueue = (error: unknown | null, token: string | null = null) => {
        failedQueue.forEach((prom) => {
            if (error) {
                prom.reject(error);
            } else {
                prom.resolve(token);
            }
        });
        failedQueue = [];
    };

    const handleAuthError = () => {
        clearAuth();
        window.location.href = "/auth/signin";
    };

    request.interceptors.response.use(
        (response) => {
            // 适配新的统一响应格式 {code, message, data}
            // 如果响应包含 code 和 data 字段，返回 data 部分
            const data = response.data;
            if (data && typeof data === 'object' && 'code' in data && 'data' in data) {
                // 检查是否成功响应 (code === 0)
                if (data.code !== 0) {
                    // 业务错误，抛出异常
                    return Promise.reject({
                        response: {
                            data: {
                                code: data.code,
                                message: data.message || 'Request failed',
                            }
                        }
                    });
                }
                // 返回 data 部分
                return {...response, data: data.data};
            }
            // 原始格式，直接返回（用于认证接口）
            return response;
        },
        async (error) => {
            const originalRequest = error.config as any;

            // 构造完整的 URL
            const refreshTokenUrl = API_PREFIX + "/auth/refresh";
            const signinUrl = "/auth/signin";
            const signupUrl = "/auth/signup";
            
            const publicUrls = [refreshTokenUrl, signinUrl, signupUrl];
            
            // 如果不是 401 或者是公共接口的 401，直接拒绝
            if (error.response?.status !== 401 || publicUrls.includes(originalRequest.url || "")) {
                return Promise.reject(error);
            }

            if (isRefreshing) {
                return new Promise((resolve, reject) => {
                    failedQueue.push({ resolve, reject });
                })
                    .then((token) => {
                        if (originalRequest.headers) {
                            originalRequest.headers.Authorization = `Bearer ${token}`;
                        }
                        return request(originalRequest);
                    })
                    .catch((err) => {
                        return Promise.reject(err);
                    });
            }

            isRefreshing = true;
            originalRequest._retry = true;

            const refreshToken = getRefreshToken();
            if (!refreshToken) {
                isRefreshing = false;
                handleAuthError();
                return Promise.reject(error);
            }

            try {
                // 使用普通 axios 而不是 request 实例调用 refresh token 接口
                // 这一点很重要！webui 项目也是这样做的！
                const { data: newToken } = await axios.post<Token>(
                    (API_BASE_URL || "") + refreshTokenUrl, 
                    { refresh_token: refreshToken }
                );

                setAuth(newToken);
                if (originalRequest.headers) {
                    originalRequest.headers.Authorization = `Bearer ${newToken.access_token}`;
                }
                processQueue(null, newToken.access_token);
                return request(originalRequest);
            } catch (refreshError) {
                const axiosError = refreshError as any;
                
                // 检查是否是刷新 token 失败
                const isRefreshError = axiosError?.response?.status === 401;
                
                processQueue(axiosError, null);
                if (isRefreshError) {
                    handleAuthError();
                }
                return Promise.reject(refreshError);
            } finally {
                isRefreshing = false;
            }
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
            } else if (String(value) === "0") {
                searchParams.set(key, "0");
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
    post: <T, B = unknown>(url: string, body?: B, options?: {
        params?: Record<string, unknown>
    }) => fetchApi<T>(url, "POST", {body, ...options}),
    put: <T, B = unknown>(url: string, body?: B, options?: {
        params?: Record<string, unknown>
    }) => fetchApi<T>(url, "PUT", {body, ...options}),
    patch: <T, B = unknown>(url: string, body?: B, options?: {
        params?: Record<string, unknown>
    }) => fetchApi<T>(url, "PATCH", {body, ...options}),
    del: <T>(url: string, params?: Record<string, unknown>) => fetchApi<T>(url, "DELETE", {params}),
};

export type {Token, ApiError};
