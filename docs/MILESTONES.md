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
| [M1](#m1-基础闭环) | 基础闭环 | 单体模式跑通 + 用户认证 + 前端框架 | 2 周 | 🔶 部分完成 (T1.1/T1.2/T1.5 done; T1.3/T1.4/T1.6 pending) |
| [M2](#m2-媒体上传与播放) | 媒体上传与播放 | 文件上传 + 基础视频播放（无转码） | 4 周 | 🔶 部分完成 (upload plan + transcoding helpers exist; full upload/ playback flow not integrated) |
| [M3](#m3-视频转码与-hls) | 视频转码与 HLS | 异步转码 + HLS 流媒体播放 | 8 周 | 🔶 部分完成 (ffmpeg helpers + HLS logic + encode profile exist; async worker pipeline not complete) |
| [M4](#m4-完整内容管理) | 完整内容管理 | 评论/收藏/频道/RBAC 权限 | 12 周 | 🔲 未开始 |
| [M5](#m5-生产就绪) | 生产就绪 | 监控/搜索/对象存储/可观测性 | 20 周 | 🔲 未开始 |

> **Architecture**: See [ARCHITECTURE.md](./ARCHITECTURE.md)
> **Key decisions**:
> - Monolith-first: `cmd/server` is the active entry point
> - Frontend split into three layers: `pages/auth` + `pages/home` (public) + `pages/admin`

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

### T1.3 实现认证服务（JWT 登录）

**问题背景**：系统无 JWT 发放/验证机制，所有 API 裸奔。

**任务清单**：

**后端（嵌入 svc-user）**：
- [ ] 添加 `Login` RPC 到 user proto（`LoginRequest{username, password}` → `LoginResponse{access_token, refresh_token, expires_at}`）
- [ ] 运行 `buf generate` 重新生成 proto 代码
- [ ] 在 `svc-user/biz/user.go` 实现 `Login` 逻辑：验证密码 → 签发 JWT
- [ ] 在 `svc-user/service/user.go` 暴露 `Login` gRPC handler
- [ ] 添加 JWT 验证中间件（参考 origadmin/contrib/security）
- [ ] 在需要认证的 gRPC 方法上加 JWT 拦截器
- [ ] 添加 `RefreshToken` RPC（可选，M1 可后补）

**前端**：
- [ ] 新建 `web/src/pages/auth/Login.tsx` 登录页面
- [ ] 新建 `web/src/pages/auth/Register.tsx` 注册页面
- [ ] 在 `lib/api.ts` 添加 `login()`、`register()` 方法
- [ ] 实现 `AuthContext`（存储 token，提供 `useAuth()` hook）
- [ ] 为所有 API 请求添加 `Authorization: Bearer <token>` header
- [ ] 添加路由守卫：未登录跳转到 `/login`
- [ ] 替换 `alert()` 错误提示为 shadcn `Toast`

**测试方案**：
```bash
# 注册用户
curl -X POST http://localhost:9091/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"Test1234!"}'

# 登录获取 Token
curl -X POST http://localhost:9091/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"Test1234!"}'
# 期望响应：{"access_token":"eyJ...","expires_at":"..."}

# 使用 Token 访问受保护接口
curl http://localhost:9091/api/v1/users/me \
  -H "Authorization: Bearer eyJ..."
# 期望响应：用户信息 JSON

# 使用过期/无效 Token
curl http://localhost:9091/api/v1/users/me \
  -H "Authorization: Bearer invalid"
# 期望响应：401 Unauthorized
```

**验收标准**：
- ✅ POST `/api/v1/users/login` 返回有效 JWT Token
- ✅ Token 可以访问受保护接口
- ✅ 无效 Token 返回 401
- ✅ 前端登录页可正常提交，登录成功后跳转到 Dashboard
- ✅ 未登录访问 `/admin` 自动跳转到 `/login`

---

### T1.4 完善 Wire 依赖注入

**问题背景**：只有 svc-user 有完整 wire setup，其余服务无法通过 Wire 正确组装。

**任务清单**：
- [ ] 为 `svc-media` 创建 `wire.go` 和 `wire_gen.go`（参照 svc-user）
- [ ] 为 `svc-portal` 创建 `wire.go` 和 `wire_gen.go`
- [ ] 实现 `svc-portal/data/data.go` 中 gRPC 客户端工厂（NewMediaClient、NewUserClient）
- [ ] 运行 `wire gen ./cmd/svc-media/...` 等生成依赖注入代码
- [ ] 所有服务通过 Wire 组装，删除手工 `New()` 调用

**测试方案**：
```bash
# 生成 Wire 代码
wire gen ./cmd/svc-user/...
wire gen ./cmd/svc-media/...
wire gen ./cmd/svc-portal/...

# 验证生成代码无编译错误
go build ./...
```

**验收标准**：
- ✅ `wire gen` 命令对所有服务无报错
- ✅ `go build ./...` 零错误
- ✅ 各服务依赖关系通过 Wire 自动组装

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

### T1.6 集成 shadcn/ui 组件库

**问题背景**：前端页面使用原生 HTML/Basic React，未统一 UI 组件库，导致：
- 视觉风格不统一
- 开发效率低
- 代码复用性差

**任务清单**：

**初始化 shadcn/ui**：
- [ ] 在 `web/` 目录初始化 shadcn/ui（`npx shadcn@latest init`）
- [ ] 配置 Tailwind CSS 和主题
- [ ] 安装基础组件：Button, Input, Card, Dialog, Dropdown, Table 等

**重构现有页面**：
- [ ] 重构 `pages/auth/SignIn/index.tsx` - 使用 shadcn Input/Button/Card
- [ ] 重构 `pages/auth/SignUp/index.tsx` - 使用 shadcn Input/Button/Card
- [ ] 重构 `pages/admin/` 管理后台页面 - 使用 shadcn Table/Card/Button
- [ ] 重构 `pages/home/` 用户侧页面 - 使用 shadcn 组件

**测试方案**：
```bash
# 启动前端
cd web && npm run dev

# 验证各页面正常渲染
# - 登录页 http://localhost:5173/signin
# - 注册页 http://localhost:5173/signup
# - 首页 http://localhost:5173/
# - 管理后台 http://localhost:5173/admin
```

**验收标准**：
- ✅ shadcn/ui 初始化完成，components.json 存在
- ✅ 所有页面使用 shadcn 组件，视觉风格统一
- ✅ 响应式布局正常（移动端/桌面端）

---

### M1 整体验收检查清单

```
[ ] go build ./... 零错误
[ ] cmd/server 单体模式可正常启动
[ ] /healthz 返回 200
[ ] 数据库 ent migrate 执行成功
[ ] 用户注册 API 正常
[ ] 用户登录返回 JWT Token
[ ] Token 认证中间件拦截未授权请求
[ ] 前端三层路由结构就位（auth/app/admin）
[ ] 前端登录页可正常使用
[ ] 前端 Admin 页面需要登录才能访问
[ ] Wire 依赖注入在单体模式下正常工作
```

---

## M2：媒体上传与播放

> **目标**：用户能上传视频/图片文件，系统存储后可列表展示并基础播放（无转码，直接原始文件）。  
> **预估时间**：M1 完成后再 2 周（累计 4 周）  
> **完成标准**：用户登录后上传一个 MP4 文件，系统保存到本地磁盘，前端媒体列表出现该条目，点击可播放。

---

### T2.1 文件上传 API

**任务清单**：
- [ ] 在 media proto 中添加 `UploadMedia` RPC（multipart/form-data）
- [ ] 在 `svc-media/service/` 实现上传 handler
  - 接收文件流（支持大文件分块，最大限制可配置）
  - 验证文件类型（允许：mp4/mov/avi/mkv/jpg/jpeg/png/gif）
  - 生成唯一文件名（UUID + 原始扩展名）
  - 写入本地存储目录（`/data/media/uploads/`，可配置）
- [ ] 创建 Media 数据库记录（状态为 `pending`）
- [ ] 发布 `media.uploaded` Pub/Sub 事件
- [ ] 在 svc-api-gateway 注册上传路由（`POST /api/v1/media/upload`）

**测试方案**：
```bash
# 上传视频文件（需先登录获取 Token）
curl -X POST http://localhost:9090/api/v1/media/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/path/to/test.mp4" \
  -F "title=测试视频" \
  -F "description=这是一个测试"

# 期望响应
# {"id":"uuid-xxx","title":"测试视频","status":"pending","file_url":"/media/uploads/uuid-xxx.mp4"}

# 验证文件存在
ls /data/media/uploads/

# 验证数据库记录
psql -c "SELECT id, title, media_type, status FROM media ORDER BY created_at DESC LIMIT 1;"
```

**验收标准**：
- ✅ 上传 MP4（< 100MB）返回 201，含 media ID
- ✅ 上传 JPG 图片返回 201
- ✅ 上传不支持格式（exe/pdf）返回 400
- ✅ 文件物理存储到磁盘指定目录
- ✅ 数据库 media 表创建对应记录

---

### T2.2 媒体列表与详情 API

**任务清单**：
- [ ] 确认 `ListMedia` RPC 正常工作（分页、按分类过滤、按关键词搜索）
- [ ] 确认 `GetMedia` RPC 返回完整媒体信息（含文件 URL）
- [ ] 添加静态文件服务（`GET /media/uploads/:filename`）供前端直接访问
- [ ] 添加 `UpdateMedia` RPC（修改标题/描述/分类/可见性）
- [ ] 添加 `DeleteMedia` RPC（软删除，同时删除物理文件）
- [ ] `IncrementViewCount` 接入前端播放事件触发

**测试方案**：
```bash
# 获取媒体列表
curl http://localhost:9090/api/v1/media \
  -H "Authorization: Bearer $TOKEN"

# 获取媒体详情
curl http://localhost:9090/api/v1/media/$MEDIA_ID \
  -H "Authorization: Bearer $TOKEN"

# 更新媒体
curl -X PUT http://localhost:9090/api/v1/media/$MEDIA_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"新标题","description":"新描述"}'

# 删除媒体
curl -X DELETE http://localhost:9090/api/v1/media/$MEDIA_ID \
  -H "Authorization: Bearer $TOKEN"

# 访问媒体文件
curl -I http://localhost:9090/media/uploads/uuid-xxx.mp4
# 期望：200 Content-Type: video/mp4
```

**验收标准**：
- ✅ `GET /api/v1/media` 返回分页列表
- ✅ `GET /api/v1/media/:id` 返回媒体详情含 `file_url`
- ✅ `file_url` 指向的文件可直接 HTTP 访问（流式播放）
- ✅ 删除后物理文件从磁盘移除
- ✅ 播放量计数正确累加

---

### T2.3 前端上传与播放功能

**任务清单**：
- [ ] 实现上传组件 `web/src/components/MediaUpload.tsx`
  - 拖拽上传 + 点击选择
  - 上传进度条（使用 axios onUploadProgress）
  - 上传成功/失败 Toast 提示
- [ ] 完善 Admin 媒体管理页 `pages/admin/Media.tsx`
  - 添加"上传"按钮，打开上传 Dialog
  - 列表展示：封面图、标题、状态、上传时间、操作按钮
  - 编辑表单（标题/描述/分类/可见性）
  - 删除确认 AlertDialog
- [ ] 实现媒体播放页 `pages/media/[id].tsx`
  - HTML5 `<video>` 原生播放器（M2 阶段不用 HLS.js）
  - 视频信息展示（标题、描述、上传者、播放量）
  - 观看时触发 `IncrementViewCount`
- [ ] 引入 TanStack Query 管理 API 请求状态（代替 useState + useEffect）

**测试方案**（手动 E2E）：
1. 登录后进入 Admin → 媒体管理
2. 点击上传按钮，选择 MP4 文件
3. 观察进度条正常显示
4. 上传完成后列表刷新出现新条目
5. 点击播放，视频正常播放
6. 检查播放量 +1

**验收标准**：
- ✅ 上传组件支持拖拽和点击选择
- ✅ 上传进度条实时更新
- ✅ 上传成功显示 Toast，列表自动刷新
- ✅ 媒体列表显示正确
- ✅ 点击视频可在播放页播放

---

### T2.4 缩略图生成（基础版）

**任务清单**：
- [ ] 视频上传后，使用 ffmpeg 截取第 5 秒帧作为缩略图（异步执行）
- [ ] 图片上传后，使用 `imaging` 库生成 300x200 缩略图
- [ ] 缩略图存储到 `/data/media/thumbnails/` 目录
- [ ] 更新 media 记录的 `thumbnail_url` 字段
- [ ] 前端媒体列表使用 `thumbnail_url` 展示封面图

**测试方案**：
```bash
# 上传视频后，等待约 5-10s
curl http://localhost:9090/api/v1/media/$MEDIA_ID | jq '.thumbnail_url'
# 期望：非空字符串

# 访问缩略图
curl -I http://localhost:9090/media/thumbnails/uuid-xxx.jpg
# 期望：200 Content-Type: image/jpeg
```

**验收标准**：
- ✅ 视频上传后自动生成缩略图
- ✅ 图片上传后自动生成缩略图
- ✅ 媒体列表封面图正确显示

---

### M2 整体验收检查清单

```
[ ] 用户可通过前端上传 MP4 视频（< 500MB）
[ ] 用户可通过前端上传 JPG/PNG 图片
[ ] 上传文件持久化到磁盘
[ ] 媒体列表正确展示所有上传内容
[ ] 点击视频可直接在浏览器播放（HTML5 player）
[ ] 播放量统计正常工作
[ ] 缩略图自动生成
[ ] 删除媒体同时删除物理文件
```

---

## M3：视频转码与 HLS

> **目标**：上传的视频经异步转码，生成多分辨率 HLS 流，前端使用 HLS.js 实现自适应码率播放。  
> **预估时间**：M2 完成后再 4 周（累计 8 周）  
> **完成标准**：上传 MP4 → 后台转码 → 生成 m3u8 → 前端自适应播放，encoding_status 状态机完整。

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
- [ ] 参考 `backend/features/identity` 实现细粒度权限
- [ ] 定义角色：`superadmin` / `admin` / `editor` / `user`
- [ ] 权限规则：
  - `superadmin`：所有操作
  - `admin`：用户管理、媒体管理、内容审核
  - `editor`：上传媒体、管理自己的内容
  - `user`：浏览、评论、收藏、上传（可配置关闭）
- [ ] 在 JWT Token 中携带 roles
- [ ] 后端中间件验证权限
- [ ] 前端根据角色显示/隐藏操作按钮

**验收标准**：
- ✅ 普通用户无法访问 Admin 管理接口
- ✅ Editor 只能管理自己上传的媒体
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

> **目标**：系统具备生产部署条件，支持对象存储、全文搜索、监控可观测性、限流、链路追踪。  
> **预估时间**：M4 完成后再 8 周（累计 20 周）

---

### T5.1 对象存储（MinIO/S3）

**任务清单**：
- [ ] 参考 `backend/features/objectstore` 实现对象存储抽象层
- [ ] 支持本地存储和 MinIO 两种后端（通过配置切换）
- [ ] 媒体文件、缩略图、HLS 切片全部存到对象存储
- [ ] 生成带签名的预签名 URL（有效期可配置）

**验收标准**：
- ✅ 配置 `storage.backend=minio` 后文件自动存到 MinIO
- ✅ 本地存储与 MinIO 切换不影响业务逻辑
- ✅ 媒体 URL 返回预签名地址，过期后无法访问

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

*最后更新：2026-03-31 | 请在每完成一个任务后更新对应 `[ ]` 为 `[x]` 并注明完成日期*
