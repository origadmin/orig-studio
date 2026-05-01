// Metadata API
import {api} from "../request";

export interface MediaMetadata {
    id: string;
    media_id: string;
    title: string;
    description: string;
    duration: number;
    width: number;
    height: number;
    format: string;
    codec: string;
    bitrate: number;
    frame_rate: number;
    resolution: string;
    aspect_ratio: string;
    file_size: number;
    create_time: string;
    update_time: string;
    // 高级元数据
    key_frames?: KeyFrame[];
    audio_waveform?: AudioWaveform;
    text_content?: TextContent;
    scene_changes?: SceneChange[];
}

export interface KeyFrame {
    id: string;
    time: number;
    url: string;
    thumbnail_url: string;
}

export interface AudioWaveform {
    id: string;
    samples: number[];
    url: string;
}

export interface TextContent {
    id: string;
    transcript: string;
    keywords: string[];
    entities: Entity[];
}

export interface Entity {
    type: string;
    text: string;
    confidence: number;
}

export interface SceneChange {
    id: string;
    time: number;
    description: string;
}

export const metadataApi = {
    // 获取媒体元数据
    getByMediaId: (mediaId: string) =>
        api.get<MediaMetadata>(`/medias/${mediaId}/metadata`),

    // 触发元数据挖掘
    triggerMining: (mediaId: string) =>
        api.post<{ success: boolean; message: string }>(`/medias/${mediaId}/metadata/mining`),

    // 获取元数据挖掘状态
    getMiningStatus: (mediaId: string) =>
        api.get<{ status: string; progress: number; message: string }>(`/medias/${mediaId}/metadata/status`),

    // 获取关键帧
    getKeyFrames: (mediaId: string) =>
        api.get<KeyFrame[]>(`/medias/${mediaId}/metadata/key-frames`),

    // 获取音频波形
    getAudioWaveform: (mediaId: string) =>
        api.get<AudioWaveform>(`/medias/${mediaId}/metadata/audio-waveform`),

    // 获取文本内容
    getTextContent: (mediaId: string) =>
        api.get<TextContent>(`/medias/${mediaId}/metadata/text-content`),

    // 获取场景变化
    getSceneChanges: (mediaId: string) =>
        api.get<SceneChange[]>(`/medias/${mediaId}/metadata/scene-changes`),
};
