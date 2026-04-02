import {api, getAccessToken} from "../request";

export const DEFAULT_CHUNK_SIZE = 5 * 1024 * 1024; // 5 MB fallback
export const MAX_CONCURRENT_CHUNKS = 3; // 并发上传数
export const MAX_RETRIES = 3; // 每个分片最大重试次数
export const RETRY_DELAY_BASE = 1000; // 重试延迟基数 (ms)

// --- Types ---

export interface PartInfo {
    part_number: number;
    etag: string;
    size: number;
}

export type UploadStatus =
    | 'waiting'
    | 'initiating'
    | 'uploading'
    | 'paused'
    | 'completing'
    | 'success'
    | 'error'
    | 'aborted';

export interface UploadTask {
    id: string;
    file: File;
    progress: number;
    status: UploadStatus;
    error?: string;
    uploadId?: string;
    parts: PartInfo[];
    // Metadata
    title?: string;
    description?: string;
    categoryId?: number;
    tags?: string[];
    thumbnail?: string;
    // Timing
    startedAt?: number;
    completedAt?: number;
    speed?: number; // bytes/s
}

export interface UploadCallbacks {
    onProgress: (taskId: string, progress: number, speed?: number) => void;
    onStatusChange: (taskId: string, status: UploadStatus) => void;
    onSuccess: (taskId: string) => void;
    onError: (taskId: string, error: string) => void;
}

// --- API layer ---

const getApiBaseUrl = () => import.meta.env.VITE_API_BASE_URL || 'http://localhost:9090';
const getToken = () => getAccessToken();

interface InitiateResponse {
    upload_id: string;
    total_parts: number;
    chunk_size: number;
}

interface UploadPartResponse {
    etag: string;
    size: number;
}

interface ListPartsResponse {
    parts: PartInfo[];
    total_parts: number;
    chunk_size: number;
    uploaded_size: number;
    total_size: number;
    status: string;
}

// 初始化分片上传
async function initiateMultipartUpload(task: UploadTask): Promise<InitiateResponse> {
    return api.post<InitiateResponse>('/uploads/multipart', {
        filename: task.file.name,
        file_size: task.file.size,
        content_type: task.file.type || 'application/octet-stream',
        title: task.title || task.file.name.replace(/\.[^.]+$/, ''),
        description: task.description || '',
        category_id: task.categoryId || null,
        tags: task.tags || [],
        thumbnail: task.thumbnail || '',
    });
}

// 上传单个分片
async function uploadPart(
    uploadId: string,
    partNumber: number,
    data: Blob,
    signal?: AbortSignal,
): Promise<UploadPartResponse> {
    const res = await fetch(
        `${getApiBaseUrl()}/api/v1/uploads/${uploadId}/parts/${partNumber}`,
        {
            method: 'POST',
            headers: {
                ...(getToken() ? {Authorization: `Bearer ${getToken()}`} : {}),
                // Content-Type will be set automatically for blob
            },
            body: data,
            signal,
        },
    );
    if (!res.ok) {
        const err = await res.json().catch(() => ({error: `HTTP ${res.status}`}));
        throw new Error(err.error || `Upload part failed: ${res.status}`);
    }
    return res.json();
}

// 查询已上传分片 (断点续传)
async function listParts(uploadId: string): Promise<ListPartsResponse> {
    return api.get<ListPartsResponse>(`/uploads/${uploadId}/parts`);
}

// 更新元数据
export async function updateUploadMetadataApi(task: UploadTask): Promise<void> {
    if (!task.uploadId) return;
    return api.patch(`/uploads/${task.uploadId}/metadata`, {
        title: task.title,
        description: task.description,
        category_id: task.categoryId,
        tags: task.tags,
        thumbnail: task.thumbnail,
    });
}

// 完成分片上传
async function completeMultipartUpload(
    task: UploadTask,
    parts: PartInfo[],
    sha256?: string,
): Promise<{ media: unknown }> {
    return api.post<{ media: unknown }>(`/uploads/${task.uploadId}/complete`, {
        upload_id: task.uploadId,
        parts: parts.sort((a, b) => a.part_number - b.part_number),
        sha256: sha256 || '',
        // Include latest metadata at completion
        title: task.title,
        description: task.description,
        category_id: task.categoryId,
        tags: task.tags,
        thumbnail: task.thumbnail,
    });
}

// 取消上传
async function abortMultipartUpload(uploadId: string): Promise<void> {
    await api.del(`/uploads/${uploadId}`);
}

// 计算 SHA-256 (备用)
async function calculateSHA256(file: File): Promise<string> {
    const buffer = await file.arrayBuffer();
    const hashBuffer = await crypto.subtle.digest('SHA-256', buffer);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return hashArray.map((b) => b.toString(16).padStart(2, '0')).join('');
}

// --- Core upload engine ---

const activeTasks = new Map<string, AbortController>();

// 启动分片上传
export async function startMultipartUpload(
    task: UploadTask,
    callbacks: UploadCallbacks,
): Promise<void> {
    const abortController = new AbortController();
    activeTasks.set(task.id, abortController);
    task.startedAt = Date.now();

    try {
        callbacks.onStatusChange(task.id, 'initiating');
        let uploadId = task.uploadId;
        let actualChunkSize = DEFAULT_CHUNK_SIZE;
        const partsMap = new Map<number, PartInfo>();

        if (uploadId) {
            // Resume: fetch existing parts
            try {
                const syncResp = await listParts(uploadId);
                actualChunkSize = syncResp.chunk_size || DEFAULT_CHUNK_SIZE;
                if (syncResp.parts) {
                    syncResp.parts.forEach((p) => partsMap.set(p.part_number, p));
                    const progress = Math.round(
                        (partsMap.size / Math.ceil(task.file.size / actualChunkSize)) * 100,
                    );
                    callbacks.onProgress(task.id, progress);
                }
                if (syncResp.status === 'completed') {
                    callbacks.onStatusChange(task.id, 'success');
                    callbacks.onProgress(task.id, 100);
                    callbacks.onSuccess(task.id);
                    return;
                }
                if (syncResp.status === 'aborted') {
                    uploadId = undefined;
                }
            } catch {
                uploadId = undefined;
            }
        }

        if (!uploadId) {
            const resp = await initiateMultipartUpload(task);
            uploadId = resp.upload_id;
            task.uploadId = uploadId;
            actualChunkSize = resp.chunk_size || DEFAULT_CHUNK_SIZE;
        }

        // 2. Calculate pending parts
        const totalChunks = Math.ceil(task.file.size / actualChunkSize);
        const pending: number[] = [];
        for (let i = 0; i < totalChunks; i++) {
            if (!partsMap.has(i + 1)) {
                pending.push(i);
            }
        }

        // 3. Upload parts concurrently
        callbacks.onStatusChange(task.id, 'uploading');
        const {signal} = abortController;

        // Worker function
        const worker = async () => {
            while (pending.length > 0 && !signal?.aborted) {
                const idx = pending.shift();
                if (idx === undefined) break;

                const partNum = idx + 1;
                const chunk = task.file.slice(
                    idx * actualChunkSize,
                    Math.min((idx + 1) * actualChunkSize, task.file.size),
                );

                let retry = 0;
                let etag: string | null = null;

                while (retry < MAX_RETRIES && !etag && !signal?.aborted) {
                    try {
                        const resp = await uploadPart(uploadId!, partNum, chunk, signal);
                        etag = resp.etag;
                        partsMap.set(partNum, {
                            part_number: partNum,
                            etag: resp.etag,
                            size: resp.size,
                        });
                    } catch (e) {
                        if (signal?.aborted) throw e;
                        retry++;
                        if (retry >= MAX_RETRIES) {
                            pending.unshift(idx);
                            throw e;
                        }
                        await new Promise((r) => setTimeout(r, RETRY_DELAY_BASE * retry));
                    }
                }

                if (etag) {
                    const progress = Math.round((partsMap.size / totalChunks) * 100);
                    const elapsed = (Date.now() - (task.startedAt || Date.now())) / 1000;
                    const speed = elapsed > 0 ? (task.file.size * (progress / 100)) / elapsed : 0;
                    task.speed = speed;
                    callbacks.onProgress(task.id, progress, speed);
                }
            }
        };

        // Launch workers
        const workerCount = Math.min(MAX_CONCURRENT_CHUNKS, pending.length);
        await Promise.all(
            Array.from({length: workerCount}, () => worker()),
        );

        if (signal?.aborted) {
            callbacks.onStatusChange(task.id, 'aborted');
            return;
        }

        // 4. Complete upload
        if (partsMap.size === totalChunks) {
            callbacks.onStatusChange(task.id, 'completing');
            callbacks.onProgress(task.id, 100);

            const sha256 = '';
            const finalParts = Array.from(partsMap.values());

            await completeMultipartUpload(task, finalParts, sha256);

            task.completedAt = Date.now();
            callbacks.onStatusChange(task.id, 'success');
            callbacks.onSuccess(task.id);
        } else {
            throw new Error(`Upload incomplete: ${partsMap.size}/${totalChunks} parts finished`);
        }
    } catch (e) {
        abortController.abort();
        if (abortController.signal.aborted && e instanceof DOMException && e.name === 'AbortError') {
            callbacks.onStatusChange(task.id, 'aborted');
        } else {
            const message = e instanceof Error ? e.message : 'Upload failed';
            callbacks.onError(task.id, message);
            callbacks.onStatusChange(task.id, 'error');
        }
    } finally {
        activeTasks.delete(task.id);
    }
}

// 取消上传
export function cancelUpload(taskId: string): void {
    const controller = activeTasks.get(taskId);
    if (controller) {
        controller.abort();
    }
}

// 暂停上传
export function pauseUpload(taskId: string): void {
    cancelUpload(taskId);
}

// 检查是否使用分片上传
export function shouldUseChunkedUpload(fileSize: number): boolean {
    return fileSize > DEFAULT_CHUNK_SIZE;
}
