# 媒体转码中心与配置管理方案 (M3 阶段扩展)

## 核心目标
1.  **可见性**：通过 SSE 实现前端进度条。
2.  **可管性**：在管理后台提供转码预设 (Encode Profile) 的全生命周期管理（CRUD）。
3.  **稳定性**：后端自动初始化默认配置，确保新环境直接可用。

## 方案设计

### 1. 后端：转码预设管理 (Encode Profile CRUD) - [DONE]
- **[COMPLETED] `api/proto/v1/media/media_service.proto`**:
    - 定义了 `ListEncodeProfiles`, `GetEncodeProfile`, `CreateEncodeProfile`, `UpdateEncodeProfile`, `DeleteEncodeProfile` RPC 接口及其 HTTP 映射。
- **[COMPLETED] `internal/svc-media/service/media.go`**:
    - 实现了上述 RPC 接口的处理函数。
- **[COMPLETED] `internal/svc-media/biz/media.go`**:
    - 在 `MediaUseCase` 中增加了对 `EncodeProfile` 的 CRUD 逻辑。
- **[COMPLETED] `internal/svc-media/data/seed.go`**:
    - 实现了 `SeedEncodeProfiles`：当数据库为空时，自动插入 1080p (5Mbps), 720p (2.5Mbps), 360p (1Mbps) 等配置。

### 2. 前端：转码配置中心 (Admin UI) - [DONE]
- **[COMPLETED] `web/src/lib/api/media.ts`**:
    - 增加了对转码预设、任务列表、SSE 地址及转码状态 API 的调用封装。
- **[COMPLETED] `web/src/pages/admin/TranscodingProfiles.tsx`**:
    - 实现了转码预设的管理界面（CRUD），支持查看、添加、编辑和删除配置。

### 3. 前端：实时监控 (Dashboard) - [DONE]
- **[COMPLETED] `web/src/hooks/useTranscoding.ts`**:
    - 封装了 SSE 连接逻辑，提供 `lastEvent` 和 `status` 状态供组件使用。
- **[COMPLETED] `web/src/pages/admin/TranscodingStatus.tsx`**:
    - 实现了转码监控仪表盘，实时展示正在处理、排队中、成功及失败的任务统计。

## 验证计划
1.  **初始化验证**：确认数据库自动生成了默认配置。
2.  **配置修改验证**：修改码率，观察转码后文件大小。
3.  **UI 验证**：通过转码监控页面观察实时进度。

