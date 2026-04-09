# API 对齐状态追踪

> 日期: 2026-04-09 | 版本: 1.0 | 状态: 进行中

---

## 一、API 路径对齐状态

### 1.1 媒体相关 API

| 功能 | 前端路径 | 后端路径 | 状态 | 备注 |
|------|---------|---------|------|------|
| 媒体列表 | `/media` | `/medias` | ❌ 不匹配 | 前端需修改为复数形式 |
| 媒体详情 | `/media/:id` | `/medias/:id` | ❌ 不匹配 | 前端需修改为复数形式 |
| 媒体上传 | `/media/upload` | `/medias/upload` | ❌ 不匹配 | 前端需修改为复数形式 |
| 媒体更新 | `/media/:id` | `/medias/:id` | ❌ 不匹配 | 前端需修改为复数形式 |
| 媒体删除 | `/media/:id` | `/medias/:id` | ❌ 不匹配 | 前端需修改为复数形式 |
| 点赞 | `/media/:id/like` | `/medias/:id/likes` | ❌ 不匹配 | 前端需修改为复数形式 |
| 收藏 | `/media/:id/favorite` | `/medias/:id/favorites` | ❌ 不匹配 | 前端需修改为复数形式 |
| 分享 | `/media/:id/share` | `/medias/:id/shares` | ❌ 不匹配 | 前端需修改为复数形式 |
| 转码状态 | `/media/encoding/status` | `/medias/encoding/status` | ❌ 不匹配 | 前端需修改为复数形式 |
| 转码任务 | `/media/encoding/tasks` | `/medias/encoding/tasks` | ❌ 不匹配 | 前端需修改为复数形式 |

### 1.2 通知相关 API

| 功能 | 前端路径 | 后端路径 | 状态 | 备注 |
|------|---------|---------|------|------|
| 通知列表 | `/notifications` | `/notifications` | ✅ 匹配 | - |
| 未读通知数量 | `/notifications/unread/count` | `/notifications/unread-count` | ❌ 不匹配 | 前端需修改路径 |
| 标记通知已读 | `/notifications/:id/read` | `/notifications/:id/read` | ⚠️ 方法不匹配 | 前端使用 PUT，后端使用 POST |
| 标记所有已读 | `/notifications/read-all` | `/notifications/read-all` | ⚠️ 方法不匹配 | 前端使用 PUT，后端使用 POST |

### 1.3 用户相关 API

| 功能 | 前端路径 | 后端路径 | 状态 | 备注 |
|------|---------|---------|------|------|
| 登录 | `/auth/signin` | `/auth/signin` | ✅ 匹配 | - |
| 注册 | `/auth/signup` | `/auth/signup` | ✅ 匹配 | - |
| 获取当前用户 | `/auth/me` | `/me` | ❌ 不匹配 | 前端需修改为 `/me` |
| 更新用户信息 | `/users/me` | `/me` | ❌ 不匹配 | 前端需修改为 `/me` |
| 修改密码 | `/users/me/password` | `/me/password` | ❌ 不匹配 | 前端需修改为 `/me/password` |

### 1.4 频道相关 API

| 功能 | 前端路径 | 后端路径 | 状态 | 备注 |
|------|---------|---------|------|------|
| 频道列表 | `/channels` | `/channels` | ✅ 匹配 | - |
| 频道详情 | `/channels/:id` | `/channels/:id` | ✅ 匹配 | - |
| 用户频道 | `/channels/user/:userId` | `/channels/user/:userId` | ✅ 匹配 | - |
| 订阅状态 | `/channels/:id/subscription` | `/channels/:id/subscription` | ✅ 匹配 | - |
| 订阅 | `/channels/:id/subscription` | `/channels/:id/subscription` | ✅ 匹配 | - |
| 取消订阅 | `/channels/:id/subscription` | `/channels/:id/subscription` | ✅ 匹配 | - |

## 二、响应格式对齐状态

### 2.1 成功响应

| 功能 | 前端期望格式 | 后端返回格式 | 状态 | 备注 |
|------|-------------|-------------|------|------|
| 媒体列表 | `{list: [...], total: 100, page: 1, page_size: 20}` | `{code: 0, message: "ok", data: {items: [...], total: 100, page: 1, page_size: 20}}` | ❌ 不匹配 | 前端需适配新格式 |
| 媒体详情 | `{id: 1, title: "...", ...}` | `{code: 0, message: "ok", data: {id: 1, title: "...", ...}}` | ❌ 不匹配 | 前端需适配新格式 |
| 点赞状态 | `{is_liked: true, like_count: 10}` | `{code: 0, message: "ok", data: {is_liked: true, like_count: 10}}` | ❌ 不匹配 | 前端需适配新格式 |
| 收藏状态 | `{is_favorited: true}` | `{code: 0, message: "ok", data: {is_favorited: true}}` | ❌ 不匹配 | 前端需适配新格式 |
| 通知列表 | `[{id: 1, title: "...", ...}]` | `{code: 0, message: "ok", data: [{id: 1, title: "...", ...}]}` | ❌ 不匹配 | 前端需适配新格式 |
| 登录响应 | `{token: "...", user: {...}}` | `{code: 0, message: "ok", data: {token: "...", user: {...}}}` | ❌ 不匹配 | 前端需适配新格式 |

### 2.2 错误响应

| 功能 | 前端期望格式 | 后端返回格式 | 状态 | 备注 |
|------|-------------|-------------|------|------|
| 错误响应 | `{error: "错误信息"}` | `{code: 10004, message: "错误信息"}` | ❌ 不匹配 | 前端需适配新格式 |

## 三、HTTP 方法对齐状态

| 功能 | 前端方法 | 后端方法 | 状态 | 备注 |
|------|---------|---------|------|------|
| 标记通知已读 | PUT | POST | ❌ 不匹配 | 前端需修改为 POST |
| 标记所有已读 | PUT | POST | ❌ 不匹配 | 前端需修改为 POST |

## 四、前端 API 文件统一状态

| 文件 | 状态 | 备注 |
|------|------|------|
| `media.ts` | ✅ 核心文件 | 需修改路径为复数形式 |
| `like.ts` | ⚠️ 重复 | 功能已在 media.ts 中实现，建议废弃 |
| `favorite.ts` | ⚠️ 重复 | 功能已在 media.ts 中实现，建议废弃 |
| `share.ts` | ⚠️ 重复 | 功能已在 media.ts 中实现，建议废弃 |
| `interaction.ts` | ⚠️ 重复 | 功能已在 media.ts 中实现，建议废弃 |
| `notification.ts` | ✅ 核心文件 | 需修改方法为 POST |
| `user.ts` | ✅ 核心文件 | 需修改路径为 `/me` |

## 五、修复进度追踪

### 5.1 Phase 1: 紧急修复（P0）

| 任务 | 状态 | 负责人 | 完成日期 |
|------|------|--------|----------|
| P1.1 统一前端API路径 `/media` → `/medias` | ❌ 未开始 | 前端 | - |
| P1.2 适配新响应格式 | ❌ 未开始 | 前端 | - |
| P1.3 修复HTTP方法不匹配 | ❌ 未开始 | 前端 | - |
| P1.4 统一API客户端 | ❌ 未开始 | 前端 | - |
| P1.5 清理后端旧handler | ❌ 未开始 | 后端 | - |

### 5.2 Phase 2: 核心功能补全（P1）

| 任务 | 状态 | 负责人 | 完成日期 |
|------|------|--------|----------|
| P2.1 实现订阅功能完整逻辑 | ❌ 未开始 | 后端 | - |
| P2.2 Dashboard真实数据统计 | ❌ 未开始 | 后端 | - |
| P2.3 前端播放页评论组件 | ❌ 未开始 | 前端 | - |
| P2.4 前端播放页点赞/收藏按钮 | ❌ 未开始 | 前端 | - |
| P2.5 搜索格式统一 | ❌ 未开始 | 后端 | - |

## 六、验证步骤

### 6.1 功能验证清单

- [ ] 媒体列表加载正确（路径 `/medias`，响应格式正确）
- [ ] 媒体详情加载正确（路径 `/medias/:id`，响应格式正确）
- [ ] 点赞/取消点赞正常（路径 `/medias/:id/likes`，响应格式正确）
- [ ] 收藏/取消收藏正常（路径 `/medias/:id/favorites`，响应格式正确）
- [ ] 分享功能正常（路径 `/medias/:id/shares`，响应格式正确）
- [ ] 通知已读功能正常（方法 POST，路径 `/notifications/:id/read`）
- [ ] 未读通知数量正常（路径 `/notifications/unread-count`）
- [ ] 用户信息获取正常（路径 `/me`）
- [ ] 密码修改正常（路径 `/me/password`）

### 6.2 测试命令

```bash
# 后端编译检查
go build ./...

# 前端编译检查
cd web && npm run build

# 启动测试
cd web && npx playwright test
```

## 七、总结

当前 API 对齐状态存在以下问题：

1. **路径不匹配**：前端使用 `/media` 单数形式，后端使用 `/medias` 复数形式
2. **响应格式不匹配**：前端期望直接返回数据，后端返回 `{code, message, data}` 包装格式
3. **HTTP方法不匹配**：通知相关的 PUT vs POST 方法不一致
4. **API文件重复**：多个前端 API 文件功能重复，需要统一

建议按照 Phase 1 的任务清单逐步修复，确保前后端 API 完全对齐。