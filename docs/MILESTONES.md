# orig-cms 开发里程碑计划

> **制定时间**：2026-03-31  
> **基准状态**：见 [PROJECT_ANALYSIS.md](./PROJECT_ANALYSIS.md)  
> **快速入手**：新 session 请先读 [QUICK_START.md](./QUICK_START.md)（防重复劳动知识地图）
> **更新规则**：每完成一个任务请打勾 `[x]`，并在对应里程碑底部追加完成日期

---

## 总览

| 里程碑 | 主题 | 目标 | 预估周期 | 状态 |
|--------|------|------|----------|------|
| [M0](#m0-架构准备) | 架构准备 | 锁定服务边界、前端目录重构、废弃 svc-portal | 3 天 | ✅ 已完成 (2026-03-31) |
| [M1](#m1-基础闭环) | 基础闭环 | 单体模式跑通 + 用户认证 + 前端框架 | 2 周 | ✅ 已完成 (2026-04-03) |
| [M2](#m2-媒体上传与播放) | 媒体上传与播放 | 文件上传 + 基础视频播放（无转码） | 4 周 | ✅ 已完成 (2026-04-03) |
| [M3](#m3-视频转码与-hls) | 视频转码与 HLS | Watermill GoChannel 异步转码 + Bento4 HLS + 前端 HLS.js | 4 周 | 🔶 部分完成 (ffmpeg/Bento4 helpers + encode profile + ProcessMedia 逻辑存在; Watermill pubsub 集成待完成) |
| [M4](#m4-完整内容管理) | 完整内容管理 | 评论/收藏/频道/RBAC 权限 | 12 周 | 🔲 未开始 |
| [M5](#m5-生产就绪) | 生产就绪 | 监控/搜索/对象存储/可观测性 | 20 周 | 🔲 未开始 |

> **Architecture**: See [ARCHITECTURE.md](./ARCHITECTURE.md)
> **Key decisions**:
> - Monolith-first: `cmd/server` is the active entry point
> - Frontend split into three layers: `pages/auth` + `pages/home` (public) + `pages/admin`
> - Runtime integration: only config loading (`bootstrap.yaml`) + logger (`runtime/log`); Gin manual DI and routing kept as-is

> **偏差决策记录**（源自 2026-04-03 诊断报告，已决策项不再视为偏差）:
>
> | 偏差ID | 描述 | 决策 | 纳入 |
> |--------|------|------|------|
> | D1 | HTTP 框架 Gin vs Kratos | **保留 Gin**（M1/M2 已验证可行） | 已解决 |
> | D3 | 依赖注入 手动 vs Wire | **M1 手动 DI**（Gin 兼容），后续用 Wire | T1.4 |
> | D5 | 配置管理 环境变量 vs runtime bootstrap | **已迁移 runtime bootstrap** (2026-04-03) | 已解决 |
> | D6 | Storage 后端 LocalStorage vs S3 | **接口抽象，配置切换**（LocalStorage 默认） | M5 |
> | D7 | ent schema 位置 | **统一 `internal/data/entity/schema/`** | 已解决 |
> | D8 | handler 跳层 | **待审计确认** | M4 |
> | D9 | RBAC 权限 | **基础角色控制**（admin/user） | M4 (T4.3) |
> | D10 | svc-content 空 | **M4 实现** | M4 |
> | D4 | 转码 goroutine vs 消息队列 | **TranscodeWorker 接口抽象**（CE: goroutine pool + semaphore） | M3 |
> | D2 | M2 入口文件 | **已补全** `cmd/svc-user/portal/api-gateway/` staged | 已解决 |

---

## M0：架构准备

> **目标**：在写任何业务代码之前，先把架构基础打牢——明确服务边界，重组前端目录，废弃不合理的 svc-portal。  
> **预估时间**：3 天  
> **完成标准**：所有架构决策落地到代码结构，团队对服务边界达成共识，可以安心进入 M1 开发。

---

### T0.1 废弃 svc-portal，迁移聚合逻辑

**背景**：svc-portal 作为独立微服务存在价值有限（见 ARCHITECTURE.md §问题2），聚合逻辑合并到 svc-api-gateway。

**任务清单**：
- [x] 将 `internal/svc-portal/biz/portal.go` 中的聚合逻辑迁移到 `internal/gateway/handler/` 下 (2026-03-31)
  - `GetHomeFeed` → `gateway/handler/feed.go`
  - `GetVideoDetail` → `gateway/handler/detail.go`
  - `Search` → `gateway/handler/search.go`
  - `GetUserProfile` → `gateway/handler/profile.go`
- [x] 删除或归档 `cmd/svc-portal/` 目录 (2026-03-31)
- [x] 删除或归档 `internal/svc-portal/` 目录 (2026-03-31)
- [x] 删除 `api/proto/v1/portal/` proto 文件 (2026-03-31)
- [x] 更新 `go.mod` 删除 svc-portal 相关引用 (2026-03-31)
- [x] 运行 `go build ./...` 确认无编译错误 (2026-03-31)

**验收标准**：
- ✅ 不存在 `cmd/svc-portal/` 独立服务 (2026-03-31)
- ✅ 原 portal 聚合逻辑在 gateway 中可访问 (2026-03-31)
- ✅ `go build ./...` 无错误 (2026-03-31)

---

### T0.2 补全前端页面骨架

**背景**：当前 `pages/home/` 已有用户侧核心页面（index/Media/Search/Watch），`pages/auth/` 和 `pages/admin/` 也已存在，整体三层结构基本到位。T0.2 的工作是**补全缺失页面**，而不是重建目录。

**现状盘点**：
- ✅ `pages/home/index.tsx`：首页 Feed
- ✅ `pages/home/Watch.tsx`：媒体播放页
- ✅ `pages/home/Search.tsx`：搜索页
- ✅ `pages/home/Media.tsx`：媒体详情页
- ✅ `pages/auth/`：认证页（已有）
- ✅ `pages/admin/`：管理后台（已有）
- ✅ `pages/user/`：空壳已清理 (2026-03-31)

**任务清单**：
- [x] 删除 `pages/user/` 目录（空壳，无实际内容） (2026-03-31)
- [x] 在 `pages/home/` 补充缺失占位页面： (2026-03-31)
  - `Channel.tsx`（频道页）
  - `Profile.tsx`（用户公开主页）
  - `me/Upload.tsx`（上传页，登录用户）
  - `me/Favorites.tsx`（我的收藏）
  - `me/Notifications.tsx`（我的通知）
- [x] 创建 `web/src/router/index.tsx` 定义完整路由表（含占位组件） (2026-03-31)
- [x] 创建 `web/src/router/guards.tsx` 路由守卫基础结构 (2026-03-31)
- [x] 创建 `web/src/hooks/` 目录及空文件：`useAuth.ts`、`useMedia.ts` (2026-03-31)
- [x] 创建 `web/src/lib/api/` 目录，拆分 `lib/api.ts` 为 `auth.ts`、`media.ts`、`content.ts`、`user.ts` (2026-03-31)

**验收标准**：
- ✅ `pages/user/` 空壳已清除 (2026-03-31)
- ✅ `pages/home/` 包含用户侧全部页面（含占位） (2026-03-31)
- ✅ 所有路由已在 router/index.tsx 中定义 (2026-03-31)
- ✅ 前端 `bun run dev` 能正常启动，路由可访问 (2026-03-31)

---

### T0.3 明确 svc-api-gateway 架构

**背景**：gateway 需要承接路由转发 + 原 portal 聚合职责，需要明确其内部结构。

**任务清单**：
- [x] 创建 `internal/gateway/` 目录结构： (2026-03-31)
  ```
  internal/gateway/
  ├── middleware/     # JWT验证、限流、CORS、日志
  ├── handler/        # 聚合型 handler（feed/search/detail）
  ├── router/         # 路由注册
  └── client/         # 下游服务 gRPC 客户端
  ```
- [x] 创建各目录的占位 `.go` 文件 (2026-03-31)
- [x] 在 `cmd/svc-api-gateway/main.go` 中引用 gateway 包，完成基础框架（暂时只有健康检查路由） (2026-03-31)

**验收标准**：
- ✅ gateway 内部结构符合架构设计 (2026-03-31)
- ✅ `go build ./cmd/svc-api-gateway/...` 无编译错误 (2026-03-31)

---

### M0 整体验收检查清单

```
[x] svc-portal 代码已归档/删除 (2026-03-31)
[x] 前端目录结构已重构（auth/app/admin 三层） (2026-03-31)
[x] 完整路由表已建立 (2026-03-31)
[x] gateway 内部结构已搭建 (2026-03-31)
[x] go build ./... 零错误 (2026-03-31)
[x] bun run dev 前端正常启动 (2026-03-31)
[x] 所有路由可访问（哪怕是占位页面） (2026-03-31)
[ ] 团队对 ARCHITECTURE.md 中的决策无异议
```

---

## M1：基础闭环

> **目标**：在单体模式（`cmd/server`）下跑通全链路——用户注册/登录/JWT，前端完整认证流程，Admin 页面需登录才能访问。  
> **预估时间**：2 周  
> **完成标准**：用户可通过前端注册、登录并拿到 JWT Token，前端各路由可正常访问（哪怕内容简单），Admin 路由有权限守卫。

> **务实策略**：M1 优先跑通功能，微服务拆分不是 M1 目标。`cmd/server` 单体模式承载所有功能，M2 起再按边界拆分。

---

### T1.1 统一 ent entity 包

**问题背景**：`internal/data/entity`（完整字段）和 `internal/svc-media/data/entity`（精简版）并存，字段不同步。

**任务清单**：
- [x] 审计 `svc-media/data/entity` 与 `internal/data/entity` 的字段差异，记录 diff (2026-03-31)
- [x] 将 `svc-media/data` 层的导入路径统一改为 `internal/data/entity` (2026-03-31)
- [x] 修复因字段变更导致的编译错误 (2026-03-31)
- [x] 删除 `internal/svc-media/data/entity` 目录 (2026-03-31)
- [x] 迁移 `svc-user` 对私有实体的引用至统一实体 (2026-03-31)
- [x] 运行 `go build ./...` 确认无编译错误 (2026-03-31)

**测试方案**：
```bash
# 编译检查
go build ./...

# 单元测试（如有）
go test ./internal/svc-media/...
```

**验收标准**：
- ✅ 项目根目录 `go build ./...` 零错误 (2026-03-31)
- ✅ 只存在一个 ent entity 包：`internal/data/entity` (2026-03-31)
- ✅ svc-media / svc-user 所有数据操作使用统一 entity (2026-03-31)

---

### T1.2 完善单体模式入口（cmd/server）

**问题背景**：M1 阶段采用单体模式优先策略，`cmd/server` 应承载所有功能，各微服务 cmd 可以是空壳。

**任务清单**：

**单体入口（优先）**：
- [x] 完善 `cmd/server/main.go`，接入基础路由和数据库迁移 (2026-03-31)
- [x] 注册所有模块的 HTTP 路由 (user/media/content) (2026-03-31)
- [x] 注册健康检查端点 `GET /healthz` (2026-03-31)
- [x] 适配统一实体包 `internal/data/entity` (2026-03-31)

**svc-api-gateway（最小可用）**：
- [x] 实现基础 HTTP 路由框架 (2026-03-31)
- [x] 接入 JWT 验证中间件 (2026-03-31)
- [x] 实现健康检查端点 `GET /healthz` (2026-03-31)
- [x] 内置聚合 handler 框架 (gateway/handler/ 占位) (2026-03-31)

**其他微服务（保持存根即可）**：
- [x] `svc-media`：保持存根 ✅
- [x] `svc-content`：同上 ✅
- [x] `svc-user`：保持现有完整实现 ✅

**测试方案**：
```bash
# 启动单体服务
go run ./cmd/server/...

# 健康检查
curl http://localhost:9090/healthz
# 期望：{"status":"ok","version":"1.0.0"}

# 验证路由注册
curl http://localhost:9090/api/v1/users/
# 期望：200 (空列表) 或 401 (取决于是否开启强制认证)
```

**验收标准**：
- ✅ `cmd/server` 单体模式可正常启动 (2026-03-31)
- ✅ `/healthz` 返回 `{"status": "ok"}` (2026-03-31)
- ✅ 所有 API 路由可访问（认证、媒体、用户） (2026-03-31)
- ✅ 数据库连接正常（ent migrate 执行成功） (2026-03-31)

---

### T1.3 实现认证服务（JWT 登录） ✅ 已完成 (2026-04-03)

**问题背景**：系统无 JWT 发放/验证机制，所有 API 裸奔。

**实际实现**：M1 单体模式下，认证通过 Gin handler 直接实现（非 gRPC proto），功能完全到位。

**后端（`internal/server/auth.go` + `internal/auth/jwt.go`）**：
- [x] JWT 签发/验证（`auth.Manager`，HS256，可配置 TTL）
- [x] `POST /api/v1/auth/signin` 登录接口
- [x] `POST /api/v1/auth/signup` 注册接口（首个用户自动设为 admin）
- [x] `POST /api/v1/auth/signout` 登出接口（stateless）
- [x] `GET /api/v1/auth/me` 获取当前用户（需 JWT）
- [x] JWT 中间件 `JWTMiddleware`（Bearer token 解析 + claims 注入）
- [x] 密码哈希/验证（bcrypt，`svc-user/biz/user.go`）

**前端**：
- [x] `pages/auth/SignIn/index.tsx` 登录页（shadcn/ui）
- [x] `pages/auth/SignUp/index.tsx` 注册页
- [x] `lib/auth.ts` 登录/注册/登出 API
- [x] `hooks/useAuth.ts` 认证状态管理（localStorage 持久化）
- [x] `lib/request.ts` Authorization header 注入（`setAuth`）
- [x] 路由守卫 `requireAuth` / `requireAdmin` / `redirectIfAuth`（`router/index.tsx`）

**验收标准**：
- ✅ POST `/api/v1/auth/signin` 返回有效 JWT Token + 用户信息
- ✅ Token 可访问 `/api/v1/auth/me`
- ✅ 无效 Token 返回 401
- ✅ 前端注册/登录正常跑通，admin 用户验证通过
- ✅ 未登录访问 `/admin` 自动跳转到 `/auth/signin`

> **备注**：M1 采用 Gin handler 而非 gRPC proto。gRPC 认证接口留到 M2 微服务拆分时通过 proto 定义。`RefreshToken` 留作后续任务。

---

### T1.4 完善模块化 Wire 依赖注入

**问题背景**：只有 svc-user 有完整 wire setup，其余 svc 模块无 Wire 支持。各 svc 应独立组装依赖，`cmd/server` 单体入口汇总所有 svc。

> **说明**：Wire DI 是未来微服务拆分的基础。当前 `cmd/server` 用 Wire 组装所有 svc 到一个进程，微服务模式下各 `cmd/svc-*` 用 Wire 各自独立组装。

**任务清单**：
- [ ] 为 `svc-media` 创建 `wire.go` 和 `wire_gen.go`（参照 svc-user 模式）
- [ ] 为 `gateway` 创建 Wire setup（聚合层需要注入下游 svc 客户端）
- [ ] 为 `cmd/server` 单体入口创建 Wire setup，汇总所有 svc 依赖
- [ ] 运行 `wire gen` 生成依赖注入代码
- [ ] 删除手工 `New()` 调用，全部通过 Wire 注入

**测试方案**：
```bash
# 生成 Wire 代码
wire gen ./cmd/svc-user/...
wire gen ./cmd/svc-media/...
wire gen ./cmd/server/...

# 验证生成代码无编译错误
go build ./...
```

**验收标准**：
- ✅ `wire gen` 命令对所有模块无报错
- ✅ `go build ./...` 零错误
- ✅ 各模块依赖关系通过 Wire 自动组装
- ✅ `cmd/server` 单体入口通过 Wire 正常启动

---

### T1.5 修复代码质量问题

**任务清单**：
- [x] 修复 C-01：将 proto 文件中 `Page_size`、`Order_by` 改为 `page_size`、`order_by`（proto3 snake_case，生成后为 camelCase）
- [x] 修复 C-02：`svc-portal/biz/portal.go` 统一使用 `req.GetXxx()` 方式访问字段
- [x] 修复 C-05：`service/user.go` 中将局部变量 `user` 重命名为 `userInfo` 避免与包名冲突
- [x] 修复 C-04：前端删除所有 `alert()`/`confirm()`，替换为 shadcn `Toast` 和 `AlertDialog`

**测试方案**：
```bash
go vet ./...
go build ./...
```

**验收标准**：
- ✅ `go vet ./...` 无警告
- ✅ proto 字段遵循命名规范
- ✅ 前端无 `alert()` 调用

---

### T1.6 集成 shadcn/ui 组件库 ✅ 已完成 (2026-04-03)

**问题背景**：前端页面使用原生 HTML/Basic React，未统一 UI 组件库。

**实际实现**：shadcn/ui 已初始化，基础组件已安装，核心页面已使用 shadcn 组件。

**初始化 shadcn/ui**：
- [x] `web/components.json` 存在，shadcn/ui 已初始化
- [x] Tailwind CSS 和主题已配置
- [x] 已安装 17 个组件：Button, Input, Card, Dialog, AlertDialog, Table, Tabs, Dropdown, Select, Badge, Avatar, Progress, ScrollArea, Skeleton, Separator, Label, Textarea

**已使用 shadcn 的页面**：
- [x] `pages/auth/SignIn/index.tsx` - Card + Input + Button
- [x] `pages/auth/SignUp/index.tsx` - Card + Input + Button
- [x] 其他管理后台页面按需使用中

> **备注**：后续新页面继续使用 shadcn 组件即可，不需要专项重构。

---

### M1 整体验收检查清单

```
[x] go build ./... 零错误
[x] cmd/server 单体模式可正常启动
[x] /healthz 返回 200
[x] 数据库 ent migrate 执行成功
[x] POST /api/v1/auth/signin 返回 JWT Token
[x] POST /api/v1/auth/signup 注册用户
[x] GET /api/v1/auth/me 受保护接口正常
[x] 前端登录/注册页面正常工作
[x] 路由守卫正常（未登录跳转 /auth/signin）
[x] shadcn/ui 已集成
[x] runtime config loading + logger 集成 (2026-04-03)
[x] bootstrap.yaml 配置驱动 (2026-04-03)
[ ] T1.4 Wire DI 完善（M2 准备工作，非阻塞）
```

---

## M2：媒体上传与播放 ✅ 已完成 (2026-04-03)

> **目标**：用户能上传视频/图片文件，系统存储后可列表展示并基础播放（无转码，直接原始文件）。  
> **实际实现**：Gin handler 层直接实现（非 gRPC proto），`internal/server/upload.go` + `internal/server/media.go`。用户已手动测试通过。

---

### T2.1 文件上传 API ✅

> 实际实现：`server/upload.go` + `svc-media/biz/upload.go` + `svc-media/data/upload_repo.go`，支持分块上传、文件类型校验、本地存储。

### T2.2 媒体列表与详情 API ✅

> 实际实现：`server/media.go`，CRUD + 静态文件服务 + 播放量统计。

### T2.3 前端上传与播放功能 ✅

> 实际实现：`components/MediaUpload.tsx`、Admin Media 页面、播放页 `pages/home/Watch.tsx`。用户已测试通过。

### T2.4 缩略图生成（基础版） ✅

> 实际实现：视频 ffmpeg 截帧 + 图片 imaging 缩放，缩略图存储 + 静态服务。

---

### M2 整体验收检查清单

```
[x] 用户可通过前端上传 MP4 视频（< 500MB）
[x] 用户可通过前端上传 JPG/PNG 图片
[x] 上传文件持久化到磁盘
[x] 媒体列表正确展示所有上传内容
[x] 点击视频可直接在浏览器播放（HTML5 player）
[x] 播放量统计正常工作
[x] 缩略图自动生成
[x] 删除媒体同时删除物理文件
```

---

## M3：视频转码与 HLS

> **目标**：上传的视频经异步转码，生成多分辨率 HLS 流，前端使用 HLS.js 实现自适应码率播放。
> **预估时间**：M2 完成后再 4 周（累计 8 周）
> **完成标准**：上传 MP4 → 后台转码 → 生成 m3u8 → 前端自适应播放，encoding_status 状态机完整。
>
> **偏差 D4 策略（已更新 2026-04-03）**：使用 **Watermill GoChannel** 作为进程内 PubSub，替代原始 `go ProcessMedia()` 直接调用。
> 转码执行层定义 `TranscodeWorker` 接口，CE 使用 `goroutineWorker`（goroutine pool + `semaphore.Weighted` 控制并发）。
> 参考 backend 的 pubsub 模式：`svc-user/service/user.go` 注入 `message.Publisher` 发布事件。
>
> **消息流**：
> ```
> CompleteMultipartUpload → publisher.Publish("media.encode.request", msg)
>                         ↓
>               [Watermill GoChannel]
>                         ↓
>               TranscodeHandler.Handle(msg)
>                         ↓
>               TranscodeWorker.Submit(job) × N profiles
>                         ↓
>               goroutine pool (semaphore limits concurrent ffmpeg)
>                         ↓
>               ffmpeg TranscodeToMP4 → Bento4 MP4HLS → 更新 DB 状态
>                         ↓
>               publisher.Publish("media.encode.progress", event)
>                         ↓
>               SSE 推送到前端（复用 MediaUseCase.Subscribe 现有机制）
> ```
>
> **现有代码审计**（2026-04-03）：
> - ✅ `internal/helpers/ffmpeg/` — ffprobe/ffmpeg/Bento4 调用封装完整
> - ✅ `svc-media/biz/media.go` — EncodeProfile/EncodingTask/EncodingEvent 定义 + Subscribe/Publish
> - ✅ `svc-media/biz/upload.go` — ProcessMedia 完整四步流程（缩略图 → Profiles → TranscodeToMP4 → Bento4 MP4HLS）
> - ✅ `svc-media/data/encoding_task_repo.go` — EncodingTask CRUD
> - ✅ `svc-media/data/encode_profile_repo.go` — EncodeProfile CRUD
> - ✅ `svc-media/data/seed.go` — 22 个预置 profile（h264/h265 240p-1080p）
> - ✅ `internal/pubsub/pubsub.go` — Topic 常量定义框架（仅 user 事件，需添加 media 事件）
> - ✅ `internal/helpers/providers/common.go` — Watermill Publisher Provider（从 runtime container 获取，优雅降级返回 nil）
> - ✅ `internal/svc-user/service/user.go` — Watermill Publisher 注入示例
> - ✅ `cmd/server/main.go` — `/hls` 静态路由已挂载
> - ✅ `go.mod` — `watermill v1.5.1` 已依赖
> - ⚠️ `ProcessMedia` 当前用 `go ProcessMedia(context.Background(), ...)` 触发，无 Watermill
> - ⚠️ `MediaUseCase.Subscribe/Publish` 是纯内存 channel，与 Watermill 并行
> - ⚠️ `TranscodeToMP4` 的 resolution 参数格式与 seed 数据不一致（seed 用 "720"，ffmpeg 需要 "1280x720"）
> - ⚠️ `EncodeProfile` biz 有 `BentoParameters` 字段，但 ent schema 和 seed 中未包含
> - ⚠️ `TranscodeToMP4` 忽略了 `video_bitrate` / `audio_bitrate` 字段（硬编码使用默认值）

---

### T3.1 ffmpeg 转码服务

**任务清单**：
- [ ] 封装 ffmpeg 命令调用工具库（`internal/helpers/ffmpeg/`）
  - `ExtractInfo(filePath)` → 视频时长/分辨率/码率（ffprobe）
  - `Transcode(input, output, profile)` → 转码为指定 profile
  - 支持转码 profile：360p / 720p / 1080p（可按原始分辨率决定）
- [ ] 定义转码任务 Pub/Sub topic：`media.encode.request`
- [ ] 实现异步转码 Worker（订阅 `media.encode.request`）：
  1. 拉取任务
  2. 更新 `encoding_status = processing`
  3. 调用 ffmpeg 转码
  4. 更新 `encoding_status = success/fail`
  5. 发布 `media.encode.completed` 事件
- [ ] 转码输出目录：`/data/media/encoded/{media_id}/`
  - `360p.m3u8`, `360p_*.ts`
  - `720p.m3u8`, `720p_*.ts`
  - `master.m3u8`（多码率 master playlist）

**测试方案**：
```bash
# 手动触发转码（发布 Pub/Sub 消息）
# 或上传一个新视频（M3 后上传自动触发转码）

# 检查转码状态
curl http://localhost:9090/api/v1/media/$MEDIA_ID | jq '.encoding_status'
# 预期变化：pending → processing → success

# 检查文件生成
ls /data/media/encoded/$MEDIA_ID/
# 期望：master.m3u8, 360p.m3u8, 720p.m3u8, *.ts 文件

# 测试 m3u8 可访问
curl http://localhost:9090/media/encoded/$MEDIA_ID/master.m3u8
# 期望：返回 HLS playlist 内容
```

**验收标准**：
- ✅ 上传 MP4 后自动触发转码任务
- ✅ encoding_status 状态机：`pending → processing → success/failed`
- ✅ 生成 master.m3u8 及各分辨率 m3u8/ts 文件
- ✅ 转码失败时状态更新为 `failed`，不影响系统稳定性
- ✅ ffprobe 正确提取视频元信息

---

### T3.2 HLS 播放前端

**任务清单**：
- [ ] 引入 `hls.js` 依赖（`bun add hls.js`）
- [ ] 实现 `HLSPlayer` 组件（`web/src/components/HLSPlayer.tsx`）
  - 自动检测 HLS 支持（iOS 原生 / hls.js fallback）
  - 支持品质切换（360p/720p/自动）
  - 播放/暂停/进度条/全屏/音量控制
  - 加载中 Skeleton 占位
- [ ] 媒体播放页替换 HTML5 原生 video 为 HLSPlayer
- [ ] 当 encoding_status 为 `processing` 时显示"转码中"状态
- [ ] 当 encoding_status 为 `failed` 时显示原始文件 fallback 播放器

**测试方案**：
1. 上传 MP4 视频
2. 等待转码完成（通过 encoding_status 轮询或 WebSocket）
3. 进入播放页，HLSPlayer 正常加载 master.m3u8
4. 切换画质（360p/720p）正常工作
5. 在 iOS Safari 中验证原生 HLS 播放

**验收标准**：
- ✅ 转码完成后使用 HLS 播放（不再使用原始 MP4）
- ✅ 支持画质手动切换
- ✅ 支持自适应码率（网络差自动降码率）
- ✅ 转码中状态友好提示
- ✅ iOS Safari 原生 HLS 正常播放

---

### T3.3 编码状态实时推送（可选 WebSocket）

**任务清单**：
- [ ] 在 svc-media 添加 WebSocket 端点（`/ws/media/:id/status`）
- [ ] 转码状态变更时推送 WebSocket 消息给订阅客户端
- [ ] 前端播放页订阅 WebSocket，实时更新 encoding_status

**验收标准**：
- ✅ 上传后前端自动感知转码进度，无需手动刷新

---

### M3 整体验收检查清单

```
[ ] 上传 MP4 后自动触发转码
[ ] encoding_status 状态机正确流转
[ ] 生成 master.m3u8 和多分辨率 HLS 文件
[ ] 前端使用 HLS.js 自适应播放
[ ] 支持手动切换画质
[ ] 转码中/失败状态有友好 UI 提示
[ ] iOS Safari 可播放
```

---

## M4：完整内容管理

> **目标**：实现评论、收藏、点赞、频道、播放列表、标签等内容管理功能，完善 RBAC 权限体系。  
> **预估时间**：M3 完成后再 4 周（累计 12 周）  
> **完成标准**：用户可评论/收藏/点赞，管理员可管理分类/频道，RBAC 控制不同角色的操作权限。
>
> **偏差策略**：
> - D8（handler 跳层）：M4 阶段审计并修复所有直接操作 ent client 的 handler，统一走 biz 层
> - D9（RBAC 权限）：实现基础角色控制（admin/user），JWT claims + 中间件
> - D10（svc-content 空）：M4 实现 svc-content 服务

---

### T4.1 svc-content 服务实现

**任务清单**：
- [ ] 实现评论系统（`Comment` entity 已有 schema）
  - `CreateComment`：创建评论（支持嵌套回复）
  - `ListComments`：按 media_id 分页获取
  - `DeleteComment`：用户删除自己的评论 / 管理员删除任意评论
- [ ] 实现通知系统（`Notification` entity 已有 schema）
  - 评论/点赞/关注触发通知
  - `ListNotifications`：获取当前用户通知
  - `MarkAsRead`：标记已读
- [ ] 实现收藏/点赞（`Favorite`、`Like` entity 已有 schema）
  - `ToggleFavorite`：收藏/取消收藏
  - `ToggleLike`：点赞/取消点赞
  - 媒体详情中返回点赞数/收藏数

**测试方案**：
```bash
# 评论
curl -X POST http://localhost:9090/api/v1/media/$MEDIA_ID/comments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"text":"很好的视频！"}'

# 获取评论列表
curl http://localhost:9090/api/v1/media/$MEDIA_ID/comments

# 点赞
curl -X POST http://localhost:9090/api/v1/media/$MEDIA_ID/like \
  -H "Authorization: Bearer $TOKEN"

# 获取通知
curl http://localhost:9090/api/v1/notifications \
  -H "Authorization: Bearer $TOKEN"
```

**验收标准**：
- ✅ 评论 CRUD 正常
- ✅ 支持嵌套回复（parent_id 字段）
- ✅ 点赞/收藏正确统计
- ✅ 创建评论/点赞后触发通知

---

### T4.2 频道与播放列表

**任务清单**：
- [ ] 频道（Channel）CRUD API
  - 用户创建/管理自己的频道
  - 频道关联多个 Media
  - 频道订阅功能
- [ ] 播放列表（Playlist）CRUD API
  - 创建/编辑/删除播放列表
  - 向播放列表添加/移除媒体
  - 播放列表顺序管理
- [ ] 标签（Tag）管理
  - 管理员管理标签
  - 媒体关联标签（多对多）
  - 按标签筛选媒体

**验收标准**：
- ✅ 用户可创建频道并关联视频
- ✅ 用户可创建播放列表并管理顺序
- ✅ 按标签过滤媒体正常工作

---

### T4.3 RBAC 权限体系

**任务清单**：
- [ ] 在 User entity 添加 `role` 字段（`admin` / `user`，默认 `user`）
- [ ] JWT Token 中携带 role
- [ ] Gin 中间件校验 admin 路由的 role 权限
- [ ] 首个注册用户自动设为 admin（已有逻辑，确认 role 字段同步）
- [ ] 前端根据 role 显示/隐藏管理入口
- [ ] 定义角色：`admin` / `user`（可扩展 `editor`）
- [ ] 权限规则：admin 全权限，user 浏览/评论/收藏/上传（可配置关闭）

**验收标准**：
- ✅ 普通用户无法访问 Admin 管理接口
- ✅ Admin 可管理所有用户/媒体
- ✅ 前端菜单按角色动态渲染

---

### T4.4 前端内容管理完善

**任务清单**：
- [ ] 媒体播放页添加评论组件
- [ ] 媒体播放页添加点赞/收藏按钮
- [ ] 实现通知中心（Bell 图标 + 下拉列表）
- [ ] 实现频道页面（`/channel/:id`）
- [ ] 实现播放列表页面（`/playlist/:id`）
- [ ] 实现用户主页（`/user/:username`）：展示用户上传的视频
- [ ] Admin 管理后台完善：分类管理、标签管理、用户角色管理

**验收标准**：
- ✅ 评论可正常发布/显示
- ✅ 点赞/收藏状态正确同步
- ✅ 通知实时展示
- ✅ Admin 可管理所有实体

---

### M4 整体验收检查清单

```
[ ] 评论 CRUD 完整（含嵌套回复）
[ ] 点赞/收藏功能正常
[ ] 通知系统正常推送和展示
[ ] 频道管理完整
[ ] 播放列表管理完整
[ ] 标签管理及过滤正常
[ ] RBAC 权限控制生效（前后端均验证）
[ ] 前端内容管理页面完善
```

---

## M5：生产就绪

> **目标**：系统具备生产部署条件，支持对象存储抽象、全文搜索、监控可观测性、限流、链路追踪。  
> **预估时间**：M4 完成后再 8 周（累计 20 周）
>
> **偏差 D6 策略**：存储层做接口抽象（`StorageBackend`），默认 LocalStorage，通过配置切换后端。

---

### T5.1 对象存储抽象层

**任务清单**：
- [ ] 定义 `StorageBackend` 接口（`Upload` / `Download` / `Delete` / `GetURL`）
- [ ] 实现 `LocalStorageBackend`（封装当前 `svc-media/data/local_storage.go`）
- [ ] 通过 `bootstrap.yaml` 配置切换后端
- [ ] 媒体文件、缩略图、HLS 切片统一走 StorageBackend 接口
- [ ] 支持预签名 URL 生成（有效期可配置）

**验收标准**：
- ✅ 所有文件操作通过 StorageBackend 接口完成
- ✅ 配置切换后端不影响业务逻辑
- ✅ 媒体 URL 正确返回

---

### T5.2 全文搜索（MeiliSearch）

**任务清单**：
- [ ] 集成 MeiliSearch 或 Elasticsearch
- [ ] 媒体创建/更新时同步索引（title/description/tags）
- [ ] 搜索 API 替换数据库 LIKE 查询为全文搜索
- [ ] 支持中文分词

**验收标准**：
- ✅ 搜索"测试视频"能找到包含该词的所有媒体
- ✅ 搜索响应时间 < 100ms（1万条数据）
- ✅ 支持模糊搜索和中文搜索

---

### T5.3 链路追踪与监控

**任务清单**：
- [ ] 集成 OpenTelemetry（参考 origadmin/runtime 是否已支持）
- [ ] 所有微服务导出 Trace 到 Jaeger
- [ ] 接入 Prometheus metrics（请求数、延迟、错误率）
- [ ] 配置 Grafana Dashboard
- [ ] 健康检查接口增加详细信息（DB 连接状态、消息队列状态）

**验收标准**：
- ✅ Jaeger UI 可看到完整请求链路
- ✅ Grafana 展示各服务 QPS/P99 延迟
- ✅ 服务异常时 Prometheus alerting 触发

---

### T5.4 API 限流与安全加固

**任务清单**：
- [ ] 在 svc-api-gateway 实现请求限流（令牌桶算法）
  - 全局限流：10000 req/min
  - 用户级限流：上传接口 10 次/min
- [ ] CORS 配置完善
- [ ] 上传文件病毒扫描（ClamAV 集成，可选）
- [ ] 请求日志审计

**验收标准**：
- ✅ 超过限流阈值返回 429 Too Many Requests
- ✅ 跨域配置正确，仅允许白名单域名

---

### T5.5 Docker Compose 生产部署

**任务清单**：
- [ ] 完善 `docker-compose.yml`（所有微服务 + PostgreSQL + Redis + Consul + MinIO + MeiliSearch）
- [ ] 每个服务创建生产 Dockerfile（多阶段构建，最终镜像 < 50MB）
- [ ] 完善 `Makefile`：`make build-all`、`make deploy`、`make migrate`
- [ ] 数据库迁移脚本（`make migrate` 运行 ent migrate）
- [ ] 编写 `docs/DEPLOYMENT.md` 部署手册

**验收标准**：
- ✅ `docker compose up -d` 一键启动全部服务
- ✅ 首次启动自动执行数据库迁移
- ✅ 所有服务镜像 < 50MB
- ✅ 部署手册清晰，包含环境变量说明

---

### M5 整体验收检查清单

```
[ ] 对象存储切换 MinIO 正常工作
[ ] 全文搜索响应时间 < 100ms
[ ] Jaeger/Grafana 监控可视化
[ ] 限流机制生效
[ ] docker compose up -d 一键部署成功
[ ] 数据库迁移脚本自动执行
```

---

## 附录：功能缺失追踪

> **来源**：2026-04-03 诊断报告 §4.1 功能完整性分析
> **目的**：追踪原始 MediaCMS 功能 vs 当前实现的差距，确保不遗漏。
> **规则**：状态更新请同步到对应 milestone 的任务清单。

### 功能对照表

| 功能 | MediaCMS 原始 | 当前实现 | 差距 | 计划纳入 | 优先级 |
|------|-------------|---------|------|---------|--------|
| **用户注册/登录** | Email + Social + SAML | ✅ JWT 邮箱登录 | 缺 Social/SAML | M4+ (Social) | P2 |
| **用户 Profile** | 名称/描述/头像/位置 | ⚠️ 基础 CRUD | 缺头像上传/位置 | M4 | P1 |
| **频道** | 每用户一个 | ✅ ent schema | 缺 API | M4 (T4.2) | P0 |
| **订阅** | 用户间订阅 | ❌ 无 | 缺 subscription 表 | M4+ | P1 |
| **媒体上传** | fine-uploader 分片 | ✅ 完整分片上传 | — | 已完成 | — |
| **媒体类型** | video/audio/image/pdf | ⚠️ video/image/audio | 缺 PDF | M4+ | P2 |
| **转码引擎** | Celery+ffmpeg+Bento4 | ⚠️ goroutine+ffmpeg+Bento4 | 缺 Watermill pubsub 集成、TranscodeWorker 接口抽象、并发控制 | M3 (D4) | P0 |
| **HLS 流媒体** | m3u8 多码率 | ✅ Bento4 MP4HLS | — | M3 验证 | P0 |
| **缩略图** | 自动提取 | ✅ ffmpeg | — | 已完成 | — |
| **Sprite 预览** | 时间轴预览 | ❌ 无 | 后端无生成逻辑 | M5 | P3 |
| **字幕** | WebVTT + Whisper | ❌ 无 | 缺 schema | M4+ | P2 |
| **视频章节** | 前端编辑器 | ❌ 无 | 缺 schema | M5 | P3 |
| **视频剪辑** | 裁剪功能 | ❌ 无 | — | 不计划 | — |
| **分类** | 全局/用户级 | ⚠️ 基础 CRUD | 缺用户级权限 | M4 (T4.2) | P1 |
| **标签** | 关键词标签 | ⚠️ 基础 CRUD | — | M4 (T4.2) | P1 |
| **播放列表** | 创建/编辑/排序 | ⚠️ 基础 CRUD | 缺拖拽排序 | M4 (T4.2) | P1 |
| **评论** | 多级回复 | ⚠️ Schema 有无实现 | 缺 API | M4 (T4.1) | P0 |
| **点赞/踩** | 用户交互 | ❌ Proto 有无实现 | 缺 API | M4 (T4.1) | P0 |
| **收藏** | 用户收藏 | ❌ Proto 有无实现 | 缺 API | M4 (T4.1) | P0 |
| **搜索** | 关键词搜索 | ⚠️ 基础搜索 | 缺全文搜索/ES | M5 (T5.2) | P1 |
| **Feed** | 最新/精选/推荐 | ⚠️ 基础实现 | 缺推荐算法 | M4+ | P2 |
| **SSO/SAML** | Identity Provider | ❌ 无 | — | 不计划 | P3 |
| **RBAC 权限** | 多层权限 | ❌ 无实现 | 缺基础角色控制 | M4 (T4.3) | P0 |
| **管理后台** | 管理界面 | ⚠️ 部分 | 缺管理专用 API | M4 (T4.4) | P1 |
| **任务监控** | Celery 任务列表 | ⚠️ SSE 进度 | 功能有限 | M3 (T3.3) | P1 |
| **内容审核** | 上报/审核 | ❌ 无 | — | M4+ | P2 |
| **RSS Feed** | 输出 RSS | ❌ 无 | — | M5 | P3 |
| **Embed 播放** | 外部嵌入 | ❌ 无 | — | M5 | P3 |

### 按优先级排序的开发计划

**P0（当前 milestone 必须完成）**：
- M3：转码端到端验证（D4）+ HLS 播放验证
- M4：评论 API、点赞/收藏 API、频道 API、基础 RBAC

**P1（下个 milestone 追踪）**：
- M4：用户 Profile 完善（头像/位置）、分类/标签/播放列表完善、订阅功能、管理后台 API
- M5：全文搜索

**P2（后续规划）**：
- Social Login、字幕支持、Feed 推荐、内容审核、PDF 支持

**P3（远期/不计划）**：
- Sprite 预览、视频章节、视频剪辑、SSO/SAML、RSS、Embed 播放

---

## 附录：技术规范

### 代码提交规范

```
feat(svc-user): 实现 JWT 登录接口
fix(svc-media): 修复双 entity 包编译错误
docs: 更新 MILESTONES.md M1 进度
test(svc-user): 添加 Login biz 单元测试
refactor(svc-portal): 补充 Wire 依赖注入
```

### 分支策略

```
main          ← 生产分支，仅合并经过测试的代码
develop       ← 开发集成分支
feature/m1-*  ← M1 各任务分支
feature/m2-*  ← M2 各任务分支
```

### 接口版本规范

- 所有 REST 接口前缀：`/api/v1/`
- gRPC 包名：`origcms.v1.{service}`
- 响应格式统一：`{"data": {...}, "code": 0, "message": "ok"}`
- 错误响应：`{"code": 40001, "message": "用户名已存在", "data": null}`

### 测试覆盖率目标

| 层级 | 目标覆盖率 |
|------|-----------|
| biz 层（业务逻辑） | ≥ 80% |
| service 层（接口处理） | ≥ 60% |
| data 层（数据访问） | ≥ 40% |
| 前端组件 | ≥ 50% |

---

*最后更新：2026-04-03 | 请在每完成一个任务后更新对应 `[ ]` 为 `[x]` 并注明完成日期*
