# M2 媒体上传与播放 - 分片上传实现计划

> **创建时间**：2026-04-01  
> **参考实现**：`projects/webui/src/utils/upload.ts` + `projects/backend/api/v1/proto/filemanager/`

## 1. 功能需求

### 1.1 分片上传
- 支持大文件（>2MB）自动分片上传
- 分片大小：2MB（与 Kratos gateway 兼容）
- 并发上传：最多 3 个分片同时上传
- 上传进度实时显示

### 1.2 断点续传
- 上传中断后可恢复
- 记录已上传分片状态
- 支持查询上传进度
- 支持取消/中止上传

### 1.3 文件类型支持
- 视频：mp4, mov, avi, mkv, webm
- 图片：jpg, jpeg, png, gif, webp
- 最大文件大小：500MB（可配置）

## 2. API 设计

### 2.1 分片上传流程

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  1. Init    │────▶│  2. Upload  │────▶│  3. Complete│
│  获取upload_id│     │  分片上传    │     │  合并分片   │
└─────────────┘     └─────────────┘     └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │  4. Abort   │
                    │  取消上传    │
                    └─────────────┘
```

### 2.2 Proto 定义

```protobuf
// 分片上传服务
service UploadService {
  // 初始化分片上传
  rpc InitiateMultipartUpload(InitiateMultipartUploadRequest) returns (InitiateMultipartUploadResponse) {
    option (google.api.http) = {post: "/api/v1/uploads/multipart"; body: "*"};
  }
  
  // 上传分片
  rpc UploadPart(UploadPartRequest) returns (UploadPartResponse) {
    option (google.api.http) = {post: "/api/v1/uploads/{upload_id}/parts/{part_number}"; body: "*"};
  }
  
  // 列出已上传分片（用于断点续传）
  rpc ListParts(ListPartsRequest) returns (ListPartsResponse) {
    option (google.api.http) = {get: "/api/v1/uploads/{upload_id}/parts"};
  }
  
  // 完成分片上传
  rpc CompleteMultipartUpload(CompleteMultipartUploadRequest) returns (CompleteMultipartUploadResponse) {
    option (google.api.http) = {post: "/api/v1/uploads/{upload_id}/complete"; body: "*"};
  }
  
  // 取消分片上传
  rpc AbortMultipartUpload(AbortMultipartUploadRequest) returns (AbortMultipartUploadResponse) {
    option (google.api.http) = {delete: "/api/v1/uploads/{upload_id}"};
  }
  
  // 简单上传（小文件）
  rpc UploadFile(UploadFileRequest) returns (UploadFileResponse) {
    option (google.api.http) = {post: "/api/v1/uploads"; body: "*"};
  }
}

// 初始化请求
message InitiateMultipartUploadRequest {
  string filename = 1 [json_name = "filename"];
  int64 file_size = 2 [json_name = "file_size"];
  string content_type = 3 [json_name = "content_type"];
  string title = 4 [json_name = "title"];
  string description = 5 [json_name = "description"];
  int64 category_id = 6 [json_name = "category_id"];
}

// 初始化响应
message InitiateMultipartUploadResponse {
  string upload_id = 1 [json_name = "upload_id"];
  int32 total_parts = 2 [json_name = "total_parts"];
  int32 chunk_size = 3 [json_name = "chunk_size"];
}

// 分片信息
message PartInfo {
  int32 part_number = 1 [json_name = "part_number"];
  string etag = 2 [json_name = "etag"];
  int64 size = 3 [json_name = "size"];
}

// 上传分片请求
message UploadPartRequest {
  string upload_id = 1 [json_name = "upload_id"];
  int32 part_number = 2 [json_name = "part_number"];
  bytes data = 3 [json_name = "data"];
}

// 上传分片响应
message UploadPartResponse {
  string etag = 1 [json_name = "etag"];
  int64 size = 2 [json_name = "size"];
}

// 列出分片请求
message ListPartsRequest {
  string upload_id = 1 [json_name = "upload_id"];
}

// 列出分片响应
message ListPartsResponse {
  repeated PartInfo parts = 1 [json_name = "parts"];
  int32 total_parts = 2 [json_name = "total_parts"];
  int64 uploaded_size = 3 [json_name = "uploaded_size"];
  int64 total_size = 4 [json_name = "total_size"];
}

// 完成上传请求
message CompleteMultipartUploadRequest {
  string upload_id = 1 [json_name = "upload_id"];
  repeated PartInfo parts = 2 [json_name = "parts"];
  string sha256 = 3 [json_name = "sha256"]; // 文件完整性校验
}

// 完成上传响应
message CompleteMultipartUploadResponse {
  api.v1.services.types.Media media = 1 [json_name = "media"];
}

// 取消上传请求
message AbortMultipartUploadRequest {
  string upload_id = 1 [json_name = "upload_id"];
}

// 取消上传响应
message AbortMultipartUploadResponse {}

// 简单上传请求
message UploadFileRequest {
  bytes data = 1 [json_name = "data"];
  string filename = 2 [json_name = "filename"];
  string content_type = 3 [json_name = "content_type"];
  string title = 4 [json_name = "title"];
  string description = 5 [json_name = "description"];
  int64 category_id = 6 [json_name = "category_id"];
}

// 简单上传响应
message UploadFileResponse {
  api.v1.services.types.Media media = 1 [json_name = "media"];
}
```

## 3. 数据库设计

### 3.1 UploadSession 表

```go
// UploadSession represents an upload session for multipart uploads
type UploadSession struct {
    ent.Schema
}

func (UploadSession) Fields() []ent.Field {
    return []ent.Field{
        field.String("upload_id").Unique().NotEmpty(),
        field.String("filename").NotEmpty(),
        field.Int64("file_size").Positive(),
        field.String("content_type").Optional(),
        field.Int("total_parts").Positive(),
        field.Int("chunk_size").Default(2 * 1024 * 1024), // 2MB
        field.Int64("uploaded_size").Default(0),
        field.String("title").Optional(),
        field.String("description").Optional(),
        field.Int64("category_id").Optional(),
        field.Int64("user_id").Optional(),
        field.Enum("status").Values("pending", "uploading", "completed", "aborted").Default("pending"),
        field.Time("expires_at").Optional(),
        field.JSON("parts", map[int]string{}), // part_number -> etag
        field.String("sha256").Optional(),
        field.Time("created_at").Default(time.Now),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}
```

### 3.2 Media 表扩展

```go
// 添加字段
field.String("upload_id").Optional(), // 关联上传会话
field.String("sha256").Optional(),    // 文件哈希
field.String("storage_path").Optional(), // 存储路径
```

## 4. 后端实现

### 4.1 目录结构

```
internal/
├── svc-upload/           # 上传服务
│   ├── service/
│   │   └── upload.go     # gRPC 服务实现
│   ├── biz/
│   │   └── upload.go     # 业务逻辑
│   └── data/
│       └── upload_repo.go # 数据访问
├── data/
│   └── entity/
│       └── upload_session.go # Ent schema
└── storage/
    └── local.go          # 本地存储实现
```

### 4.2 核心逻辑

```go
// biz/upload.go
type UploadUseCase struct {
    repo        UploadRepository
    storage     Storage
    mediaRepo   MediaRepository
    chunkSize   int
}

// InitiateMultipartUpload 初始化分片上传
func (uc *UploadUseCase) InitiateMultipartUpload(ctx context.Context, req *InitiateMultipartUploadRequest) (*InitiateMultipartUploadResponse, error) {
    // 1. 验证文件类型
    if !isValidFileType(req.Filename, req.ContentType) {
        return nil, errors.New("invalid file type")
    }
    
    // 2. 验证文件大小
    if req.FileSize > maxFileSize {
        return nil, errors.New("file too large")
    }
    
    // 3. 生成 upload_id
    uploadID := uuid.New().String()
    
    // 4. 计算分片数量
    totalParts := int(math.Ceil(float64(req.FileSize) / float64(uc.chunkSize)))
    
    // 5. 创建上传会话
    session := &UploadSession{
        UploadID:    uploadID,
        Filename:    req.Filename,
        FileSize:    req.FileSize,
        ContentType: req.ContentType,
        TotalParts:  totalParts,
        ChunkSize:   uc.chunkSize,
        Title:       req.Title,
        Description: req.Description,
        CategoryID:  req.CategoryID,
        UserID:      ctx.Value("user_id").(int64),
        Status:      "pending",
        Parts:       make(map[int]string),
        ExpiresAt:   time.Now().Add(24 * time.Hour),
    }
    
    if err := uc.repo.CreateSession(ctx, session); err != nil {
        return nil, err
    }
    
    return &InitiateMultipartUploadResponse{
        UploadID:   uploadID,
        TotalParts: int32(totalParts),
        ChunkSize:  int32(uc.chunkSize),
    }, nil
}

// UploadPart 上传分片
func (uc *UploadUseCase) UploadPart(ctx context.Context, uploadID string, partNumber int, data []byte) (*UploadPartResponse, error) {
    // 1. 获取上传会话
    session, err := uc.repo.GetSession(ctx, uploadID)
    if err != nil {
        return nil, err
    }
    
    // 2. 检查会话状态
    if session.Status == "completed" || session.Status == "aborted" {
        return nil, errors.New("upload session is not active")
    }
    
    // 3. 存储分片
    etag, err := uc.storage.StorePart(ctx, uploadID, partNumber, data)
    if err != nil {
        return nil, err
    }
    
    // 4. 更新会话状态
    session.Parts[partNumber] = etag
    session.UploadedSize += int64(len(data))
    session.Status = "uploading"
    
    if err := uc.repo.UpdateSession(ctx, session); err != nil {
        return nil, err
    }
    
    return &UploadPartResponse{
        Etag: etag,
        Size: int64(len(data)),
    }, nil
}

// CompleteMultipartUpload 完成分片上传
func (uc *UploadUseCase) CompleteMultipartUpload(ctx context.Context, uploadID string, parts []*PartInfo, sha256 string) (*Media, error) {
    // 1. 获取上传会话
    session, err := uc.repo.GetSession(ctx, uploadID)
    if err != nil {
        return nil, err
    }
    
    // 2. 验证所有分片已上传
    if len(parts) != session.TotalParts {
        return nil, errors.New("not all parts uploaded")
    }
    
    // 3. 合并分片
    finalPath, err := uc.storage.MergeParts(ctx, uploadID, parts)
    if err != nil {
        return nil, err
    }
    
    // 4. 计算文件哈希（可选）
    if sha256 != "" {
        // 验证哈希...
    }
    
    // 5. 创建媒体记录
    media := &Media{
        Title:       session.Title,
        Description: session.Description,
        Type:        getMediaType(session.ContentType),
        URL:         finalPath,
        Size:        session.FileSize,
        MimeType:    session.ContentType,
        UserID:      session.UserID,
        CategoryID:  session.CategoryID,
        Status:      1, // pending
        UploadID:    uploadID,
        Sha256:      sha256,
    }
    
    if err := uc.mediaRepo.Create(ctx, media); err != nil {
        return nil, err
    }
    
    // 6. 更新会话状态
    session.Status = "completed"
    uc.repo.UpdateSession(ctx, session)
    
    return media, nil
}
```

## 5. 前端实现

### 5.1 目录结构

```
web/src/
├── lib/
│   └── upload/
│       ├── index.ts          # 导出
│       ├── multipart.ts      # 分片上传核心逻辑
│       ├── types.ts          # 类型定义
│       └── api.ts            # API 封装
├── components/
│   └── upload/
│       ├── UploadZone.tsx    # 拖拽上传区域
│       ├── UploadProgress.tsx # 上传进度
│       ├── UploadTask.tsx    # 单个上传任务
│       └── UploadManager.tsx # 上传管理器
└── hooks/
    └── useUpload.ts          # 上传 Hook
```

### 5.2 核心代码

```typescript
// lib/upload/multipart.ts
export const CHUNK_SIZE = 2 * 1024 * 1024; // 2MB
export const MAX_CONCURRENT_CHUNKS = 3;
export const CHUNK_TIMEOUT = 30000; // 30s

export interface UploadPartInfo {
  part_number: number;
  etag: string;
  size: number;
}

export interface UploadTaskState {
  id: string;
  file: File;
  fileName: string;
  fileSize: number;
  progress: number;
  status: UploadStatus;
  error?: string;
  uploadedBytes: number;
  totalBytes: number;
  uploadId?: string;
  parts: UploadPartInfo[];
  title?: string;
  description?: string;
  categoryId?: number;
}

export type UploadStatus = 
  | "waiting" 
  | "uploading" 
  | "paused" 
  | "error" 
  | "success" 
  | "completing";

// 分片上传 API
export interface MultipartUploadApi {
  initiate: (data: InitiateRequest) => Promise<InitiateResponse>;
  uploadPart: (uploadId: string, partNumber: number, data: Blob, signal?: AbortSignal) => Promise<string>;
  listParts: (uploadId: string) => Promise<ListPartsResponse>;
  complete: (uploadId: string, parts: UploadPartInfo[], sha256?: string) => Promise<Media>;
  abort: (uploadId: string) => Promise<void>;
}

// 分片上传核心逻辑
export async function uploadMultipartFile(
  task: UploadTaskState,
  api: MultipartUploadApi,
  options: UploadOptions
): Promise<void> {
  const { file } = task;
  const { onProgress, onSuccess, onError, onStatusChange, signal, chunkSize = CHUNK_SIZE } = options;

  let uploadId = task.uploadId;
  const partsMap = new Map<number, string>();

  // 1. 初始化或恢复上传
  if (!uploadId) {
    const resp = await api.initiate({
      filename: file.name,
      file_size: file.size,
      content_type: file.type || "application/octet-stream",
      title: task.title,
      description: task.description,
      category_id: task.categoryId,
    });
    uploadId = resp.upload_id;
    onProgress(0, { uploadId, parts: [] });
  } else {
    // 断点续传：同步已上传分片
    const syncResp = await api.listParts(uploadId);
    if (syncResp?.parts) {
      syncResp.parts.forEach(p => partsMap.set(p.part_number, p.etag));
      const progress = Math.round((partsMap.size / Math.ceil(file.size / chunkSize)) * 100);
      onProgress(progress, { parts: Array.from(partsMap.entries()).map(([pn, et]) => ({ part_number: pn, etag: et })) });
    }
  }

  // 2. 计算待上传分片
  const totalChunks = Math.ceil(file.size / chunkSize);
  const pending: number[] = [];
  for (let i = 0; i < totalChunks; i++) {
    if (!partsMap.has(i + 1)) pending.push(i);
  }

  // 3. 并发上传分片
  const worker = async () => {
    while (pending.length > 0 && !signal?.aborted) {
      const i = pending.shift();
      if (i === undefined) break;

      const partNum = i + 1;
      const chunk = file.slice(i * chunkSize, Math.min((i + 1) * chunkSize, file.size));

      let retry = 0;
      let etag: string | null = null;

      while (retry < 3 && !etag && !signal?.aborted) {
        try {
          etag = await api.uploadPart(uploadId!, partNum, chunk, signal);
        } catch (e) {
          if (signal?.aborted) throw e;
          retry++;
          if (retry >= 3) {
            pending.unshift(i);
            throw e;
          }
          await new Promise(r => setTimeout(r, 1000 * retry));
        }
      }

      if (etag) {
        partsMap.set(partNum, etag);
        const progress = Math.round((partsMap.size / totalChunks) * 100);
        onProgress(progress, {
          parts: Array.from(partsMap.entries()).map(([pn, et]) => ({ part_number: pn, etag: et })),
        });
      }
    }
  };

  await Promise.all(Array(Math.min(MAX_CONCURRENT_CHUNKS, pending.length)).fill(null).map(worker));

  // 4. 完成上传
  if (partsMap.size === totalChunks && !signal?.aborted) {
    onStatusChange("completing");
    const sha256 = await calculateFileHash(file);
    const finalParts = Array.from(partsMap.entries())
      .map(([pn, et]) => ({ part_number: pn, etag: et }))
      .sort((a, b) => a.part_number - b.part_number);

    const result = await api.complete(uploadId!, finalParts, sha256);
    onProgress(100, { status: "success", progress: 100 });
    onSuccess(result);
  }
}
```

## 6. 实现步骤

### Phase 1: 后端基础 (预计 3 天)

1. **T2.1** 创建 proto 定义
   - [ ] 创建 `api/proto/v1/upload/upload.proto`
   - [ ] 运行 `buf generate` 生成代码

2. **T2.2** 创建数据模型
   - [ ] 创建 `internal/data/entity/upload_session.go` Ent schema
   - [ ] 运行 `go generate ./...` 生成代码

3. **T2.3** 实现存储层
   - [ ] 创建 `internal/storage/local.go` 本地存储实现
   - [ ] 实现分片存储、合并、删除功能

### Phase 2: 后端服务 (预计 4 天)

4. **T2.4** 实现上传服务
   - [ ] 创建 `internal/svc-upload/` 目录结构
   - [ ] 实现 biz 层业务逻辑
   - [ ] 实现 service 层 gRPC handler
   - [ ] 注册路由到 cmd/server

5. **T2.5** 实现断点续传
   - [ ] 实现 ListParts 接口
   - [ ] 添加上传会话过期清理逻辑

### Phase 3: 前端实现 (预计 3 天)

6. **T2.6** 实现上传核心逻辑
   - [ ] 创建 `web/src/lib/upload/` 目录
   - [ ] 实现分片上传、断点续传逻辑
   - [ ] 实现 API 封装

7. **T2.7** 实现上传组件
   - [ ] 实现 UploadZone 拖拽上传区域
   - [ ] 实现 UploadProgress 进度显示
   - [ ] 实现 UploadManager 任务管理

### Phase 4: 集成测试 (预计 2 天)

8. **T2.8** 集成测试
   - [ ] 测试小文件上传
   - [ ] 测试大文件分片上传
   - [ ] 测试断点续传
   - [ ] 测试取消上传

## 7. 验收标准

- [ ] 支持上传 500MB 以内的视频/图片文件
- [ ] 大文件（>2MB）自动分片上传
- [ ] 上传中断后可恢复（断点续传）
- [ ] 上传进度实时显示
- [ ] 支持取消/中止上传
- [ ] 文件类型验证正确
- [ ] 上传完成后媒体列表正确显示
