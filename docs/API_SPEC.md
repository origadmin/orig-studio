# API 规范说明

> 版本: 1.0 | 日期: 2026-04-09 | 状态: 正式

---

## 一、API 契约唯一来源

**重要声明**：API 契约以 `api/proto` 目录下的 proto 文件为唯一可信来源。

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

## 二、文档生成

- **OpenAPI 文档**：通过 grpc-gateway 或 buf + openapiv2 从 proto 文件自动生成
- **API 路径索引**：参考 `BACKEND_ARCHITECTURE_DESIGN.md` 中的完整路径设计
- **接口变更**：所有 API 变更必须先修改 proto 文件，再通过代码生成更新

## 三、API 版本

- 当前版本：`/api/v1`
- 版本管理：通过 proto 文件中的 `google.api.http` 注解声明

## 四、响应格式

所有 API 响应统一使用以下格式：

```json
{
  "code": 0,        // 错误码，0 表示成功
  "message": "ok",  // 消息
  "data": {...}     // 数据
}
```

## 五、错误码体系

参考 `internal/server/error.go` 中的错误码定义。

## 六、路由规范

1. **静态路由优先**：静态路径必须在参数路径之前注册
2. **`/me` 独立**：`/me` 是独立顶级端点，不属于 `/users/` 资源树
3. **统计参数**：统计类接口使用 query 参数，不建独立路径
4. **列表响应**：列表接口响应中携带统计字段

## 七、参数命名规范

- 搜索关键词：`keyword`
- 分页大小：`page_size`
- 页码：`page`（从 1 开始）

## 八、参考文档

- **架构设计**：`BACKEND_ARCHITECTURE_DESIGN.md`
- **API 对齐**：`API_ALIGNMENT.md`
- **历史版本**：`reference/API_DESIGN_V3.md`（已废弃）

---

**注意**：本文档内容由 proto 文件生成，手动修改无效。
