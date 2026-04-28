# Orig-CMS API 测试文档

## 测试架构

使用内存SQLite数据库 + httptest进行集成测试，完全隔离不影响实际数据。

### 测试角色
- **Guest** (未登录) - 只读公开资源
- **User** (普通用户) - 管理自己的资源
- **Staff** - 管理所有资源
- **Admin** - 完全权限

## 测试文件说明

| 文件 | 覆盖模块 | 优先级 |
|------|----------|--------|
| `auth_test.go` | 登录/注册/登出/当前用户 | P0 |
| `user_test.go` | 用户CRUD/订阅 | P1 |
| `media_test.go` | 媒体CRUD/点赞/收藏/分享 | P1 |
| `content_test.go` | 分类/标签/评论/搜索/信息流 | P1 |
| `interaction_test.go` | 点赞/收藏/订阅/播放列表/频道/通知 | P1 |
| `permission_test.go` | 权限边界测试 | P2 |

## 运行测试

### 运行所有集成测试
```bash
go test ./tests/integration/... -v
```

### 运行特定测试
```bash
# 只运行认证测试
go test ./tests/integration/... -v -run TestAuth

# 只运行权限测试
go test ./tests/integration/... -v -run TestPermission
```

### 运行端到端工作流测试
```bash
go test ./tests/e2e/... -v
```

## 测试覆盖率

### API端点覆盖清单

#### 认证 (auth)
- [x] POST /auth/signin
- [x] POST /auth/signup
- [x] POST /auth/signout
- [x] GET /auth/me

#### 用户 (users)
- [x] GET /users
- [x] GET /users/:id
- [x] POST /users
- [x] DELETE /users/:id
- [x] GET /users/:id/subscription
- [x] POST /users/:id/subscribe
- [x] DELETE /users/:id/subscribe

#### 媒体 (media)
- [x] GET /media
- [x] GET /media/:id
- [x] PUT /media/:id
- [x] DELETE /media/:id
- [x] POST /media/upload
- [x] GET /media/:id/variants
- [x] POST /media/:id/like
- [x] POST /media/:id/dislike
- [x] GET /media/:id/like
- [x] POST /media/:id/favorite
- [x] GET /media/:id/favorite
- [x] GET /media/:id/share
- [x] POST /media/:id/share

#### 分类 (categories)
- [x] GET /categories
- [x] POST /categories
- [x] GET /categories/:id
- [x] DELETE /categories/:id

#### 标签 (tags)
- [x] GET /tags
- [x] POST /tags
- [x] GET /tags/:id
- [x] GET /tags/:id/media
- [x] DELETE /tags/:id

#### 评论 (comments)
- [x] GET /comments
- [x] POST /comments
- [x] GET /comments/media/:id
- [x] PUT /comments/:id
- [x] DELETE /comments/:id

#### 播放列表 (playlists)
- [x] GET /playlists
- [x] GET /playlists/:id
- [x] POST /playlists
- [x] GET /playlists/my

#### 频道 (channels)
- [x] GET /channels
- [x] GET /channels/:id
- [x] POST /channels
- [x] GET /channels/user/:id

#### 订阅 (subscriptions)
- [x] GET /subscriptions
- [x] GET /followers

#### 点赞 (likes)
- [x] POST /likes
- [x] GET /likes/media/:id
- [x] GET /likes/check/:id

#### 收藏 (favorites)
- [x] GET /favorites
- [x] POST /favorites
- [x] GET /favorites/check/:id

#### 通知 (notifications)
- [x] GET /notifications
- [x] PUT /notifications/:id/read

#### 统计 (stats)
- [x] GET /stats/dashboard

#### 搜索 (search)
- [x] GET /search

#### 信息流 (feed)
- [x] GET /feed

## 权限测试矩阵

| 角色 | 公开端点 | 认证端点 | 管理员端点 |
|------|---------|----------|------------|
| Guest | ✓ | ✗ (401) | ✗ (401) |
| User | ✓ | ✓ | ✗ (403) |
| Staff | ✓ | ✓ | ✓ |
| Admin | ✓ | ✓ | ✓ |

## 测试辅助函数

位于 `helper_test.go`：
- `SetupTestServer(t *testing.T) *TestServer` - 创建测试服务器
- `MakeRequest(opts RequestOptions)` - 发送HTTP请求
- `AssertStatus(t *testing.T, resp *http.Response, expected int)` - 断言状态码
- `AssertJSON(t *testing.T, body []byte, checks map[string]interface{})` - 断言JSON字段
