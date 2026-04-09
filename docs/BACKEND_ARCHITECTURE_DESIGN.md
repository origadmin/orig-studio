# Orig-CMS 后端架构重构方案

> 版本: 1.0 | 日期: 2026-04-09 | 状态: 正式方案

---

## 一、现状问题诊断

### 1.1 问题全景

通过分析 6 份 API 文档和现有代码，问题可以归为三个层次：

```
┌─────────────────────────────────────────────────────────┐
│  Layer 1: API 契约层 (最根本)                             │
│  - Proto 定义 vs 文档 vs 实现 三套标准并存，互相矛盾      │
│  - 没有"唯一可信来源"                                     │
├─────────────────────────────────────────────────────────┤
│  Layer 2: 路由设计层 (影响稳定性)                         │
│  - 路由冲突：静态路径 vs 参数路径未严格分离               │
│  - 命名混乱：media/medias、encoding/transcoding 混用      │
│  - 模块归属不清：点赞/收藏/订阅放在哪里没有定论            │
├─────────────────────────────────────────────────────────┤
│  Layer 3: 实现完成度层 (影响功能)                         │
│  - 已实现 ~30 个 API，缺失 ~70 个                         │
│  - 9 个完整功能模块完全没有后端实现                        │
│  - 前后端路径不匹配导致功能直接失效                        │
└─────────────────────────────────────────────────────────┘
```

### 1.2 关键问题清单

#### 🔴 P0 - 路由冲突（会导致运行时 500/404）

| 问题 | 位置 | 根因 |
|------|------|------|
| `/media/retry` 被匹配为 `/media/:id`(id="retry") | media.go | 静态路由在参数路由之后注册 |
| `/notifications/unread-count` 被匹配为 `/:id` | notification.go | 同上 |
| `/channels/:id/subscribers/count` 被匹配为 `/:subId` | channel.go | 嵌套参数路由冲突 |
| `/users/me/stats` 被匹配为 `/:id/stats`(id="me") | user.go | `/users/me` 路径语义错误，`me` 不是 ID |

#### 🔴 P0 - 前后端路径不匹配（功能直接失效）

| 功能 | 前端期望 | 后端实现 |
|------|---------|---------|
| 点赞 | `/media/:id/like` | `/likes/*` |
| 收藏 | `/media/:id/favorite` | `/favorites/*` |

#### 🟠 P1 - API 契约无唯一来源

- Proto 定义用 `/medias`（复数），文档用 `/media`（单数）
- Proto 用 `/transcoding`，文档用 `/encoding`
- Proto 无 `/users/me`，文档添加了但 Proto 未同步
- 开发者不知道以哪个为准

#### 🟠 P1 - 完整功能模块缺失

| 缺失模块 | 影响功能 | 优先级 |
|---------|---------|--------|
| Search | 搜索页面完全不可用 | 高 |
| Stats | Admin Dashboard 不可用 | 高 |
| Subscription | 订阅/关注功能不可用 | 高 |
| Share | InteractionBar 分享失效 | 高 |
| History | 观看历史不可用 | 中 |
| Subtitle | 字幕功能不可用 | 中 |
| Review | 内容审核不可用 | 中 |
| Metadata | 媒体分析不可用 | 低 |
| Content | 门户内容页不可用 | 中 |

#### 🟡 P2 - 架构设计问题

- `svc-content` 职责过重（8 个子功能混在一起）
- `server/` 包含所有 handler，耦合严重
- 服务间直接调用，无防腐层接口
- 无统一错误码体系

---

## 二、架构设计决策

### 2.1 核心原则：以 Proto 为唯一契约来源

**当前混乱根因**：Proto、V2文档、V3文档三者互相矛盾，开发者无所适从。

**解决方案**：

```
┌──────────────────────────────────────────────────┐
│              Single Source of Truth               │
│                                                   │
│  Proto 文件 (.proto)  ←  唯一权威定义             │
│       ↓                                           │
│  HTTP Binding 注解   ←  路径在 proto 中声明        │
│       ↓                                           │
│  Generated Code      ←  代码从 proto 生成         │
│       ↓                                           │
│  OpenAPI Spec        ←  文档从 proto 生成         │
└──────────────────────────────────────────────────┘
```

**立即行动**：选定 V3 文档（已解决路由冲突）作为 Proto 更新基准，更新 proto 后：
1. 所有 HTTP 路径在 proto 中通过 `google.api.http` 注解声明
2. 文档从 proto 自动生成（grpc-gateway / buf + openapiv2）
3. 不再允许文档和 proto "各自演进"

### 2.2 API 命名最终裁决

经过对 V2/V3 文档分析，采用以下决策：

| 争议点 | 裁定结果 | 理由 |
|--------|---------|------|
| `/media` vs `/medias` | **`/medias`** | Proto 已用复数，RESTful 标准复数 |
| `encoding` vs `transcoding` | **`encoding`** | 更通用，适合未来音频等场景 |
| `/users/login` vs `/auth/login` | **`/auth/login`** | 认证独立模块，更清晰 |
| 点赞归属 `/likes` vs `/medias/:id/likes` | **`/medias/:id/likes`** | 资源归属清晰，符合 RESTful 嵌套规范 |
| 订阅归属 `/interactions` vs `/channels/:id` | **`/channels/:id/subscription`** | 订阅是频道的行为，资源归属明确 |
| `/playlists/me` vs `/users/me/playlists` | **`/users/:id/playlists`** | 统一到用户资源树，避免 `/me` 子路由冲突 |
| 统计 `/stats` vs `/admin/stats` | **`/admin/stats`** | 管理功能统一在 `/admin` 前缀下 |

### 2.3 路由冲突彻底解决方案

**规则一：`/me` 是独立顶级端点，不属于 `/users/` 资源树**

`me` 不是 ID，不是 `:id` 的特殊值。`/users/:id` 的语义是"通过 ID 访问某个用户资源"，`me` 是身份标识符，两者本质不同。`/users/me` 从语义上就是错误的。

```
✅ 正确：/me 作为顶级端点，独立路由树
  GET    /me                    # 我的信息
  PUT    /me                    # 更新我的信息
  PUT    /me/password           # 修改密码
  GET    /me/playlists          # 我的播放列表
  GET    /me/favorites          # 我的收藏
  GET    /me/likes              # 我的点赞
  GET    /me/subscriptions      # 我的订阅
  GET    /me/history            # 我的历史
  GET    /me/stats              # 我的统计

✅ 正确：/:id 路由树只处理具体用户（数字/UUID ID）
  GET    /users/:id             # 查看某用户
  GET    /users/:id/playlists   # 某用户的播放列表（公开）
  GET    /users/:id/followers   # 某用户的粉丝
  PATCH  /users/:id/status      # 管理员操作

❌ 错误：把 me 放在 /users/ 下，无论如何设计都是错的
  ❌ /users/me                  ← 语义错误，me 不是 ID
  ❌ /users/me/playlists        ← 同上
  ❌ /users/:id/playlists?owner=me  ← 参数语义不清
```

**规则二：统计类接口用 query 参数，不建独立路径**

```
✅ GET /channels/:id/subscribers?count=true     ← 返回数量
✅ GET /channels/:id/subscribers                ← 返回列表

❌ GET /channels/:id/subscribers/count          ← 禁止，/count 与 /:subscriberId 冲突
```

**规则三：列表接口响应中携带统计字段**

```json
// GET /notifications 响应
{
  "items": [...],
  "total": 100,
  "unread_count": 5,    ← 直接附带，不需要独立接口
  "page": 1,
  "page_size": 20
}
```

**规则四：路由注册顺序规范**

```go
func (h *Handler) Register(group *gin.RouterGroup) {
    g := group.Group("/medias")
    
    // Step 1: 静态子路由（按字母序）
    g.POST("/upload", ...)
    
    // Step 2: 带静态前缀的子路由组（encoding, admin 等）
    encoding := g.Group("/encoding")
    encoding.GET("/status", ...)
    encoding.GET("/tasks", ...)
    encoding.GET("/profiles", ...)
    // ...
    
    // Step 3: 集合操作
    g.GET("", h.list)
    g.POST("", h.create)
    
    // Step 4: 嵌套资源路由（带 :id 的子路由）
    g.GET("/:id/encoding/tasks", ...)
    g.GET("/:id/encoding/variants", ...)
    g.GET("/:id/likes", ...)
    g.POST("/:id/likes", ...)
    
    // Step 5: 主资源参数路由（必须最后）
    g.GET("/:id", h.get)
    g.PUT("/:id", h.update)
    g.DELETE("/:id", h.delete)
}
```

---

## 三、最终 API 路径设计（权威版）

### 3.1 完整路径索引

#### 🔑 认证模块 `/auth`

```
POST   /api/v1/auth/login              # 登录（原 /auth/signin）
POST   /api/v1/auth/logout             # 登出
POST   /api/v1/auth/refresh            # 刷新 Token
POST   /api/v1/auth/register           # 注册
POST   /api/v1/auth/forgot-password    # 忘记密码
POST   /api/v1/auth/reset-password     # 重置密码
GET    /api/v1/auth/me                 # 获取当前用户信息（已实现，保留）
```

**路径统一决定**：
- `/auth/signin` → `/auth/login`（signin 是前端词汇，login 是 HTTP 标准）

#### 👤 当前用户模块 `/me`

```
GET    /api/v1/me                    # 我的信息
PUT    /api/v1/me                    # 更新我的信息
PUT    /api/v1/me/password           # 修改密码
GET    /api/v1/me/playlists          # 我的播放列表
GET    /api/v1/me/favorites          # 我的收藏
GET    /api/v1/me/likes              # 我的点赞
GET    /api/v1/me/subscriptions      # 我的订阅频道
GET    /api/v1/me/history            # 观看历史
GET    /api/v1/me/stats              # 我的统计
```

#### 👥 用户资源模块 `/users`

```
GET    /api/v1/users                 # 用户列表（管理员）
POST   /api/v1/users                 # 创建用户（管理员）
GET    /api/v1/users/:id             # 用户详情
PUT    /api/v1/users/:id             # 更新用户（管理员）
DELETE /api/v1/users/:id             # 删除用户（管理员）
PATCH  /api/v1/users/:id/status      # 启用/禁用用户（管理员）
GET    /api/v1/users/:id/stats       # 用户统计
GET    /api/v1/users/:id/playlists   # 某用户的公开播放列表
GET    /api/v1/users/:id/followers   # 某用户的粉丝
```

**设计说明**：
- `/me` 是**身份端点**，表示"当前登录用户"，与 `/users/:id` 完全独立
- `/me` 下的子路由是**当前用户**的资源，需要登录
- `/users/:id` 下的子路由是**指定用户**的公开资源，部分需要登录

#### 📺 频道模块 `/channels`

```
GET    /api/v1/channels                          # 频道列表
POST   /api/v1/channels                          # 创建频道
GET    /api/v1/channels/:id                      # 频道详情
PUT    /api/v1/channels/:id                      # 更新频道
DELETE /api/v1/channels/:id                      # 删除频道
GET    /api/v1/channels/:id/medias               # 频道媒体列表
POST   /api/v1/channels/:id/medias               # 添加媒体到频道
DELETE /api/v1/channels/:id/medias/:mediaId      # 从频道移除媒体
GET    /api/v1/channels/:id/subscribers          # 订阅者（?count=true 返回数量）
GET    /api/v1/channels/:id/subscription         # 当前用户订阅状态
POST   /api/v1/channels/:id/subscription         # 订阅
DELETE /api/v1/channels/:id/subscription         # 取消订阅
```

#### 🎬 媒体模块 `/medias`

```
GET    /api/v1/medias                            # 媒体列表
POST   /api/v1/medias/upload                     # 上传文件（静态路由，优先）
GET    /api/v1/medias                            # 媒体列表
POST   /api/v1/medias                            # 创建媒体记录
GET    /api/v1/medias/:id                        # 媒体详情
PUT    /api/v1/medias/:id                        # 更新媒体
DELETE /api/v1/medias/:id                        # 删除媒体
POST   /api/v1/medias/:id/view                   # 增加播放量
GET    /api/v1/medias/:id/stream                 # 流媒体播放
GET    /api/v1/medias/:id/download               # 下载
GET    /api/v1/medias/:id/thumbnail              # 缩略图
GET    /api/v1/medias/:id/tasks                  # 该媒体的转码任务
GET    /api/v1/medias/:id/variants               # 可用变体
POST   /api/v1/medias/:id/tasks/:taskId/retry    # 重试单个转码任务
GET    /api/v1/medias/:id/likes                  # 点赞列表（附带数量）
POST   /api/v1/medias/:id/likes                  # 点赞/取消点赞（toggle）
DELETE /api/v1/medias/:id/likes                  # 取消点赞（明确删除）
GET    /api/v1/medias/:id/favorites              # 收藏列表
POST   /api/v1/medias/:id/favorites              # 收藏/取消收藏（toggle）
DELETE /api/v1/medias/:id/favorites              # 取消收藏（明确删除）
GET    /api/v1/medias/:id/shares                 # 分享统计
POST   /api/v1/medias/:id/shares                 # 创建分享记录
GET    /api/v1/medias/:id/comments               # 媒体评论列表
GET    /api/v1/medias/:id/subtitles              # 字幕列表
POST   /api/v1/medias/:id/subtitles              # 上传字幕
GET    /api/v1/medias/:id/metadata               # 媒体元数据
```

#### 📋 播放列表模块 `/playlists`

```
GET    /api/v1/playlists                         # 列表
POST   /api/v1/playlists                         # 创建
GET    /api/v1/playlists/:id                     # 详情
PUT    /api/v1/playlists/:id                     # 更新
DELETE /api/v1/playlists/:id                     # 删除
POST   /api/v1/playlists/:id/medias              # 添加媒体
PUT    /api/v1/playlists/:id/medias/reorder      # 重排序（静态路由，优先）
DELETE /api/v1/playlists/:id/medias/:mediaId     # 移除媒体
```

#### 💬 评论模块 `/comments`

```
GET    /api/v1/comments                          # 列表（支持 ?media_id= 过滤）
POST   /api/v1/comments                          # 创建
GET    /api/v1/comments/:id                      # 详情
PUT    /api/v1/comments/:id                      # 更新
DELETE /api/v1/comments/:id                      # 删除
GET    /api/v1/comments/:id/replies              # 回复列表
POST   /api/v1/comments/:id/replies              # 创建回复
```

#### 🔔 通知模块 `/notifications`

```
GET    /api/v1/notifications                     # 列表（响应含 unread_count）
POST   /api/v1/notifications/read-all            # 全部已读（静态路由，优先）
POST   /api/v1/notifications/:id/read            # 标记已读
DELETE /api/v1/notifications/:id                 # 删除通知
```

> **注意**：删除 `/notifications/unread-count` 独立接口，改为在列表响应体中附带 `unread_count` 字段。

#### 🔍 搜索模块 `/search`

```
GET    /api/v1/search                            # 综合搜索（?q=keyword&type=all）
GET    /api/v1/search/suggestions                # 搜索建议/自动补全
```

#### 🏷️ 分类 & 标签模块

```
GET    /api/v1/categories                        # 分类列表（支持树形）
POST   /api/v1/categories                        # 创建分类（管理员）
GET    /api/v1/categories/:id                    # 分类详情
PUT    /api/v1/categories/:id                    # 更新（管理员）
DELETE /api/v1/categories/:id                    # 删除（管理员）

GET    /api/v1/tags                              # 标签列表
POST   /api/v1/tags                              # 创建标签（管理员）
GET    /api/v1/tags/:id                          # 标签详情
PUT    /api/v1/tags/:id                          # 更新（管理员）
DELETE /api/v1/tags/:id                          # 删除（管理员）
```

#### 🛡️ 管理模块 `/admin`（管理员专属）

```
# 统计
GET    /api/v1/admin/stats/dashboard
GET    /api/v1/admin/stats/medias
GET    /api/v1/admin/stats/users
GET    /api/v1/admin/stats/traffic

# 转码管理（全局视角，区别于 /medias/:id/tasks）
GET    /api/v1/admin/encoding/tasks
GET    /api/v1/admin/encoding/status
POST   /api/v1/admin/encoding/retry-failed       # 重试所有失败
POST   /api/v1/admin/encoding/tasks/:taskId/retry

# 转码配置
GET    /api/v1/admin/encoding/profiles
POST   /api/v1/admin/encoding/profiles
GET    /api/v1/admin/encoding/profiles/:id
PUT    /api/v1/admin/encoding/profiles/:id
DELETE /api/v1/admin/encoding/profiles/:id

# 内容审核
GET    /api/v1/admin/review/pending
GET    /api/v1/admin/review/history
GET    /api/v1/admin/review/:id
PUT    /api/v1/admin/review/:id                  # 审核操作
PUT    /api/v1/admin/review/batch                # 批量审核

# 系统配置
GET    /api/v1/admin/settings
PUT    /api/v1/admin/settings
```

#### 🌐 门户模块 `/portal`（公开内容）

```
GET    /api/v1/portal/home                       # 首页推荐
GET    /api/v1/portal/trending                   # 热门内容
GET    /api/v1/portal/subscriptions              # 关注动态（需登录）
```

---

## 四、后端代码架构重构方案

### 4.1 分层架构（保持不变，强化落实）

```
HTTP Request
     ↓
┌──────────────────────────────────────────────────────┐
│  server/  (HTTP Handler 层)                           │
│  - 参数解析、请求验证、响应格式化                      │
│  - 只依赖 biz 层的接口，不直接调用 ent                 │
└──────────────────────────────────────────────────────┘
     ↓ (接口调用)
┌──────────────────────────────────────────────────────┐
│  biz/  (业务逻辑层)                                   │
│  - 核心业务规则                                        │
│  - 编排多个 dal 调用                                   │
│  - 不依赖 Gin，只依赖 dal 接口                         │
└──────────────────────────────────────────────────────┘
     ↓ (接口调用)
┌──────────────────────────────────────────────────────┐
│  dal/  (数据访问层)                                   │
│  - ent 查询封装                                        │
│  - 只暴露接口，不暴露 ent 实体                          │
└──────────────────────────────────────────────────────┘
     ↓
┌──────────────────────────────────────────────────────┐
│  data/entity/  (ent 生成代码)                         │
└──────────────────────────────────────────────────────┘
```

### 4.2 Server 层重构：按功能分文件，统一结构

**当前问题**：`interaction.go` 和 `like.go`、`favorite.go`、`share.go` 并存，职责重叠。

**重构后文件结构**：

```
internal/server/
├── router.go           # 路由聚合（RegisterRoutes）
├── middleware.go        # 中间件（JWT、CORS、Rate Limit 等）
├── response.go          # 统一响应格式
├── error.go             # 错误码映射
├── auth.go              # /auth 模块
├── me.go                # /me 模块（新增）
├── user.go              # /users 模块
├── channel.go           # /channels 模块
├── media.go             # /medias 模块（含交互子路由）
├── playlist.go          # /playlists 模块
├── comment.go           # /comments 模块
├── notification.go      # /notifications 模块
├── search.go            # /search 模块
├── category.go          # /categories 模块
├── tag.go               # /tags 模块
├── admin.go             # /admin 模块
└── portal.go            # /portal 模块

# 删除以下文件（功能并入 media.go）：
# - like.go       → media.go 中的 /:id/likes 处理
# - favorite.go   → media.go 中的 /:id/favorites 处理
# - share.go      → media.go 中的 /:id/shares 处理
# - interaction.go → 已重复，删除
# - stats.go       → 并入 admin.go
# - system.go      → 并入 admin.go
# - feed.go        → 并入 portal.go
# - upload.go      → 并入 media.go
```

### 4.3 统一响应格式

```go
// internal/server/response.go

type Response[T any] struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    T      `json:"data,omitempty"`
}

type PageResponse[T any] struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    struct {
        Items      []T   `json:"items"`
        Total      int64 `json:"total"`
        Page       int   `json:"page"`
        PageSize   int   `json:"page_size"`
    } `json:"data"`
}

// 用于通知等需要附带统计字段的响应
type NotificationListResponse struct {
    Items       []NotificationItem `json:"items"`
    Total       int64              `json:"total"`
    UnreadCount int64              `json:"unread_count"`
    Page        int                `json:"page"`
    PageSize    int                `json:"page_size"`
}

func OK(c *gin.Context, data any) {
    c.JSON(200, Response[any]{Code: 0, Message: "ok", Data: data})
}

func Fail(c *gin.Context, code int, message string) {
    c.JSON(getHTTPStatus(code), Response[any]{Code: code, Message: message})
}
```

### 4.4 统一错误码体系

```go
// internal/server/error.go

const (
    // 通用
    ErrOK          = 0
    ErrInternal    = 10000
    ErrNotFound    = 10001
    ErrUnauthorized = 10002
    ErrForbidden   = 10003
    ErrBadRequest  = 10004
    ErrConflict    = 10005

    // 业务
    ErrUserNotFound    = 20001
    ErrUserExists      = 20002
    ErrPasswordWrong   = 20003
    ErrTokenExpired    = 20004
    ErrTokenInvalid    = 20005

    ErrMediaNotFound   = 30001
    ErrMediaTooLarge   = 30002
    ErrMediaForbidden  = 30003
    ErrEncodingFailed  = 30004

    ErrCommentNotFound = 40001
    ErrCommentForbidden = 40002
)

func getHTTPStatus(code int) int {
    switch {
    case code == ErrOK:       return 200
    case code == ErrNotFound: return 404
    case code == ErrUnauthorized: return 401
    case code == ErrForbidden:    return 403
    case code == ErrBadRequest:   return 400
    case code == ErrConflict:     return 409
    default:                  return 500
    }
}
```

### 4.5 `/me` 路由组实现

```go
// internal/server/router.go

func RegisterRoutes(r *gin.Engine, handlers *Handlers) {
    v1 := r.Group("/api/v1")
    
    // 认证模块
    auth := v1.Group("/auth")
    auth.POST("/login", handlers.Auth.Login)
    auth.POST("/register", handlers.Auth.Register)
    // ...
    
    // /me 独立路由组 - 当前用户相关
    me := v1.Group("/me")
    me.Use(RequireAuth())
    me.GET("", handlers.Me.Get)
    me.PUT("", handlers.Me.Update)
    me.PUT("/password", handlers.Me.UpdatePassword)
    me.GET("/playlists", handlers.Me.GetPlaylists)
    me.GET("/favorites", handlers.Me.GetFavorites)
    me.GET("/likes", handlers.Me.GetLikes)
    me.GET("/subscriptions", handlers.Me.GetSubscriptions)
    me.GET("/history", handlers.Me.GetHistory)
    me.GET("/stats", handlers.Me.GetStats)
    
    // /users 路由组 - 用户资源（管理员操作）
    users := v1.Group("/users")
    users.Use(RequireAuth(), RequireAdmin())
    users.GET("", handlers.User.List)
    users.POST("", handlers.User.Create)
    users.GET("/:id", handlers.User.Get)
    users.PUT("/:id", handlers.User.Update)
    users.DELETE("/:id", handlers.User.Delete)
    users.PATCH("/:id/status", handlers.User.UpdateStatus)
    users.GET("/:id/stats", handlers.User.GetStats)
    users.GET("/:id/playlists", handlers.User.GetPlaylists)
    users.GET("/:id/followers", handlers.User.GetFollowers)
    // ...
}
```

---

## 五、实施计划

### Phase 0：建立契约基准（1-2天）

> 目标：消除"多套标准并存"的根本问题

- [ ] 更新 `api/proto/v1/user/user_service.proto` 按 V3 文档修正路径
- [ ] 更新 `api/proto/v1/media/media_service.proto` 按 V3 文档修正路径
- [ ] 将 V3 文档（API_DESIGN_V3.md）标记为 **DEPRECATED**，注明"以 Proto 为准"
- [ ] 创建 `docs/API_SPEC.md`，内容说明"API 契约以 proto 文件为准，本文档由 proto 生成"

### Phase 1：修复 P0 问题（2-3天）

> 目标：消除路由冲突，修复前后端路径不匹配

**Task 1.1：修复路由注册顺序（media.go）**

```go
// 修复前（有问题）
media.GET("/transcoding/status", ...)
media.GET("/:id/variants", ...)
media.POST("/retry", ...)              // ← 静态路由在参数路由之后！

// 修复后
media.GET("/encoding/status", ...)     // ← 静态路由全部提前
media.POST("/encoding/retry", ...)
media.POST("/encoding/retry-failed", ...)
media.GET("/:id/variants", ...)        // ← 参数路由最后
```

**Task 1.2：修复点赞/收藏路径不匹配**

```go
// internal/server/media.go
// 在 /:id 子路由组中添加
mediaItem := media.Group("/:id")
mediaItem.POST("/likes", h.toggleLike)
mediaItem.DELETE("/likes", h.unlike)
mediaItem.POST("/favorites", h.toggleFavorite)
mediaItem.DELETE("/favorites", h.unfavorite)

// 同时删除 like.go 和 favorite.go 中的独立路由（或标记废弃）
```

**Task 1.3：修复 notification 路由冲突**

```go
// 将 /notifications/read-all 注册在 /:id 之前
notifications.POST("/read-all", h.readAll)   // ← 静态路由
notifications.GET("", h.list)
notifications.POST("/:id/read", h.markRead)  // ← 参数路由最后
notifications.DELETE("/:id", h.delete)
```

**Task 1.4：修复 channel 订阅路由冲突**

```go
// 删除 /subscribers/count，改用 query 参数
channel.GET("/:id/subscribers", func(c *gin.Context) {
    if c.Query("count") == "true" {
        // 返回数量
        return
    }
    // 返回列表
})
```

### Phase 2：实现高优先级缺失功能（1-2周）

> 目标：打通核心用户流程

| Task | 文件 | 依赖 |
|------|------|------|
| 实现 Search API | server/search.go + biz/search.go | 可先做 DB 全文搜索，后升级 ES |
| 实现 Stats API | server/admin.go | 聚合查询即可 |
| 实现 Subscription API | server/channel.go | ent channel schema 需添加 M2M |
| 实现 Share API | server/media.go | 简单计数即可 |
| 实现 `PUT /users/:id` | server/user.go | 已有 Create，复用逻辑 |
| 实现 `PATCH /users/:id/status` | server/user.go | 简单字段更新 |
| 实现 Token Refresh | server/auth.go | JWT 已有基础 |

### Phase 3：结构性重构（2-3周）

> 目标：消除重复代码，提升可维护性

- [ ] 合并 `like.go`/`favorite.go`/`share.go`/`interaction.go` → 迁移到 `media.go` 中
- [ ] 合并 `stats.go`/`system.go` → 迁移到 `admin.go` 中
- [ ] 合并 `feed.go` → 迁移到 `portal.go` 中
- [ ] 合并 `upload.go` → 迁移到 `media.go` 中
- [ ] 添加统一响应格式（`response.go`）
- [ ] 添加统一错误码（`error.go`）
- [ ] 添加 `me.go` 模块，实现 `/me` 独立路由组

### Phase 4：实现中优先级功能（1个月内）

- [ ] History 观看历史
- [ ] Subtitle 字幕
- [ ] Review 审核
- [ ] Content 门户内容（首页推荐、热门等）
- [ ] Portal 模块

---

## 六、迁移兼容策略

### 6.1 API 版本过渡

由于 orig-cms 是早期项目，API 未对外公开，可以采用**硬迁移**策略：

```
1. 旧路径：直接修改为新路径
2. 前端同步更新（同一个仓库，可原子操作）
3. 不需要维护兼容层

注意：修改时前后端必须同步提交，避免出现一方先上线导致断开。
```

### 6.2 参数名统一

| 当前混用 | 统一为 |
|---------|--------|
| `q` / `keyword` / `query` | **`keyword`** |
| `limit` / `page_size` | **`page_size`** |
| `offset` / `page` | **`page`**（从 1 开始） |

---

## 七、监控与可观察性建议

### 7.1 必须加入的中间件

```go
// internal/server/middleware.go

// 1. Request ID（所有请求打上唯一 ID，便于追踪）
func RequestID() gin.HandlerFunc { ... }

// 2. 访问日志（记录 method/path/status/latency/request_id）
func AccessLog(logger *slog.Logger) gin.HandlerFunc { ... }

// 3. 统一错误恢复（panic → 500，不泄露堆栈）
func Recovery() gin.HandlerFunc { ... }

// 4. 速率限制（可选，使用 redis 或内存 token bucket）
func RateLimit(config RateLimitConfig) gin.HandlerFunc { ... }
```

### 7.2 健康检查扩展

```go
// 当前只有 /health 返回 {"status": "ok"}
// 建议扩展为：
GET /health         → 基础存活检查
GET /health/ready   → 就绪检查（包含 DB 连通性检查）
GET /health/live    → 存活检查（轻量级）
```

---

## 八、总结

### 本方案的核心价值

1. **消除 API 混乱的根本原因**：确立 Proto 为唯一可信来源，文档从 proto 生成
2. **路由冲突零容忍**：三条明确规则彻底解决 Gin 路由冲突
3. **优先级驱动**：P0 问题（路由冲突/路径不匹配）优先修复，让功能先跑起来
4. **渐进式重构**：不要求一次重写，Phase 1 可以在 2-3 天内完成并显著改善现状
5. **前后端契约明确**：API_PATH_ALIGNMENT.md 作为前端迁移指南，后端按 Phase 逐步实现

### 当前可立即执行的最小行动

如果只有一天时间，做这三件事效果最大：

```
1. 修复 media.go 路由注册顺序（30分钟）
   → 解决 /media/retry 被识别为 /:id 的 bug

2. 在 media.go 中添加 /:id/likes 和 /:id/favorites 路由（1小时）
   → 修复前端点赞/收藏功能直接 404

3. 在 notifications 中将静态路由 /read-all 移到 /:id 之前（15分钟）
   → 修复全部已读功能
```

> 以上三项修改代码量不超过 50 行，但能修复最影响用户体验的 3 个核心 bug。
