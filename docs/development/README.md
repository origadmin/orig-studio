# Development Documentation

> 规范化开发文档体系，覆盖 design -> implement -> acceptance 全流程。

---

## 目录结构

```
docs/development/
├── README.md                        ← 当前文件
├── modules/                         ← 按模块独立的设计文档
│   ├── media/
│   │   ├── media/                   ← 媒体基础（上传/列表/详情/播放）
│   │   ├── transcode/               ← 转码与 HLS
│   │   └── comment/                 ← 评论系统
│   ├── content/                     ← 内容交互（点赞/收藏/通知/订阅）
│   ├── user/                        ← 用户与认证
│   └── common/                       ← 跨模块共用设计（DB/配置/错误处理）
└── standards/                       ← 编码规范与流程标准
```

## 模块列表

| 模块 | 负责人 | 状态 |
|------|--------|------|
| [media/media](./modules/media/media/) | — | ⚠️ 部分完成 |
| [media/transcode](./modules/media/transcode/) | — | ✅ 完成 |
| [media/comment](./modules/media/comment/) | — | 🔲 未开始 |
| [content](./modules/content/) | — | ⚠️ 部分完成 |
| [user](./modules/user/) | — | ✅ 完成 |
| [common](./modules/common/) | — | 🔲 未开始 |

## 文档编写规范

- **不使用中文注释**：所有代码注释使用英文
- **Design -> Implement -> Acceptance**：每个功能模块必须包含此三部分
- **模块独立**：各模块文档互不干扰，共用部分统一放到 `common/`
- **测试驱动**：优先写测试，再实现，测试文件与业务文件同目录

## 开发流程

1. **Design**：在对应模块目录下编写 `DESIGN.md`
2. **Implement**：按设计实现代码，同时编写单元测试
3. **Acceptance**：编写 `TESTING.md` 或在设计文档中补充验收标准
4. **Review**：提交 PR，CI 必须通过 `go vet ./... && go test ./...`

## 贡献指南

- 分支命名：`feature/{module}-{task}`
- Commit 格式：`type(module): description`
- 所有文档使用 Markdown，表格优先于列表

---

*Last updated: 2026-04-13*
