// Subtitle API
import {api} from "../request";

export interface Subtitle {
    id: string;
    media_id: string;
    language: string;
    language_name: string;
    file_path: string;
    status: string;
    created_at: string;
    updated_at: string;
}

export const subtitleApi = {
    // 获取媒体的字幕列表
    getByMediaId: (mediaId: string) =>
        api.get<Subtitle[]>(`/media/${mediaId}/subtitles`),

    // 上传字幕
    upload: (mediaId: string, file: File, language: string) => {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('language', language);
        return api.post<Subtitle>(`/media/${mediaId}/subtitles`, formData, {
            headers: {
                'Content-Type': 'multipart/form-data',
            },
        });
    },

    // 删除字幕
    delete: (id: string) =>
        api.del<void>(`/subtitles/${id}`),

    // 获取支持的语言列表
    getLanguages: () =>
        api.get<{ code: string; name: string }[]>('/subtitles/languages'),
};
