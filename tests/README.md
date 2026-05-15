# 测试目录结构规范

> **适用**: orig-studio 项目
> **关联**: `projects/team-flow/team/workflows/roles/development-standards.md`、`projects/team-flow/team/workflows/roles/bugfix-standards.md`

---

## 目录结构

```
tests/
├── README.md                        ← 本文件
│
├── unit/                            ← 单元测试（公共）
│   ├── enums_test.go
│   ├── media_test.go
│   └── user_test.go
│
├── integration/                     ← 集成测试（公共）
│   ├── api_test.go
│   ├── auth_test.go
│   ├── content_test.go
│   ├── media_test.go
│   ├── permission_test.go
│   └── user_test.go
│
├── api/                             ← API 测试（公共）
│   └── test_content_api.go
│
├── e2e/                             ← E2E 测试（公共）
│   ├── e2e_test.go
│   └── workflow_test.go
│
├── features/                        ← 📌 按功能组织的测试
│   ├── F014-unified-pagination/
│   │   ├── TEST_COVERAGE.md          ← 必须：规定测什么
│   │   ├── TEST_CASES.md             ← 必须：具体测试用例
│   │   └── integration_*.go          ← 功能级集成测试
│   │
│   └── F015-xxx/
│       ├── TEST_COVERAGE.md
│       └── ...
│
└── bugs/                            ← 📌 按 Bug 组织的测试
    ├── B001-xxx/
    │   ├── TEST_CASE.md              ← 必须：复现测试用例
    │   └── regression_*.go           ← 回归测试
    │
    └── B002-xxx/
        └── ...
```

---

## 命名规则

### 功能测试（features/）

| 文件 | 说明 | 必须 |
|------|------|------|
| `TEST_COVERAGE.md` | 测试覆盖声明 | ✅ |
| `TEST_CASES.md` | 具体测试用例 | ✅ |
| `*_test.go` | 测试代码 | ✅ |

**目录命名**: `F{xxx}-{short-name}`（如 `F014-unified-pagination`）

### Bug 测试（bugs/）

| 文件 | 说明 | 必须 |
|------|------|------|
| `TEST_CASE.md` | 复现和回归测试用例 | ✅ |
| `regression_*.go` | 回归测试代码 | ✅ |

**目录命名**: `B{xxx}-{short-name}`（如 `B001-media-status-bug`）

---

## 模板文件

测试文档模板位于 `projects/team-flow/team/templates/`：

| 模板 | 用途 |
|------|------|
| `projects/team-flow/team/templates/feature-test-template.md` | Feature 测试覆盖声明 |
| `projects/team-flow/team/templates/bug-test-template.md` | Bug 复现测试用例 |

---

## 触发条件

| 变更类型 | 创建目录 | 使用模板 |
|---------|---------|---------|
| Feature 开发 | `tests/features/Fxxx/` | `feature-test-template.md` |
| Bug 修复 | `tests/bugs/Bxxx/` | `bug-test-template.md` |

---

## 禁止项

- ❌ 在根目录创建散乱的测试文件（统一按功能/Bug 组织）
- ❌ 测试文件与文档分离（代码和文档必须成对存在）
- ❌ 跳过 TEST_COVERAGE.md / TEST_CASE.md 直接写测试代码