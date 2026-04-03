# M3：视频转码与 HLS — 详细开发计划

> **制定时间**：2026-04-03
> **最后更新**：2026-04-03
> **参考**：[MILESTONES.md](../MILESTONES.md) M3 章节
> **架构决策**：D4 使用 Watermill GoChannel 替代原始 `go ProcessMedia()`

---

## 架构设计

### 消息流

```
CompleteMultipartUpload → publisher.Publish("media.encode.request", msg)
                        ↓
              [Watermill GoChannel]
                        ↓
              TranscodeHandler.Handle(msg)
                        ↓
              TranscodeWorker.Submit(job) × N profiles
                        ↓
          ┌─────────────────────────────┐
          │  Goroutine Pool + Semaphore  │  ← semaphore.Weighted(maxWorkers)
          │  limits concurrent ffmpeg    │     e.g. maxWorkers=3
          └─────────────────────────────┘
                        ↓
              ffmpeg TranscodeToMP4 → Bento4 MP4HLS
                        ↓
              publisher.Publish("media.encode.progress", event)
```

### TranscodeWorker 接口

```go
// svc-media/biz/transcode_worker.go
type TranscodeWorker interface {
    // Submit enqueues a transcode job. Returns error if worker pool is full.
    Submit(ctx context.Context, job TranscodeJob) error
    // Status returns current worker pool status.
    Status() WorkerPoolStatus
    // Shutdown gracefully waits for running jobs to complete.
    Shutdown(ctx context.Context) error
}

type TranscodeJob struct {
    MediaID       int64
    TaskID        int64
    Profile       *EncodeProfile
    InputPath     string
    OutputPath    string
    ContentType   string
}

type WorkerPoolStatus struct {
    MaxWorkers    int32
    ActiveWorkers int32
    PendingJobs   int32
}
```

CE uses `goroutineWorker` (goroutine pool + `semaphore.Weighted`).
The interface allows EE to provide an `asynqWorker` implementation without changing any caller code.

### 并发控制策略

- 每个 profile 启动一个 goroutine，通过 `semaphore.Weighted(maxWorkers)` 限制同时执行的 ffmpeg 数量
- 默认 `maxWorkers=3`，通过配置文件可调
- 获取不到信号量的 goroutine 阻塞等待，不会无限创建 ffmpeg 进程
- `Shutdown()` 优雅等待所有进行中的任务完成

---

## 现状盘点

### 已有代码（可直接复用）

| 模块 | 文件 | 说明 |
|------|------|------|
| ffmpeg 封装 | `internal/helpers/ffmpeg/ffmpeg.go` | ExtractThumbnail, GetVideoDuration, TranscodeToMP4 |
| Bento4 HLS | `internal/helpers/ffmpeg/bento4.go` | MP4HLS（多码率自适应）, MP4Mux |
| ffmpeg HLS（备选） | `internal/helpers/ffmpeg/hls.go` | TranscodeToHLS（单码率 720p，未使用） |
| 转码流程 | `svc-media/biz/upload.go:341-473` | ProcessMedia 完整四步流程 |
| Profile 定义 | `svc-media/biz/media.go:37-50` | EncodeProfile struct |
| Task 定义 | `svc-media/biz/media.go:62-79` | EncodingTask struct |
| 事件机制 | `svc-media/biz/media.go:81-224` | MediaUseCase.Subscribe/Publish（内存 channel） |
| Task 仓储 | `svc-media/data/encoding_task_repo.go` | Create/Update/Get/ListByMedia |
| Profile 仓储 | `svc-media/data/encode_profile_repo.go` | ListActive/ListAll/Get/Create/Update/Delete |
| 种子数据 | `svc-media/data/seed.go` | 22 个预置 profile |
| PubSub 框架 | `internal/pubsub/pubsub.go` | Topic 常量（仅 user 事件） |
| Publisher Provider | `internal/helpers/providers/common.go` | ProvidePublisher（runtime container → nil 降级） |
| Watermill 依赖 | `go.mod` | `watermill v1.5.1` |
| HLS 静态路由 | `cmd/server/main.go:182` | `/hls` → `./data/uploads/hls` |
| Ent Schema | `internal/data/entity/schema/encoding_task.go` | media_id, profile_id, status, progress, output_path, error_message |
| Ent Schema | `internal/data/entity/schema/encode_profile.go` | name, resolution, video_codec, video_bitrate, audio_codec, is_active |
| Ent Schema | `internal/data/entity/schema/media.go` | hls_file, encoding_status |
| Proto 类型 | `api/gen/v1/types/media.pb.go` | Media.HlsFile, Media.EncodingStatus, EncodeProfile |
| 参考实现 | `internal/svc-user/service/user.go` | Watermill Publisher 注入示例（backend 模式） |

### 已知问题（需在 M3 中修复）

1. **resolution 格式不一致**：seed 用 `"720"`，`TranscodeToMP4` 需要 `"1280x720"` → 需要转换函数
2. **bitrate 未使用**：`TranscodeToMP4` 忽略了 `video_bitrate` / `audio_bitrate` 字段
3. **BentoParameters 断层**：biz 有字段但 ent schema 和 seed 都没包含，MP4HLS 调用未传入
4. **进程触发**：`go uc.ProcessMedia(context.Background(), ...)` 无超时控制、无重试、无并发限制
5. **内存 Pub/Sub**：`MediaUseCase.Subscribe` 是纯内存 channel，重启丢失

---

## 开发任务分解

### Phase 1：基础设施（Watermill 集成）

#### T3.0 引入 GoChannel 依赖 + PubSub Topic 定义

**文件**：`internal/pubsub/pubsub.go`（修改）、`go.mod`（修改）

**新增 Topic 常量**：
```go
// Media encoding events
const MediaEncodeRequestTopic  = "media.encode.request"   // 触发转码
const MediaEncodeProgressTopic = "media.encode.progress"  // 进度推送
const MediaEncodeCompletedTopic = "media.encode.completed" // 转码完成
```

**新增 GoChannel 依赖**：
```bash
go get github.com/ThreeDotsLabs/watermill@v1.5.1
go get github.com/ThreeDotsLabs/watermill/message/infrastructure/gochannel@latest
```

> 注：watermill v1.5.1 的 GoChannel 包路径需确认，可能是 `github.com/ThreeDotsLabs/watermill/pubsub/gochannel`

**验收**：`go build ./...` 无错误

---

#### T3.1 GoChannel Publisher + Subscriber 工厂

**文件**：`internal/pubsub/pubsub.go`（修改）或新建 `internal/pubsub/factory.go`

**核心逻辑**：
```go
// NewGoChannelPubSub creates an in-process publisher/subscriber (zero external deps)
func NewGoChannelPubSub(logger watermill.LoggerAdapter) (message.Publisher, message.Subscriber) {
    pubSub := gochannel.NewGoChannel(
        gochannel.Config{
            OutputChannelBuffer: 64, // buffer size for pending messages
        },
        logger,
    )
    return pubSub, pubSub // GoChannel implements both interfaces
}
```

**验收**：`go build ./...` 无错误

---

### Phase 2：转码 Worker

#### T3.2 TranscodeWorker 接口 + goroutineWorker 实现

**文件**：
- `svc-media/biz/transcode_worker.go`（新建）— 接口定义 + goroutine 实现
- `svc-media/biz/transcode_worker_test.go`（新建）

**接口定义**（见上方"架构设计"）。

**CE 实现：goroutineWorker**

```go
type goroutineWorker struct {
    sem          *semaphore.Weighted
    maxWorkers   int32
    activeCount  atomic.Int32
    pendingCount atomic.Int32
    logger       *log.Helper
}

func NewGoroutineWorker(maxWorkers int32, logger *log.Helper) *goroutineWorker {
    return &goroutineWorker{
        sem:        semaphore.NewWeighted(int64(maxWorkers)),
        maxWorkers: maxWorkers,
        logger:     logger,
    }
}

func (w *goroutineWorker) Submit(ctx context.Context, job TranscodeJob) error {
    w.pendingCount.Add(1)

    // Acquire semaphore — blocks if maxWorkers are already running
    if err := w.sem.Acquire(ctx, 1); err != nil {
        w.pendingCount.Add(-1)
        return fmt.Errorf("worker pool shut down: %w", err)
    }

    w.pendingCount.Add(-1)
    w.activeCount.Add(1)

    go func() {
        defer func() {
            w.activeCount.Add(-1)
            w.sem.Release(1)
        }()

        // Execute the transcode job
        if err := executeTranscodeJob(ctx, job); err != nil {
            w.logger.Errorf("transcode job failed: media=%d profile=%s err=%v",
                job.MediaID, job.Profile.Name, err)
        }
    }()

    return nil
}

func (w *goroutineWorker) Status() WorkerPoolStatus {
    return WorkerPoolStatus{
        MaxWorkers:    w.maxWorkers,
        ActiveWorkers: w.activeCount.Load(),
        PendingJobs:   w.pendingCount.Load(),
    }
}

func (w *goroutineWorker) Shutdown(ctx context.Context) error {
    // Drain: wait for all active workers to finish
    for w.activeCount.Load() > 0 {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(500 * time.Millisecond):
        }
    }
    return nil
}
```

**配置项**（`configs/bootstrap.yaml`）：

```yaml
media:
  transcode:
    max_workers: 3       # concurrent ffmpeg processes
    job_timeout: 7200    # seconds (2h per job)
```

**验收**：`go test ./svc-media/biz/... -run TestGoroutineWorker` 通过

---

#### T3.3 TranscodeHandler + Watermill Router

**文件**：
- `internal/svc-media/biz/transcode_handler.go`（新建）
- `cmd/server/main.go`（修改）

**TranscodeHandler**：订阅 `media.encode.request`，为每个 profile 提交 worker job。

```go
type TranscodeHandler struct {
    mediaUC     *MediaUseCase
    profileRepo EncodeProfileRepo
    worker      TranscodeWorker    // 接口注入
    publisher   message.Publisher  // 进度事件
    logger      *log.Helper
}

func (h *TranscodeHandler) Handle(msg *message.Message) error {
    var req MediaEncodeRequest
    if err := json.Unmarshal(msg.Payload, &req); err != nil {
        return err
    }

    profiles, err := h.profileRepo.ListActive(msg.Context())
    if err != nil {
        return err
    }

    var wg sync.WaitGroup
    errCh := make(chan error, len(profiles))

    for _, p := range profiles {
        if p.ProfileType == "preview" {
            // GIF preview — skip or handle separately (see T3.8)
            continue
        }

        wg.Add(1)
        go func(profile *biz.EncodeProfile) {
            defer wg.Done()

            job := biz.TranscodeJob{
                MediaID:    req.MediaID,
                Profile:    profile,
                InputPath:  req.MediaPath,
                // OutputPath computed from profile settings
            }

            if err := h.worker.Submit(msg.Context(), job); err != nil {
                errCh <- fmt.Errorf("profile %s: %w", profile.Name, err)
            }
        }(p)
    }

    wg.Wait()
    close(errCh)

    // Aggregate errors, update media encoding_status accordingly
    var errs []error
    for e := range errCh {
        errs = append(errs, e)
    }

    if len(errs) > 0 {
        return fmt.Errorf("%d profile(s) failed: %v", len(errs), errors.Join(errs...))
    }

    // Publish completion event
    h.publishCompletion(msg.Context(), req.MediaID)
    return nil
}
```

**Router 注册**（`cmd/server/main.go`）：

```go
// Create GoChannel pubsub
pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)

// Create worker + handler
worker := biz.NewGoroutineWorker(cfg.Media.Transcode.MaxWorkers, logger)
handler := biz.NewTranscodeHandler(mediaUC, profileRepo, worker, pubSub, logger)

// Create and start router
router, _ := message.NewRouter(message.RouterConfig{}, logger)
router.AddHandler(
    "media_transcode",
    pubsub.MediaEncodeRequestTopic,
    pubSub,
    handler.Handle,
)
go router.Run(ctx)

// Inject publisher into UploadUseCase
uploadUC.SetPublisher(pubSub)
```

**验收**：服务启动后 router running，上传视频 → worker 接收 → 并发受 semaphore 限制

---

#### T3.4 UploadUseCase 改为发布消息

**文件**：`svc-media/biz/upload.go`（修改）

**改动**：
```go
type UploadUseCase struct {
    // ... existing fields
    publisher message.Publisher  // 新增
}

// CompleteMultipartUpload 末尾：
if strings.HasPrefix(session.ContentType, "video/") {
    // 旧：go uc.ProcessMedia(context.Background(), createdMedia.Id, finalPath, session.ContentType)
    // 新：发布转码请求到 Watermill
    payload, _ := json.Marshal(MediaEncodeRequest{
        MediaID:      createdMedia.Id,
        MediaPath:    finalPath,
        ContentType:  session.ContentType,
    })
    msg := message.NewMessage(watermill.NewUUID(), payload)
    uc.publisher.Publish(pubsub.MediaEncodeRequestTopic, msg)
}
```

**验收**：上传视频后触发 Watermill 消息，TranscodeHandler 接收并执行

---

### Phase 3：Bugs 修复

#### T3.5 Resolution 转换函数

**文件**：`internal/helpers/ffmpeg/ffmpeg.go`（修改）或新建 `internal/helpers/ffmpeg/resolution.go`

**问题**：seed 中 resolution 为 `"720"`，但 `TranscodeToMP4` 的 `-s` 参数需要 `"1280x720"`

**方案**：
```go
// ResolutionToSize converts "720" to "1280x720" etc.
// For non-standard values, returns as-is.
func ResolutionToSize(resolution string) string {
    sizes := map[string]string{
        "240":  "426x240",
        "360":  "640x360",
        "480":  "854x480",
        "720":  "1280x720",
        "1080": "1920x1080",
        "1440": "2560x1440",
        "2160": "3840x2160",
    }
    if s, ok := sizes[resolution]; ok {
        return s
    }
    // Already in WxH format
    if strings.Contains(resolution, "x") {
        return resolution
    }
    return resolution
}
```

**验收**：转码时 ffmpeg 收到正确的 `-s 1280x720` 参数

---

#### T3.6 Bitrate 参数透传

**文件**：`internal/helpers/ffmpeg/ffmpeg.go`（修改）

**当前**：`TranscodeToMP4` 不使用 bitrate 参数
**修改**：添加 `videoBitrate` / `audioBitrate` 参数，映射到 `-b:v` / `-b:a`

```go
func TranscodeToMP4(ctx, inputPath, outputPath, resolution, videoCodec, audioCodec,
    videoBitrate, audioBitrate string) error {
    args := []string{
        "-i", inputPath,
        "-s", ResolutionToSize(resolution),
        "-c:v", codec,  // libx264 or libx265
    }
    if videoBitrate != "" && videoBitrate != "auto" {
        args = append(args, "-b:v", videoBitrate)
    }
    args = append(args,
        "-c:a", "aac",
    )
    if audioBitrate != "" {
        args = append(args, "-b:a", audioBitrate)
    }
    // ... rest unchanged
}
```

**验收**：转码后视频码率符合 profile 配置（可通过 ffprobe 验证）

---

#### T3.7 Ent Schema 补充 BentoParameters

**文件**：`internal/data/entity/schema/encode_profile.go`（修改）

**新增字段**：
```go
field.Text("bento_parameters").Optional(), // Bento4 mp4hls extra args
```

**同步修改**：
- `svc-media/data/encode_profile_repo.go` — convertEncodeProfileToBiz 补充 BentoParameters
- `svc-media/data/seed.go` — 取消注释 `SetBentoParameters(p.BentoParams)`
- `svc-media/biz/upload.go` — ProcessMedia 调用 MP4HLS 时传入 profile.BentoParameters

**验收**：`ent generate` 无错误，seed 数据包含 bento_parameters，转码时传入 MP4HLS

---

#### T3.8 Preview Profile 特殊处理

**问题**：seed 中 `"preview"` profile 的 resolution 为 `"-"`，codec 为 `"-"`，extension 为 `"gif"`。
这不是视频转码，而是 GIF 预览生成。

**方案**：
- 在 TranscodeHandler 中检测 profile.Name == "preview"，跳过常规 TranscodeToMP4
- 可选：使用 ffmpeg 生成 GIF（`-vf "fps=10,scale=320:-1"`）
- 如果暂时不做 GIF 生成，可以在 worker 中跳过该 profile（设置 task status = "skipped"）

**验收**：preview profile 不会导致 ffmpeg 错误

---

### Phase 4：前端 HLS 播放

#### T3.9 HLSPlayer 组件

**文件**：`web/src/components/HLSPlayer.tsx`（新建）

**功能**：
- 引入 `hls.js`
- 检测浏览器原生 HLS 支持（iOS Safari）
- 加载 `master.m3u8` 并自适应播放
- 画质切换 UI（360p / 720p / 自动）
- 播放控制（播放/暂停/进度条/全屏/音量）
- 加载中 Skeleton 占位

**验收**：组件编译通过，可嵌入 Watch 页面

---

#### T3.10 播放页集成 HLSPlayer

**文件**：`web/src/pages/home/Watch.tsx`（修改）

**逻辑**：
```tsx
// 根据 media.encoding_status 决定播放器：
// - "success" + hls_file 非 → <HLSPlayer src={media.hls_file} />
// - "processing" → "转码中..." 状态 + 轮询进度（每 5s 查询 GET /api/v1/media/{id}/tasks）
// - "failed" → fallback 原始文件播放 <video src={media.url} />
// - "pending" → 等待转码
```

**验收**：上传视频 → 等待转码 → HLS 播放正常，画质可切换

---

#### T3.11 转码进度 API

**文件**：`internal/server/media.go`（修改）

**新增端点**：
```
GET /api/v1/media/:id/tasks
```
返回该媒体的所有 EncodingTask 列表（包含每个 profile 的 status/progress）。

**验收**：前端可轮询获取转码进度

---

### Phase 5：SSE 实时推送（可选）

#### T3.12 转码进度 SSE 推送

**文件**：`internal/server/media.go`（修改）

**新增端点**：
```
GET /api/v1/media/:id/tasks/stream
```

**实现**：复用 `MediaUseCase.Subscribe` 现有机制：
```go
// TranscodeHandler 在转码进度变更时调用：
uc.mediaUseCase.Publish(mediaID, &EncodingEvent{MediaId: mediaID, Task: task})

// SSE endpoint 订阅并推送给前端：
ch, cleanup := uc.mediaUseCase.Subscribe(ctx, mediaID)
defer cleanup()
for event := range ch {
    // write SSE: data: {json}\n\n
}
```

**验收**：前端通过 SSE 实时接收转码进度，无需轮询

---

## 依赖关系

```
T3.0 (Topic 定义) ──→ T3.1 (GoChannel 工厂) ──→ T3.2 (TranscodeHandler)
                                                       ↓
T3.3 (Router 启动) ──→ T3.4 (UploadUseCase 改造) ←──┘

T3.5 (Resolution 转换) ──→ T3.2 需要使用
T3.6 (Bitrate 透传)  ──→ T3.2 需要使用
T3.7 (BentoParameters) ──→ T3.2 需要使用
T3.8 (Preview 跳过)  ──→ T3.2 需要使用

T3.9 (HLSPlayer)  ──→ T3.10 (Watch 页集成)
T3.11 (进度 API)  ──→ T3.10 (Watch 页轮询)
T3.12 (SSE 推送)  ──→ 可选，替代轮询
```

## 建议开发顺序

1. **T3.0 + T3.1**：基础设施（GoChannel pubsub）
2. **T3.5 + T3.6 + T3.7 + T3.8**：Bugs 修复（先修这些，TranscodeHandler 才能正常工作）
3. **T3.2 + T3.3 + T3.4**：核心转码 pipeline（GoChannel → TranscodeHandler → UploadUseCase 改造）
4. **T3.11**：进度 API（方便前端调试）
5. **T3.9 + T3.10**：前端 HLS 播放
6. **T3.12**：SSE 推送（可选增强）

## 风险与缓解

| 风险 | 影响 | 缓解 |
|------|------|------|
| ffmpeg 进程过多导致 CPU/内存耗尽 | 系统卡顿 | semaphore.Weighted 限制并发，默认 maxWorkers=3 |
| 单个 ffmpeg 转码时间过长 | 后续任务排队 | 每个任务 2h 超时，worker 失败后标记 task failed |
| GoChannel buffer 满导致 publish 阻塞 | Upload handler 阻塞 | 设置合理的 buffer size + publish 带超时 context |
| Bento4 mp4hls 不兼容某些 MP4 | HLS 生成失败 | fallback 到 `hls.go` 的 `TranscodeToHLS`（ffmpeg 直出） |
| 前端 HLS.js 与 iOS Safari 兼容性 | 部分设备无法播放 | 检测原生 HLS 支持，优先使用原生播放器 |

---

*最后更新：2026-04-03*
