# orig-cms 项目状态与开发整合文档

> 日期: 2026-04-09 | 版本: 1.0 | 状态: 紧急修复中

---

## 一、当前状态总览

### 1.1 代码实际状态

| 模块 | 状态 | 说明 |
|------|------|------|
| **后端 - 路由注册** | ✅ 已完成 | `media.go` 路由顺序正确 |
| **后端 - 响应格式** | ✅ 已完成 | `OK()` / `Fail()` 统一响应 |
| **后端 - 错误码** | ✅ 已完成 | `error.go` 定义完整 |
| **后端 - /me 路由** | ✅ 已完成 | `me.go` 实现完整 |
| **后端 - 订阅功能** | ❌ 空壳 | `channel.go` 订阅相关全是 TODO |
| **后端 - 旧 handler** | ⚠️ 残留 | `like.go`/`favorite.go` 未清理 |
| **后端 - 搜索** | ⚠️ 旧格式 | `search.go` 返回旧格式 |
| **后端 - 统计** | ❌ 空壳 | `admin.go` 统计全是 TODO |
| **前端 - API 路径** | ❌ 错误 | 使用 `/media` 单数，后端是 `/medias` 复数 |
| **前端 - 响应格式** | ❌ 错误 | 期望 `response.list`，后端返回 `response.data.items` |
| **前端 - 多套 API** | ❌ 混乱 | `media.ts`、`like.ts`、`interaction.ts` 三套并存 |

### 1.2 文档状态

| 文档 | 状态 | 问题 |
|------|------|------|
| `MIGRATION_GUIDE.md` | ⚠️ 部分过时 | Phase 1-2 大部分已完成，未标注 |
| `DEVELOPMENT_PLAN_FAST.md` | ❌ 错误 | 写着 GORM，实际用 Ent |
| `DEVELOPMENT_PLAN.md` | ❌ 错误 | 写着 GORM，实际用 Ent |

---

## 二、关键问题清单

### 🔴 P0：前后端 API 路径完全不匹配

**问题描述：**

| 功能 | 前端调用路径 | 后端实际路径 | 匹配？ |
|------|-------------|-------------|--------|
| 媒体列表 | `GET /media` | `GET /medias` | ❌ |
| 媒体详情 | `GET /media/:id` | `GET /medias/:id` | ❌ |
| 点赞 | `POST /media/:id/like` | `POST /medias/:id/likes` | ❌ |
| 收藏 | `POST /media/:id/favorite` | `POST /medias/:id/favorites` | ❌ |
| 分享 | `POST /media/:id/share` | `POST /medias/:id/shares` | ❌ |

**前端代码证据（media.ts）：**
```typescript
// 文件: web/src/lib/api/media.ts
list: (params) => api.get<MediaListResponse>(\"/media\", params),
get: (id) => api.get<Media>(`/media/${cleanId}`),
```

**后端代码证据（media.go）：**
```go
// 文件: internal/server/media.go
media := group.Group(\"/medias\")  // 复数形式
{
    media.GET(\"\", h.listMedia())          // GET /medias
    media.GET(\"/:id\", h.getMedia())       // GET /medias/:id
    media.POST(\"/:id/likes\", ...)         // POST /medias/:id/likes
    media.POST(\"/:id/favorites\", ...)    // POST /medias/:id/favorites
}
```

**结论：** 前端调用的路径在后端根本不存在，会 404。

---

### 🔴 P0：响应格式完全不匹配

**问题描述：**

后端返回新格式 `{code: 0, message: "ok", data: {...}}`，但前端期望旧格式 `{list: [...], total: 100}`。

**后端代码证据（media.go）：**
```go
// 文件: internal/server/media.go - listMedia()
OK(c, gin.H{
    \"items\":     items,
    \"total\":     total,
    \"page\":      page,
    \"page_size\": pageSize,
})
// 实际返回: {code: 0, message: \"ok\", data: {items: [...], total: 100, ...}}
```

**前端类型定义（media.ts）：**
```typescript
// 文件: web/src/lib/api/media.ts
export interface MediaListResponse {
    list: Media[];      // ← 期望 list 字段
    total: number;      // ← 期望 total 字段
    page: number;
    page_size: number;
}
```

**axios 实际行为：**
```typescript
// request.ts 使用 axios，response.data 是 HTTP 响应 body
const response = await request<T>({url, method, data});
return response.data;  // 返回的是 {code: 0, message: \"ok\", data: {items: [...]}}
```

**前端使用：**
```typescript
const data = await api.get<MediaListResponse>(\"/medias\");
// data.list === undefined ❌
// data.data.items === [...] ✅
```

**结论：** 前端永远拿不到数据，因为字段名和嵌套层级都对不上。

---

### 🔴 P0：HTTP 方法不匹配

| 操作 | 前端方法 | 后端方法 | 文件 |
|------|---------|---------|------|
| 标记通知已读 | `PUT /notifications/:id/read` | `POST /notifications/:id/read` | notification.go |
| 全部已读 | `PUT /notifications/read-all` | `POST /notifications/read-all` | notification.go |
| 未读数 | `GET /notifications/unread/count` | `GET /notifications/unread-count` | notification.go |

---

### 🟠 P1：旧 handler 是死代码

**文件：** `internal/server/like.go`、`internal/server/favorite.go`

这两个文件定义了 `LikeHandler` 和 `FavoriteHandler`，实现了 `Register()` 方法注册路由到 `/likes` 和 `/favorites`。

但是！`media.go` 里已经重新实现了一套完整的点赞/收藏逻辑，注册在 `/medias/:id/likes` 和 `/medias/:id/favorites`。

**问题：**
1. 两套路由同时存在，路径命名不一致
2. 旧 handler 使用旧响应格式 `{error: ...}`
3. 旧 handler 的函数 `toggleLikeStatus`、`getFavoriteStatus` 等根本没注册路由，是死代码
4. 前端调用的路径（`/media/:id/like`）在两套里都不存在

---

### 🟠 P1：订阅功能是空壳

**文件：** `internal/server/channel.go`

```go
// GetSubscriptionStatus - TODO: Implement IsSubscribed
OK(c, gin.H{\"is_subscribed\": false})  // 永远返回 false

// SubscribeToChannel - TODO: Implement Subscribe
OK(c, gin.H{\"success\": true, \"message\": \"Subscribed to channel\"})  // 假成功

// UnsubscribeFromChannel - TODO: Implement Unsubscribe
OK(c, gin.H{\"success\": true, \"message\": \"Unsubscribed from channel\"})  // 假成功
```

前端两套订阅 API：
- 旧：`/users/:userId/subscribe`（不存在）
- 新：`/channels/:id/subscription`（路径对，但后端是假的）

---

### 🟠 P1：文档写错 ORM

**文件：** `DEVELOPMENT_PLAN_FAST.md`、`DEVELOPMENT_PLAN.md`

```markdown
技术栈:
- 后端: Go + Gin + GORM  ❌ 错误！
```

**实际代码：**
```go
// ent/schema/media.go - 项目使用 Ent ORM
type Media struct {
    ent.Schema
}
```

---

### 🟡 P2：搜索返回旧格式

**文件：** `internal/server/search.go`

```go
c.JSON(http.StatusOK, gin.H{
    \"list\":      medias,
    \"total\":     total,
    \"page\":      page,
    \"page_size\": pageSize,
})
// 返回: {list: [...], total: 100, ...}
// 不是: {code: 0, message: \"ok\", data: {items: [...], total: 100, ...}}
```

而 `media.go` 的列表使用 `OK(c, ...)` 返回新格式。两边格式不一致。

---

### 🟡 P2：统计功能是空壳

**文件：** `internal/server/admin.go`

```go
func (h *AdminHandler) getDashboardStats() gin.HandlerFunc {
    return func(c *gin.Context) {
        // TODO: Implement dashboard stats
        c.JSON(http.StatusOK, gin.H{
            \"total_medias\":  100,  // 硬编码
            \"total_users\":   50,
            // ...
        })
    }
}
```

前端调用 `/stats/dashboard`，后端返回假数据。

---

## 三、API 契约（统一标准）

### 3.1 路径标准（以代码为准）

```
资源名: media → medias（复数）
基础路径: /api/v1
```

| 功能 | Method | 路径 | 状态 |
|------|--------|------|------|
| 媒体列表 | GET | `/medias` | ✅ 已实现 |
| 媒体详情 | GET | `/medias/:id` | ✅ 已实现 |
| 媒体上传 | POST | `/medias/upload` | ✅ 已实现 |
| 媒体更新 | PUT | `/medias/:id` | ✅ 已实现 |
| 媒体删除 | DELETE | `/medias/:id` | ✅ 已实现 |
| 点赞 | POST | `/medias/:id/likes` | ✅ 已实现 |
| 取消点赞 | DELETE | `/medias/:id/likes` | ✅ 已实现 |
| 获取点赞状态 | GET | `/medias/:id/likes` | ✅ 已实现 |
| 点踩 | DELETE | `/medias/:id/likes` | ✅ 已实现 |
| 收藏 | POST | `/medias/:id/favorites` | ✅ 已实现 |
| 获取收藏状态 | GET | `/medias/:id/favorites` | ✅ 已实现 |
| 分享 | POST | `/medias/:id/shares` | ✅ 已实现 |
| 获取分享链接 | GET | `/medias/:id/shares` | ✅ 已实现 |
| 获取转码状态 | GET | `/medias/encoding/status` | ✅ 已实现 |
| 获取转码任务 | GET | `/medias/encoding/tasks` | ✅ 已实现 |
| 媒体变体 | GET | `/medias/:id/variants` | ✅ 已实现 |
| 用户资料 | GET | `/me` | ✅ 已实现 |
| 更新资料 | PUT | `/me` | ✅ 已实现 |
| 密码修改 | PUT | `/me/password` | ⚠️ TODO: 实现 |
| 我的播放列表 | GET | `/me/playlists` | ⚠️ TODO: 实现 |
| 我的收藏 | GET | `/me/favorites` | ⚠️ TODO: 实现 |
| 我的点赞 | GET | `/me/likes` | ⚠️ TODO: 实现 |
| 我的订阅 | GET | `/me/subscriptions` | ✅ 已实现 |
| 观看历史 | GET | `/me/history` | ⚠️ TODO: 实现 |
| 我的统计 | GET | `/me/stats` | ⚠️ TODO: 实现 |
| 频道列表 | GET | `/channels` | ✅ 已实现 |
| 频道详情 | GET | `/channels/:id` | ✅ 已实现 |
| 用户频道 | GET | `/channels/user/:userId` | ✅ 已实现 |
| 订阅状态 | GET | `/channels/:id/subscription` | ❌ TODO: 实现 |
| 订阅 | POST | `/channels/:id/subscription` | ❌ TODO: 实现 |
| 取消订阅 | DELETE | `/channels/:id/subscription` | ❌ TODO: 实现 |
| 频道订阅者 | GET | `/channels/:id/subscribers` | ⚠️ TODO: 实现 |
| 通知列表 | GET | `/notifications` | ✅ 已实现 |
| 标记已读 | POST | `/notifications/:id/read` | ✅ 已实现 |
| 全部已读 | POST | `/notifications/read-all` | ✅ 已实现 |
| 未读数 | GET | `/notifications/unread-count` | ✅ 已实现 |
| 搜索 | GET | `/search` | ⚠️ 需改新格式 |
| 仪表盘统计 | GET | `/admin/stats/dashboard` | ❌ TODO: 实现 |
| 媒体统计 | GET | `/admin/stats/medias` | ❌ TODO: 实现 |
| 用户统计 | GET | `/admin/stats/users` | ❌ TODO: 实现 |
| 流量统计 | GET | `/admin/stats/traffic` | ❌ TODO: 实现 |

### 3.2 响应格式标准

**成功响应：**
```json
{
  \"code\": 0,
  \"message\": \"ok\",
  \"data\": { ... }
}
```

**列表响应：**
```json
{
  \"code\": 0,
  \"message\": \"ok\",
  \"data\": {
    \"items\": [...],
    \"total\": 100,
    \"page\": 1,
    \"page_size\": 20
  }
}
```

**错误响应：**
```json
{
  \"code\": 10004,
  \"message\": \"Invalid ID\"
}
```

### 3.3 错误码

| 错误码 | HTTP 状态码 | 说明 |
|--------|-------------|------|
| 0 | 200 | 成功 |
| 10000 | 500 | 服务器内部错误 |
| 10001 | 404 | 资源不存在 |
| 10002 | 401 | 未授权 |
| 10003 | 403 | 禁止访问 |
| 10004 | 400 | 请求参数错误 |
| 10005 | 409 | 资源冲突 |

---

## 四、任务清单

### Phase 1：修复前端 API 路径（P0，1天）

**任务：**
1. [前端] 修改 `media.ts`：`/media` → `/medias`
2. [前端] 修改 `like.ts`：`/media/:id/like` → `/medias/:id/likes`
3. [前端] 修改 `favorite.ts`：`/media/:id/favorite` → `/medias/:id/favorites`
4. [前端] 修改 `share.ts`：`/media/:id/share` → `/medias/:id/shares`
5. [前端] 修改 `notification.ts`：`PUT` → `POST`
6. [前端] 统一使用一套 API（删除 `interaction.ts` 或废弃 `like.ts`/`favorite.ts`）

### Phase 2：修复响应格式适配（P0，1天）

**方案 A（推荐）：前端适配**
1. [前端] 修改 `MediaListResponse` 类型：`items` 替换 `list`
2. [前端] 添加响应包装类型：
   ```typescript
   interface ApiResponse<T> {
     code: number;
     message: string;
     data: T;
   }
   ```
3. [前端] 修改所有 API 调用处适配新格式

**方案 B：后端兼容**
1. [后端] 添加旧格式兼容路由（可选，不推荐）

### Phase 3：修复 HTTP 方法（P0，0.5天）

1. [前端] `notification.ts`：`PUT /notifications/:id/read` → `POST`
2. [前端] `notification.ts`：`PUT /notifications/read-all` → `POST`
3. [前端] `notification.ts`：`GET /notifications/unread/count` → `/notifications/unread-count`

### Phase 4：清理旧代码（P1，0.5天）

1. [后端] 删除 `like.go`（功能已在 `media.go` 实现）
2. [后端] 删除 `favorite.go`（功能已在 `media.go` 实现）
3. [后端] 删除 `interaction.ts`（或标记废弃，前端不再使用）

### Phase 5：实现缺失功能（P1-P2，3-5天）

| 任务 | 优先级 | 预计时间 |
|------|--------|----------|
| 订阅功能实现 | P0 | 1天 |
| 统计 API 实现 | P1 | 1天 |
| 搜索格式统一 | P1 | 0.5天 |
| /me 子路由实现 | P1 | 1天 |
| 密码修改 | P2 | 0.5天 |

---

## 五、废弃代码清理清单

### 5.1 后端待删除

| 文件 | 说明 |
|------|------|
| `internal/server/like.go` | 旧点赞 handler，功能已在 media.go 实现 |
| `internal/server/favorite.go` | 旧收藏 handler，功能已在 media.go 实现 |

### 5.2 前端待删除/废弃

| 文件 | 说明 |
|------|------|
| `web/src/lib/api/interaction.ts` | 与 media.ts 重复，建议统一用 media.ts |
| `web/src/lib/api/like.ts` | 与 media.ts 重复，建议统一用 media.ts |
| `web/src/lib/api/favorite.ts` | 与 media.ts 重复，建议统一用 media.ts |
| `web/src/lib/api/share.ts` | 与 media.ts 重复，建议统一用 media.ts |
| `web/src/lib/api/subscription.ts` | 路径错误，需重写或使用 interaction.ts |

---

## 六、验证步骤

### 6.1 功能验证清单

- [ ] 媒体列表加载正常
- [ ] 媒体详情加载正常
- [ ] 点赞/取消点赞正常
- [ ] 收藏/取消收藏正常
- [ ] 分享功能正常
- [ ] 订阅功能正常
- [ ] 通知已读正常
- [ ] 搜索结果正常
- [ ] 统计面板有真实数据

### 6.2 测试命令

```bash
# 后端
cd internal && go build ./...

# 前端
cd web && npm run build

# 启动测试
cd web && npx playwright test
```

---

## 七、总结

**核心问题：迁移做了一半，前后端各改各的，没对齐。**

1. **后端**：路由和响应格式已改到新标准
2. **前端**：还在用旧路径和旧格式
3. **文档**：写的是 GORM，实际是 Ent

**修复优先级：**
1. 先修 P0（API 路径 + 响应格式 + HTTP 方法）
2. 再清 P1（旧代码 + 订阅功能）
3. 最后做 P2（统计 + 搜索格式 + /me）

**修复后状态：**
- 前后端 API 100% 对齐
- 响应格式统一
- 代码整洁，无死代码
- 文档准确
