# orig-cms 架构决策文档

> **制定时间**：2026-03-31  
> **状态**：草案（需开发团队确认后锁定）  
> **背景**：在进入 M1 开发前，需先明确以下架构问题，避免后续开发混乱

---

## 一、核心问题与决策

### 问题 1：orig-cms 与 backend 的关系？

**结论：有限参考，按需裁剪，不直接照搬**

backend 是通用后台管理系统，拥有完整的系统管理能力（RBAC、用户管理、日志审计、系统配置）。  
orig-cms 是**媒体内容站点**，面向两类用户：
- **管理员**：内容审核、媒体管理、用户管理（场景比 backend 窄）
- **普通用户/访客**：浏览、播放、上传、互动（backend 完全没有这块）

| backend 模块 | orig-cms 处理策略 | 说明 |
|-------------|-----------------|------|
| `features/identity` | **部分采用** | 认证/JWT 逻辑参考，但角色定义按媒体场景裁剪 |
| `features/system` | **精简采用** | 站点设置（站名/logo/注册开关），不需要完整系统管理 |
| `features/filemanager` | **深度定制** | 媒体文件管理与转码绑定，不能直接复用通用文件管理 |
| `features/objectstore` | **直接复用** | 对象存储抽象层可直接引入 |
| `features/notification` | **参考改写** | 通知触发逻辑不同（评论/点赞/关注），需重写 biz 层 |
| `features/cron` | **直接复用** | 转码任务调度、清理任务 |
| `features/logging` | **直接复用** | 日志审计 |

---

### 问题 2：svc-portal 是否保留？

**结论：保留，但重新定位其职责**

**当前误区**：svc-portal 被实现成了 BFF（Backend For Frontend）聚合层，直接调用 svc-user 和 svc-media 的 gRPC。

**问题**：
- 如果有多个前端（Admin 管理后台、用户 Web App、移动 App），每个都需要不同的聚合逻辑，svc-portal 会变成大杂烩
- 聚合层本质是每个客户端的业务逻辑，放服务端会造成过度设计

**重新定位**：

```
svc-portal 职责 = 用户侧公开内容聚合（SEO 友好的服务端渲染数据源）
```

具体边界：
- ✅ 首页 Feed 聚合（热门/最新/推荐内容）
- ✅ 搜索结果聚合（跨 media/user/tag 搜索）
- ✅ 媒体详情页数据聚合（视频 + 作者 + 相关推荐）
- ✅ 用户公开主页数据（用户信息 + 公开视频列表）
- ❌ 不处理认证逻辑（交给 svc-user）
- ❌ 不处理上传、转码（交给 svc-media）
- ❌ 不处理评论/点赞（交给 svc-content）

**实现方式调整**：svc-portal 不再作为独立微服务部署，**合并到 svc-api-gateway 的 HTTP 聚合层**，减少一跳网络调用。

---

### 问题 3：用户侧需要什么架构？

**结论：需要明确区分三层前端架构**

当前 orig-cms 前端已有基础三层雏形，以现有结构为准，不做无谓的目录迁移：

```
┌─────────────────────────────────────────────────────┐
│                   前端架构                           │
├─────────────────────────────────────────────────────┤
│  web/src/pages/admin/     管理后台（需要登录，角色限制）  │
│  web/src/pages/home/      用户 Web App（可选登录）      │
│                           ✅ 已有：index/Media/Search/Watch │
│  web/src/pages/auth/      认证页（登录/注册/找回密码）   │
└─────────────────────────────────────────────────────┘
```

> **注**：`pages/home/` 就是用户侧入口，名字已经描述清楚，无需改为 `pages/app/`。  
> `pages/user/` 当前是空壳（引用不存在的 HomePage），可以清理掉。

**用户侧 Web App 需要的页面**：
- 首页（Feed 流，热门/最新）
- 媒体播放页（视频详情 + 评论 + 相关推荐）
- 搜索页
- 用户主页（公开展示）
- 频道页 / 播放列表页
- 上传页（登录用户）
- 个人中心（已登录用户：历史/收藏/通知）

**管理后台需要的页面**（精简版，非 backend 完整版）：
- Dashboard（媒体统计、用户统计）
- 媒体管理（审核/下架/删除）
- 用户管理（封号/角色调整）
- 分类/标签管理
- 站点设置（基础配置）

---

### 问题 4：服务划分最终方案

**调整后的服务划分**：

```
┌──────────────────────────────────────────────────────────────┐
│                    orig-cms 服务架构                          │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  前端层                                                       │
│  ┌──────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │ Admin UI │  │  用户 Web    │  │  移动端(未来) │           │
│  │ (React)  │  │  App (React) │  │              │           │
│  └─────┬────┘  └──────┬───────┘  └──────┬───────┘           │
│        └──────────────┼──────────────────┘                   │
│                       ▼                                      │
│  ┌────────────────────────────────────────┐                  │
│  │         svc-api-gateway                │                  │
│  │  - HTTP REST 路由 + JWT 验证中间件      │                  │
│  │  - 限流 / CORS / 请求日志              │                  │
│  │  - 内置聚合层（原 portal 职责）         │                  │
│  └────┬──────────┬──────────┬─────────────┘                  │
│       │          │          │                                 │
│       ▼          ▼          ▼                                 │
│  ┌─────────┐ ┌────────┐ ┌──────────┐                         │
│  │svc-user │ │svc-    │ │svc-      │                         │
│  │         │ │media   │ │content   │                         │
│  │用户/认证 │ │媒体/   │ │评论/通知 │                         │
│  │/角色    │ │转码/流  │ │/收藏/赞  │                         │
│  └────┬────┘ └───┬────┘ └────┬─────┘                         │
│       └──────────┴───────────┘                               │
│                  ▼                                           │
│  ┌─────────────────────────────────────┐                     │
│  │     共享基础层                       │                     │
│  │  PostgreSQL │ Redis │ Consul        │                     │
│  │  Watermill (Pub/Sub) │ MinIO(M5)    │                     │
│  └─────────────────────────────────────┘                     │
└──────────────────────────────────────────────────────────────┘
```

**关键决策**：
1. **svc-portal 合并到 svc-api-gateway**，作为聚合路由层，不再独立部署
2. **svc-content 独立存在**，负责所有用户互动内容（评论/通知/收藏/点赞）
3. **svc-user 承担认证**，登录/Token 发放/角色管理全在此服务
4. **svc-media 专注媒体**，上传/存储/转码/流媒体，不碰用户互动

---

### 问题 5：前后端分工的 API 路径规划

```
# 用户侧（匿名可访问）
GET  /api/v1/feed                    → gateway 聚合 → svc-media
GET  /api/v1/media/:id               → svc-media
GET  /api/v1/media/:id/comments      → svc-content
GET  /api/v1/search?q=xxx            → gateway 聚合 → svc-media + svc-user
GET  /api/v1/users/:id/profile       → svc-user
GET  /api/v1/channels/:id            → svc-media

# 用户侧（需要登录）
POST /api/v1/media/upload            → svc-media
POST /api/v1/media/:id/comments      → svc-content
POST /api/v1/media/:id/like          → svc-content
POST /api/v1/media/:id/favorite      → svc-content
GET  /api/v1/notifications           → svc-content
PUT  /api/v1/users/me                → svc-user

# 认证
POST /api/v1/auth/login              → svc-user
POST /api/v1/auth/register           → svc-user
POST /api/v1/auth/refresh            → svc-user
POST /api/v1/auth/logout             → svc-user

# 管理后台（需要 admin/superadmin 角色）
GET  /api/v1/admin/dashboard         → gateway 聚合
GET  /api/v1/admin/media             → svc-media
PUT  /api/v1/admin/media/:id/review  → svc-media
GET  /api/v1/admin/users             → svc-user
PUT  /api/v1/admin/users/:id/ban     → svc-user
GET  /api/v1/admin/categories        → svc-media
POST /api/v1/admin/categories        → svc-media
GET  /api/v1/admin/settings          → svc-user (system settings)
```

---

## 二、服务职责边界（最终版）

### svc-user

**职责**：用户账号、认证、角色权限、站点配置

| 功能 | 实现位置 | 备注 |
|------|---------|------|
| 注册/登录/登出 | svc-user/biz/auth.go | JWT 发放 |
| Token 刷新 | svc-user/biz/auth.go | Refresh Token |
| 用户 CRUD | svc-user/biz/user.go | ✅ 已实现 |
| 角色管理 | svc-user/biz/role.go | ✅ 已实现 |
| 用户封禁 | svc-user/biz/user.go | 扩展 status 字段 |
| 站点设置 | svc-user/biz/settings.go | 参考 backend/features/system 精简版 |
| 密码重置 | svc-user/biz/auth.go | 邮件/手机验证码（M4） |

**不属于 svc-user 的**：
- 用户发布的媒体列表 → svc-media（通过 user_id 过滤）
- 用户的通知 → svc-content
- 用户的收藏/点赞 → svc-content

---

### svc-media

**职责**：媒体内容的全生命周期管理

| 功能 | 实现位置 | 备注 |
|------|---------|------|
| 文件上传 | svc-media/service/upload.go | 支持大文件 chunked |
| 媒体 CRUD | svc-media/service/media.go | ✅ 已实现 |
| 转码任务触发 | svc-media/biz/transcode.go | 发布 Pub/Sub 事件 |
| 转码 Worker | svc-media/worker/transcode.go | 订阅事件，调用 ffmpeg |
| HLS 流媒体 | svc-media/service/stream.go | m3u8/ts 文件服务 |
| 缩略图生成 | svc-media/worker/thumbnail.go | ffmpeg/imaging |
| 分类管理 | svc-media/service/category.go | 管理员操作 |
| 标签管理 | svc-media/service/tag.go | 管理员操作 |
| 频道管理 | svc-media/service/channel.go | 用户操作 |
| 播放列表管理 | svc-media/service/playlist.go | 用户操作 |
| 播放量统计 | svc-media/biz/stats.go | ✅ 已实现 |

> **注**：频道和播放列表属于媒体组织结构，归入 svc-media 而非 svc-content

---

### svc-content

**职责**：用户互动内容（评论、通知、收藏、点赞）

| 功能 | 实现位置 | 备注 |
|------|---------|------|
| 评论 CRUD | svc-content/service/comment.go | 含嵌套回复 |
| 评论审核 | svc-content/biz/moderation.go | 管理员操作 |
| 通知发送 | svc-content/biz/notification.go | 订阅媒体/用户事件 |
| 通知列表 | svc-content/service/notification.go | 用户查询自己的通知 |
| 收藏/取消 | svc-content/service/favorite.go | ToggleFavorite |
| 点赞/取消 | svc-content/service/like.go | ToggleLike |
| 举报 | svc-content/service/report.go | 用户举报内容（M4） |

---

### svc-api-gateway（含原 portal 功能）

**职责**：统一 HTTP 入口 + 聚合层

| 功能 | 实现位置 | 备注 |
|------|---------|------|
| JWT 认证中间件 | gateway/middleware/auth.go | 验证 Token |
| 限流中间件 | gateway/middleware/ratelimit.go | M5 实现 |
| CORS 中间件 | gateway/middleware/cors.go | |
| 请求日志 | gateway/middleware/logging.go | |
| 路由转发 | gateway/router/ | HTTP → gRPC |
| 首页 Feed 聚合 | gateway/handler/feed.go | 原 portal 职责 |
| 搜索聚合 | gateway/handler/search.go | 原 portal 职责 |
| 媒体详情聚合 | gateway/handler/detail.go | 原 portal 职责 |
| Admin Dashboard 聚合 | gateway/handler/dashboard.go | 统计数据聚合 |

---

## 三、前端架构决策

### 目录结构调整

```
web/src/
├── pages/
│   ├── auth/                # 认证相关
│   │   ├── Login.tsx
│   │   ├── Register.tsx
│   │   └── ForgotPassword.tsx
│   ├── app/                 # 用户侧 Web App（公开访问）
│   │   ├── Home.tsx         # 首页 Feed
│   │   ├── Watch.tsx        # 媒体播放页
│   │   ├── Search.tsx       # 搜索页
│   │   ├── Channel.tsx      # 频道页
│   │   ├── Playlist.tsx     # 播放列表页
│   │   ├── Profile.tsx      # 用户公开主页
│   │   └── me/              # 已登录用户专区
│   │       ├── Upload.tsx   # 上传页
│   │       ├── History.tsx  # 观看历史
│   │       ├── Favorites.tsx# 收藏
│   │       └── Notifications.tsx # 通知中心
│   └── admin/               # 管理后台（需 admin 角色）
│       ├── Dashboard.tsx
│       ├── Media.tsx        # 媒体管理（含审核）
│       ├── Users.tsx        # 用户管理
│       ├── Categories.tsx   # 分类管理
│       ├── Tags.tsx         # 标签管理
│       └── Settings.tsx     # 站点设置
├── components/
│   ├── player/
│   │   ├── HLSPlayer.tsx    # HLS 播放器
│   │   └── VideoCard.tsx    # 视频卡片（列表展示用）
│   ├── upload/
│   │   └── MediaUpload.tsx  # 上传组件
│   ├── comment/
│   │   └── CommentSection.tsx
│   └── ui/                  # shadcn/ui 组件
├── hooks/
│   ├── useAuth.ts           # 认证状态
│   ├── useMedia.ts          # 媒体数据
│   └── useNotifications.ts  # 通知
├── lib/
│   ├── api/
│   │   ├── auth.ts
│   │   ├── media.ts
│   │   ├── content.ts
│   │   └── user.ts
│   └── utils.ts
└── router/
    ├── index.tsx            # 路由定义
    └── guards.tsx           # 路由守卫（权限控制）
```

### 路由设计

```
/                       → app/Home.tsx         (公开)
/watch/:id              → app/Watch.tsx         (公开)
/search                 → app/Search.tsx        (公开)
/channel/:id            → app/Channel.tsx       (公开)
/playlist/:id           → app/Playlist.tsx      (公开)
/user/:username         → app/Profile.tsx       (公开)
/upload                 → app/me/Upload.tsx     (需登录)
/me/history             → app/me/History.tsx    (需登录)
/me/favorites           → app/me/Favorites.tsx  (需登录)
/me/notifications       → app/me/Notifications.tsx (需登录)
/login                  → auth/Login.tsx
/register               → auth/Register.tsx
/admin                  → admin/Dashboard.tsx   (需 admin 角色)
/admin/media            → admin/Media.tsx       (需 admin 角色)
/admin/users            → admin/Users.tsx       (需 admin 角色)
/admin/categories       → admin/Categories.tsx  (需 admin 角色)
/admin/settings         → admin/Settings.tsx    (需 admin 角色)
```

---

## 四、后端功能模块归属（明确边界）

### 从 backend 借鉴的模块（直接复用或轻度改造）

| backend 模块 | orig-cms 使用方式 | 改造程度 |
|-------------|-----------------|---------|
| `features/identity` JWT 逻辑 | 复制到 svc-user/biz/auth.go | 轻度（去掉 OAuth provider 部分，M3 后再加）|
| `features/objectstore` | 复制为 internal/storage/ | 零改造，直接用抽象接口 |
| `features/cron` | 复制为 svc-media/worker/cron.go | 轻度（去掉通用 job，只留转码/清理） |
| `features/logging` 审计日志 | 复制为 internal/helpers/audit/ | 轻度改造 |

### 不从 backend 复用的模块（按媒体场景重写）

| backend 模块 | 原因 | orig-cms 替代方案 |
|-------------|------|-----------------|
| `features/system` 完整系统管理 | backend 管理上百个配置项，orig-cms 只需站点基础设置 | 精简的 SiteSettings struct，5-10 个配置项 |
| `features/notification` | backend 的通知是系统通知，orig-cms 的通知是社交通知（评论/点赞） | svc-content 从头实现 |
| `features/filemanager` | backend 的文件管理是通用目录树，orig-cms 的媒体文件与转码深度绑定 | svc-media 自己管理文件存储 |

---

## 五、数据流设计

### 媒体上传完整流程

```
用户 → Gateway → svc-media
                    │
                    ├─ 存文件到本地/MinIO
                    ├─ 创建 Media 记录（status=pending）
                    └─ 发布 Pub/Sub: media.uploaded
                              │
                    ┌─────────┘
                    ▼
              Transcode Worker（svc-media 内部）
                    │
                    ├─ 更新 encoding_status=processing
                    ├─ ffmpeg 转码 → HLS 切片
                    ├─ 更新 encoding_status=success
                    └─ 发布 Pub/Sub: media.encoded
                              │
                    ┌─────────┘
                    ▼
              Notification Worker（svc-content）
                    │
                    └─ 给上传者发通知："你的视频已处理完成"
```

### 用户评论完整流程

```
用户 → Gateway (JWT 验证) → svc-content
                                │
                                ├─ 创建 Comment 记录
                                └─ 发布 Pub/Sub: content.commented
                                          │
                                ┌─────────┘
                                ▼
                          Notification Worker（svc-content 内部）
                                │
                                └─ 给视频作者发通知："有人评论了你的视频"
```

---

## 六、待确认事项

以下决策需要在实际开发中根据情况确认：

| 序号 | 问题 | 倾向方案 | 需确认原因 |
|------|------|---------|-----------|
| D-01 | svc-portal 完全废弃 vs 保留壳 | 废弃，功能并入 gateway | 确认 gateway 承接聚合逻辑的复杂度可接受 |
| D-02 | 前端是否拆成两个独立 App | 保持单一 React App，按路径区分 | 如有 SSR/SEO 需求则考虑拆分 Next.js |
| D-03 | 用户端是否支持游客浏览 | 是（内容公开访问，互动需登录） | 确认内容是否有私密权限控制 |
| D-04 | Redis 何时引入 | M2（缓存 Feed 聚合结果） | 评估 PostgreSQL 直查性能是否够用 |
| D-05 | WebSocket 实现方式 | 直接在 svc-media/svc-content 中实现 | 避免引入额外的实时消息服务 |
| D-06 | 移动端是否规划 | 是，但 M5 之后 | 确认 REST API 设计需支持移动端 |

---

## 七、架构演进路线

```
M1（当前）：单体模式为主，微服务壳子就位
  ├─ server/ 单体入口跑通所有功能
  ├─ svc-user 独立 → 认证逻辑
  └─ 前端：auth + admin + app 目录结构建立

M2：媒体服务独立
  └─ svc-media 真正拆出来（上传/存储/缩略图）

M3：转码异步化
  └─ svc-media 内 Worker 独立运行（Watermill）

M4：内容互动独立
  └─ svc-content 拆出来（评论/通知/收藏）

M5：生产就绪
  ├─ svc-api-gateway 真正承担网关职责（限流/监控）
  └─ svc-portal 废弃（聚合逻辑已在 gateway 中）
```

> **M1 的务实策略**：M1 阶段优先在 `cmd/server`（单体模式）中跑通全流程，  
> 微服务拆分是架构目标，但不是 M1 必须完成的事。  
> 先让功能跑起来，再按服务边界做拆分，比追求微服务完美性更重要。

---

*本文档应在每次重大架构调整后更新。架构决策一旦锁定，对应的 MILESTONES.md 任务也需同步调整。*  
*文档版本：v1.0 | 2026-03-31*
