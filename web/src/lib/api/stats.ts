// API 客户端 - 统计模块
import {api} from "../request";

export interface Stats {
    users: number;
    media: number;
    content: number;
    storage: string;
}

export const statsApi = {
    // 获取统计数据
    get: () => api.get<Stats>("/stats"),
};
