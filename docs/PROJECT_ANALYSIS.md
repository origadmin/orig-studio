# orig-cms 项目分析文档

> **生成时间**：2026-03-31  
> **项目路径**：`D:\workspace\project\golang\origadmin\framework\projects\orig-cms`

---

## 一、项目概述

### 目标定位

orig-cms 是一个基于 Go 微服务架构的现代媒体内容管理系统，采用以下技术栈：

| 层次 | 技术选型 |
|------|----------|
| 后端语言 | Go 1.21+ |
| 微服务框架 | go-kratos v2 |
| 自有运行时 | origadmin/runtime |
| ORM | ent v0.14.6 |
| 数据库 | PostgreSQL |
| 消息总线 | Watermill (Pub/Sub) |
| 服务发现/配置 | Consul |
| 前端 | React + Bun + shadcn/ui |
| API 协议 | gRPC (protobuf) + HTTP REST |
| 依赖注入 | Wire |

### 微服务划分

```
cmd/
├── server/          # 单体模式入口（M1 主力）
├── svc-user/        # 用户微服务 ✅ 主力已实现
├── svc-media/       # 媒体微服务 🔶 存根
├── svc-portal/      # 门户聚合服务 🔶 存根
├── svc-content/     # 内容服务    🔴 仅 README 占位
└── svc-api-gateway/ # API 网关    🔶 存根
```

---

## 二、当前完成度总览

### 2.1 后端完成度评估

| 模块 | 完成度 | 状态说明 |
|------|--------|----------|
| **数据层 (ent schema)** | 85% | 13 个实体 schema 完整 |
| **用户服务 (svc-user)** | 75% | CRUD 完整、角色管理、密码管理、事件发布均已实现 |
| **媒体服务 (svc-media)** | 60% | biz/service/data/repo 层完整，但 cmd 入口是存根 |
| **门户服务 (svc-portal)** | 40% | biz 层实现了聚合逻辑，但 cmd 是存根，service 层缺失 |
| **内容服务 (svc-content)** | 5% | 仅 README 文件，无任何实现 |
| **API 网关 (svc-api-gateway)** | 5% | 存根，无实际逻辑 |
| **Proto/API 定义** | 70% | user/media/portal 三个 proto 文件定义完整 |
| **配置系统** | 70% | 基于 origadmin/runtime 配置加载，conf pb 已接入 |
| **Pub/Sub 事件系统** | 50% | 仅定义了 user 相关 3 个 topic，media 事件缺失 |
| **认证/鉴权** | 30% | 有 origadmin/contrib/security 接入，identity service 未实现 |
| **文件上传/转码** | 0% | 核心功能，完全缺失 |
| **服务间通信 (gRPC)** | 40% | svc-portal 中使用了 gRPC client，但 client 工厂未实现 |
| **Wire 依赖注入** | 60% | svc-user 完整，其余服务不完整 |
| **数据库迁移** | 60% | ent schema 完整但 migrate 脚本不完善 |

### 2.2 前端完成度评估

| 模块 | 完成度 | 状态说明 |
|------|--------|----------|
| **Admin - Dashboard** | 50% | 静态结构，无实时数据 |
| **Admin - 媒体管理** | 50% | 列表/删除基础实现，无上传、无编辑表单 |
| **Admin - 用户管理** | 50% | 列表/删除基础实现，无新增表单、无角色管理 |
| **Admin - 内容管理** | 30% | 页面存在但内容简单 |
| **用户端门户** | 20% | pages/user 目录存在但实现不完整 |
| **API 层封装** | 60% | lib/api.ts 有基础封装，但不对应 proto 生成接口 |
| **路由系统** | 60% | React Router 基础路由，无权限路由守卫 |
| **主题/亮暗模式** | 30% | shadcn/ui 已集成，但主题切换未实现 |
| **国际化 (i18n)** | 0% | 完全缺失 |
| **状态管理** | 20% | 使用 useState 本地状态，无全局状态管理 |
| **错误处理/Toast** | 10% | 使用 alert() 弹窗，未使用 shadcn Toast |
| **表单验证** | 10% | 基本确认弹窗，无表单组件 |

---

## 三、功能实现对比

### 3.1 已实现 ✅

| 功能 | 状态 | 位置 |
|------|------|------|
| 用户模型 | ✅ 完成 | `entity/schema/user.go` + `svc-user` |
| 媒体模型 | ✅ Schema 完整 | `entity/schema/media.go` |
| 分类模型 | ✅ Schema 完整 | `entity/schema/category.go` |
| 频道模型 | ✅ Schema 完整 | `entity/schema/channel.go` |
| 播放列表模型 | ✅ Schema 完整 | `entity/schema/playlist.go` |
| 标签模型 | ✅ Schema 完整 | `entity/schema/tag.go` |
| 评论模型 | ✅ Schema 完整 | `entity/schema/comment.go` |
| 通知模型 | ✅ Schema 完整 | `entity/schema/notification.go` |
| 收藏模型 | ✅ Schema 完整 | `entity/schema/favorite.go` |
| 点赞模型 | ✅ Schema 完整 | `entity/schema/like.go` |
| 用户 CRUD API | ✅ 完整实现 | `svc-user/service/user.go` |
| 用户角色管理 | ✅ 完整实现 | `svc-user/biz/user.go` |
| 密码哈希/验证 | ✅ 完整实现 | origadmin/toolkits/crypto/hash |
| 媒体 CRUD API | ✅ 基本实现 | `svc-media/service/media.go` |
| 媒体分类 API | ✅ 基本实现 | `svc-media/data/media_repo.go` |
| 播放量统计 | ✅ 基本实现 | IncrementViewCount |
| 事件驱动架构 | ✅ 基本实现 | Watermill + pubsub topics |
| 门户首页聚合 | ✅ 基本实现 | `svc-portal/biz/portal.go` |
| 搜索（基础） | ✅ 基本实现 | ListMedias + keyword 过滤 |
| 用户主页 | ✅ 基本实现 | GetUserProfile |

### 3.2 待实现 🔴

| 功能 | 优先级 | 说明 |
|------|--------|------|
| **文件上传 (upload)** | 🔴 P0 | 多种上传方式 |
| **视频转码 (encoding)** | 🔴 P0 | HLS、多分辨率转码，ffmpeg 集成 |
| **视频流 (HLS 播放)** | 🔴 P0 | m3u8 分片，带宽自适应 |
| **缩略图生成** | 🔴 P0 | 自动截帧，支持上传自定义 |
| **API Token 认证** | 🔴 P0 | REST API 认证机制 |
| **OAuth/SSO** | 🔴 P0 | 第三方登录集成 |
| **评论系统 API** | 🔴 P1 | Schema 有，service/biz 层缺失 |
| **通知系统 API** | 🔴 P1 | Schema 有，service/biz 层缺失 |
| **收藏/点赞 API** | 🔴 P1 | Schema 有，service/biz 层缺失 |
| **频道管理 API** | 🔴 P1 | Schema 有，service/biz 层缺失 |
| **播放列表管理 API** | 🔴 P1 | Schema 有，service/biz 层缺失 |
| **标签管理 API** | 🔴 P1 | Schema 有，service/biz 层缺失 |
| **媒体举报** | 🔴 P2 | 需要审核流程 |
| **管理后台 (admin UI)** | 🔴 P2 | 后台管理界面 |
| **内容嵌入 (embed)** | 🔴 P2 | 媒体嵌入其他页面 |
| **RSS/Atom Feed** | 🔴 P2 | 媒体订阅功能 |
| **定时任务** | 🔴 P2 | 定时清理、转码调度 |
| **对象存储 (S3/MinIO)** | 🔴 P2 | 云存储集成 |
| **全文搜索 (ES)** | 🔴 P2 | 当前仅数据库模糊搜索 |
| **Media Info 提取** | 🔴 P2 | ffprobe 集成 |
| **字幕/Whisper AI** | 🔴 P3 | allow_whisper_transcribe 字段已有 |
| **RBAC 细粒度权限** | 🔴 P1 | 角色权限体系 |
| **访问控制（密码保护媒体）** | 🔴 P2 | password 字段已有 |
| **API 限流** | 🔴 P2 | 网关层缺失 |

---

## 四、问题清单

### 4.1 架构问题

| 编号 | 问题 | 严重度 | 说明 |
|------|------|--------|------|
| A-01 | **svc-media/svc-portal/svc-api-gateway cmd 是存根** | 🔴 严重 | 4 个微服务入口均为 `fmt.Println("Hello from...")` 无实际启动逻辑 |
| A-02 | **svc-content 完全未实现** | 🔴 严重 | 仅有 README，无任何 Go 代码 |
| A-03 | **两套 ent entity** | 🔴 严重 | `internal/data/entity` 和 `internal/svc-media/data/entity` 并存，字段不一致 |
| A-04 | **svc-media 未使用 UseCase 模式** | 🟡 中 | MediaService 直接使用 repo，跳过 biz 层 |
| A-05 | **svc-portal 缺 service 层** | 🔴 严重 | portal biz 层实现了聚合，但没有 gRPC service 注册 |
| A-06 | **gRPC 客户端工厂未实现** | 🔴 严重 | svc-portal/data/data.go 中 MediaClient/UserClient 未实现 |
| A-07 | **Wire 依赖注入不完整** | 🔴 严重 | 只有 svc-user 有完整 wire setup |
| A-08 | **认证服务未实现** | 🔴 严重 | Login/Token 发放流程完全缺失 |
| A-09 | **数据库迁移策略不明确** | 🟡 中 | 新建库 vs 兼容旧库未处理 |

### 4.2 代码质量问题

| 编号 | 问题 | 严重度 | 说明 |
|------|------|--------|------|
| C-01 | **proto 字段命名不统一** | 🟡 中 | `Page_size`、`Order_by` 违反 proto3 camelCase 规范 |
| C-02 | **portal.go 字段访问不一致** | 🟡 中 | 混用 `req.Id` 和 `req.GetId()`，有空指针风险 |
| C-03 | **前端 API 层缺少类型安全** | 🟡 中 | 手写类型，未从 proto/openapi 生成 |
| C-04 | **前端错误处理原始** | 🟡 中 | 使用 `alert()`/`confirm()` 而非 shadcn Toast/Dialog |
| C-05 | **CreateUserResponse 变量命名冲突** | 🟡 中 | 局部变量名 `user` 与包名 `user` 冲突 |
| C-06 | **媒体服务双 entity 包** | 🔴 严重 | svc-media 私有 entity 与全局 data/entity 字段不同步 |

### 4.3 功能缺口问题

| 编号 | 问题 | 严重度 | 说明 |
|------|------|--------|------|
| F-01 | **文件上传流程完全缺失** | 🔴 严重 | 无上传 endpoint、无文件存储 |
| F-02 | **转码流程完全缺失** | 🔴 严重 | ffmpeg 未集成 |
| F-03 | **认证流程未闭环** | 🔴 严重 | 前端无登录页，后端无 JWT 发放 |
| F-04 | **media 事件缺失** | 🟡 中 | 只有 user 事件，媒体相关事件未定义 |
| F-05 | **健康检查/探活接口缺失** | 🟡 中 | 微服务部署必需 |
| F-06 | **链路追踪/可观测性缺失** | 🟡 中 | 无 OpenTelemetry 集成 |

---

## 五、文件结构快照

```
orig-cms/
├── api/
│   └── proto/v1/
│       ├── media/       ✅ media_service.proto 完整
│       ├── portal/      ✅ portal_service.proto 完整
│       ├── types/       ✅ 共享类型定义
│       └── user/        ✅ user_service.proto 完整
├── cmd/
│   ├── server/          🔶 基础存在
│   ├── svc-api-gateway/ 🔶 Hello World 存根
│   ├── svc-content/     🔶 Hello World 存根
│   ├── svc-media/       🔶 Hello World 存根
│   ├── svc-portal/      🔶 Hello World 存根
│   └── svc-user/        ✅ 完整启动逻辑
├── configs/             🔶 配置文件结构
├── docs/                ✅ 项目文档（本目录）
├── internal/
│   ├── conf/            🔶 配置加载
│   ├── data/entity/     ✅ 13个 ent schema
│   ├── helpers/         ✅ conf/i18n/idutil/mixin/providers/repo
│   ├── pubsub/          🔶 仅 3 个 user topic
│   ├── server/          🔶 基础路由（media/user/category等）
│   ├── svc-content/     🔴 仅 README
│   ├── svc-media/       🔶 biz/service/data 完整，但使用私有 entity
│   ├── svc-portal/      🔶 biz 完整，缺 service 层
│   └── svc-user/        ✅ 完整的 biz/service/data/dto
└── web/
    └── src/
        ├── pages/admin/ 🔶 基础 5 个管理页面
        ├── pages/user/  🔴 基本未实现
        ├── components/  🔶 基础 shadcn 组件
        └── lib/api.ts   🔶 基础 API 封装
```

---

*本文档基于 2026-03-31 代码快照生成，请在每次重要进展后更新。*  
*详细里程碑计划请见 [MILESTONES.md](./MILESTONES.md)*
