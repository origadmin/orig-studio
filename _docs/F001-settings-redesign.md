# F001 - Admin Settings 页面全面修复设计文档

> 版本: 1.0 | 日期: 2026-05-09 | 状态: Draft

## 目录

1. [概述](#1-概述)
2. [当前架构分析](#2-当前架构分析)
3. [问题1: 系统TAB无内容](#3-问题1-系统tab无内容)
4. [问题2: 本地存储设置无效](#4-问题2-本地存储设置无效)
5. [问题3: 对象存储选项错误](#5-问题3-对象存储选项错误)
6. [问题4: 站点信息未展示](#6-问题4-站点信息未展示)
7. [问题5: 站点URL多地址支持](#7-问题5-站点url多地址支持)
8. [问题6: 媒体设置无效](#8-问题6-媒体设置无效)
9. [问题7: 邮件发送缺失](#9-问题7-邮件发送缺失)
10. [问题8: 安全设置无效](#10-问题8-安全设置无效)
11. [数据库Schema变更](#11-数据库schema变更)
12. [DefaultSettings完整清单](#12-defaultsettings完整清单)
13. [实施顺序](#13-实施顺序)

---

## 1. 概述

### 1.1 问题总结

Admin Settings 页面存在 8 个核心问题，根本原因是**前端设置表单与后端设置消费严重脱节**：

- 前端可以保存设置到数据库，但后端大部分模块不从数据库读取设置
- 前端表单字段与后端设置键名不匹配
- 前端选项与后端实际支持的类型不一致
- 部分后端功能完全缺失（邮件发送、速率限制等）

### 1.2 设计原则

1. **设置即配置源**: 数据库设置应作为运行时配置的权威来源，优先级高于 YAML 硬编码
2. **前后端契约一致**: 前端表单字段必须与后端 DefaultSettings 键名一一对应
3. **渐进式增强**: 新增设置不应破坏现有功能，需提供合理的 fallback
4. **安全优先**: 敏感设置（密码、密钥）必须标记 `is_sensitive`，API 返回时脱敏

---

## 2. 当前架构分析

### 2.1 设置数据流（当前）

```
前端 Settings.tsx
    │
    ├── GET /admin/settings ──→ AdminHandler.getSystemSettings()
    │                              └── SettingUseCase.ListAll() → DB
    │
    ├── PUT /admin/settings ──→ AdminHandler.updateSystemSettings()
    │                              └── SettingUseCase.BatchUpsert() → DB
    │
    └── GET /admin/settings/info ─→ AdminHandler.getSystemInfo()
                                     └── runtime.MemStats (硬编码)

后端消费方:
    ├── SpriteUseCase ←── ConfigProvider (✅ 唯一正常消费者)
    ├── ModuleGuard   ←── ConfigProvider (✅ 正常)
    ├── PortalConfig  ←── ConfigProvider (✅ 正常, 仅展示)
    ├── StorageConfig ←── conf.DefaultStorageConfig() (❌ 硬编码)
    ├── UploadConfig  ←── conf.DefaultUploadConfig() (❌ 硬编码)
    ├── AuthHandler   ←── 无设置消费 (❌ 不检查)
    └── EmailSender   ←── 不存在 (❌ 缺失)
```

### 2.2 Setting Entity Schema

```go
// schema/setting.go 当前字段
type Setting struct {
    ID            string    // UUIDv7
    Key           string    // 唯一键, max 200
    Value         string    // 文本值
    Type          enum      // string | int | bool | json
    Category      enum      // general | upload | review | email | module
    Description   string    // 可选描述
    IsSensitive   bool      // 是否脱敏
    FallbackValue string    // 默认值
    IsBuiltin     bool      // 是否内置
    CreateTime    time.Time
    UpdateTime    time.Time
}
```

### 2.3 当前 DefaultSettings (25项)

| Category | Key | Type |
|----------|-----|------|
| general | site_name | string |
| general | site_description | string |
| general | base_url | string |
| general | allow_registration | bool |
| upload | allow_upload | bool |
| upload | max_upload_size_video | int |
| upload | max_upload_size_image | int |
| upload | sprite_frame_interval | int |
| upload | sprite_columns | int |
| upload | sprite_frame_width | int |
| upload | sprite_frame_height | int |
| upload | sprite_max_frames | int |
| upload | thumbnail_quality | int |
| upload | thumbnail_resolution | string |
| upload | thumbnail_position | string |
| review | auto_approve | bool |
| review | require_review | bool |
| email | smtp_host | string |
| email | smtp_port | int |
| email | smtp_user | string |
| email | smtp_password | string (sensitive) |
| module | module_articles | bool |
| module | module_videos | bool |
| module | module_music | bool |
| module | homepage_layout | string |

---

## 3. 问题1: 系统TAB无内容

### 3.1 根因

前端 `fetchSystemInfo()` 使用原生 `fetch` 而非 `api` 封装，**不携带 JWT token**，导致请求返回 401。

```typescript
// 当前代码 (Settings.tsx:209)
const response = await fetch('/api/v1/admin/settings/info');
```

### 3.2 修复方案

**前端改动**:

```typescript
// 改为使用 api 封装
import {api} from '@/lib/request';

const fetchSystemInfo = async () => {
    try {
        const info = await api.get<SystemInfo>('/admin/settings/info');
        setSystemInfo(info);
    } catch (error) {
        console.error('Failed to fetch system info:', error);
        setSystemInfo(null);
    }
};
```

**SystemInfo 接口扩展**:

```typescript
interface SystemInfo {
    version: string;
    goVersion: string;
    database: string;
    os: string;
    uptime: string;
    totalMemory: string;
    usedMemory: string;
    cpuUsage: string;
    memoryUsage: number;
    numCPU: number;          // 新增
    numGoroutine: number;    // 新增
}
```

**UI 增强**: 在系统信息卡片中显示 `numCPU` 和 `numGoroutine`。

### 3.3 涉及文件

| 文件 | 改动 |
|------|------|
| `web/src/pages/admin/Settings.tsx` | 替换 fetch 为 api 封装, 扩展 SystemInfo 接口, 添加新字段显示 |

### 3.4 后端改动

无。后端 `getSystemInfo()` 已正确返回所有字段。

---

## 4. 问题2: 本地存储设置无效

### 4.1 根因

1. 前端显示 3 个独立路径（upload_dir/encode_dir/thumb_dir），但后端使用**单一 BasePath** 派生子目录
2. 后端 `NewStorageConfig()` 和 `NewUploadConfig()` 使用硬编码默认值，不从数据库读取
3. 前端保存的路径设置从未被后端消费

### 4.2 当前 vs 目标

| 当前 (前端) | 当前 (后端) | 目标 |
|-------------|-------------|------|
| upload_dir: `/var/media/uploads` | BasePath: `./data/uploads` | storage_base_path (单一) |
| encode_dir: `/var/media/encoded` | ❌ 不存在 | 由 StoragePaths 自动派生 |
| thumb_dir: `/var/media/thumbnails` | 由 StoragePaths 自动派生 | 由 StoragePaths 自动派生 |

### 4.3 修复方案

#### 4.3.1 后端: 添加 storage_base_path 设置

在 `DefaultSettings()` 中新增:

```go
{
    Key:           "storage_base_path",
    Value:         "./data/uploads",
    Type:          setting.TypeString,
    Category:      setting.CategoryUpload,
    Description:   "Base path for media storage (subdirectories are auto-created)",
    FallbackValue: "./data/uploads",
    IsBuiltin:     true,
},
```

#### 4.3.2 后端: NewStorageConfig 从 DB 读取

修改 `wire.go` 中的 `NewStorageConfig` 和 `NewStoragePaths`，使其接受 `ConfigProvider`:

```go
// wire.go - 修改前
func NewStorageConfig() *config.StorageConfig {
    return config.DefaultStorageConfig()
}

func NewStoragePaths(cfg *config.UploadConfig) *config.StoragePaths {
    return config.NewStoragePaths(cfg.StorageBasePath)
}

// wire.go - 修改后
func NewStorageConfig(settingUC *systembiz.SettingUseCase) *config.StorageConfig {
    cfg := config.DefaultStorageConfig()
    basePath := settingUC.Get(context.Background(), "storage_base_path")
    if basePath != "" {
        cfg.BasePath = basePath
    }
    // 读取 storage_type 和 S3 配置 (见问题3)
    return cfg
}

func NewStoragePaths(cfg *config.StorageConfig) *config.StoragePaths {
    return config.NewStoragePaths(cfg.BasePath)
}
```

> **注意**: `NewStorageConfig` 在 wire 初始化时调用，此时 SettingUseCase 的缓存可能尚未加载。
> 解决方案: 使用 `OnceCell` 模式延迟初始化，或在 `NewStoragePaths` 中直接读取 DB。

**替代方案 (推荐)**: 在 `StoragePaths` 上添加 `ReloadFromSettings` 方法，由启动后的 hook 调用:

```go
// 启动后刷新存储配置
func (app *App) PostStart() {
    basePath := app.settingUC.Get(context.Background(), "storage_base_path")
    if basePath != "" && basePath != app.storagePaths.BasePath() {
        app.storagePaths.Reload(basePath)
    }
}
```

#### 4.3.3 前端: 更新存储设置表单

将 3 个独立路径输入替换为单一 BasePath + 子目录预览:

```tsx
// 存储设置 - 新设计
<Card>
    <CardHeader>
        <CardTitle>本地存储</CardTitle>
        <CardDescription>设置媒体文件的基础存储路径</CardDescription>
    </CardHeader>
    <CardContent className="space-y-4">
        <div className="space-y-2">
            <label>存储基础路径</label>
            <Input
                value={formData.storage_base_path}
                onChange={(e) => handleInputChange('storage_base_path', e.target.value)}
                placeholder="./data/uploads"
            />
            <p className="text-xs text-muted-foreground">
                修改后需要重启服务生效
            </p>
        </div>

        {/* 子目录预览 (只读) */}
        <div className="p-4 rounded-lg bg-muted/50 space-y-1 text-sm">
            <p className="font-medium">自动创建的子目录:</p>
            <p>📁 originals/ — 原始文件</p>
            <p>📁 temp/ — 临时上传</p>
            <p>📁 thumbnails/ — 缩略图</p>
            <p>📁 hls/ — HLS 流媒体</p>
            <p>📁 previews/ — GIF 预览</p>
            <p>📁 sprites/ — 雪碧图</p>
        </div>
    </CardContent>
</Card>
```

### 4.4 涉及文件

| 文件 | 改动 |
|------|------|
| `internal/features/system/biz/setting.go` | DefaultSettings 添加 storage_base_path |
| `cmd/server/wire.go` | NewStorageConfig/NewStoragePaths 接受 ConfigProvider |
| `cmd/server/wire_gen.go` | 重新生成 |
| `web/src/pages/admin/Settings.tsx` | 替换 3 个路径输入为单一 BasePath + 预览 |

---

## 5. 问题3: 对象存储选项错误

### 5.1 根因

1. 前端提供 `minio`/`oss` 选项，后端不支持
2. 后端支持 `hybrid` 但前端未提供
3. 存储类型由 YAML 配置决定，不从数据库读取
4. 未检查后端是否实际配置了 S3

### 5.2 后端实际支持的存储类型

```go
// conf/storage_config.go
const (
    StorageTypeLocal  StorageType = "local"
    StorageTypeS3     StorageType = "s3"
    StorageTypeHybrid StorageType = "hybrid"
)
```

### 5.3 修复方案

#### 5.3.1 后端: 添加存储相关设置到 DefaultSettings

```go
// 新增设置项
{
    Key:           "storage_type",
    Value:         "local",
    Type:          setting.TypeString,
    Category:      setting.CategoryUpload,
    Description:   "Storage backend type: local, s3, hybrid",
    FallbackValue: "local",
    IsBuiltin:     true,
},
{
    Key:           "s3_endpoint",
    Value:         "",
    Type:          setting.TypeString,
    Category:      setting.CategoryUpload,
    Description:   "S3/MinIO endpoint URL",
    FallbackValue: "",
    IsBuiltin:     true,
},
{
    Key:           "s3_region",
    Value:         "",
    Type:          setting.TypeString,
    Category:      setting.CategoryUpload,
    Description:   "S3 region",
    FallbackValue: "",
    IsBuiltin:     true,
},
{
    Key:           "s3_bucket",
    Value:         "",
    Type:          setting.TypeString,
    Category:      setting.CategoryUpload,
    Description:   "S3 bucket name",
    FallbackValue: "",
    IsBuiltin:     true,
},
{
    Key:           "s3_access_key",
    Value:         "",
    Type:          setting.TypeString,
    Category:      setting.CategoryUpload,
    Description:   "S3 access key ID",
    FallbackValue: "",
    IsSensitive:   true,
    IsBuiltin:     true,
},
{
    Key:           "s3_secret_key",
    Value:         "",
    Type:          setting.TypeString,
    Category:      setting.CategoryUpload,
    Description:   "S3 secret access key",
    FallbackValue: "",
    IsSensitive:   true,
    IsBuiltin:     true,
},
{
    Key:           "s3_use_path_style",
    Value:         "false",
    Type:          setting.TypeBool,
    Category:      setting.CategoryUpload,
    Description:   "Use path-style S3 URLs (true for MinIO)",
    FallbackValue: "false",
    IsBuiltin:     true,
},
```

#### 5.3.2 后端: Storage Capabilities API

添加新端点，让前端知道后端是否支持 S3:

```go
// system_handler.go
func (h *SystemHandler) getStorageCapabilities() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        gc := ginadapter.GetGinContext(r)

        // 检查 S3 配置是否可用
        s3Configured := h.settingUC.Get(r.Context(), "s3_endpoint") != "" &&
            h.settingUC.Get(r.Context(), "s3_bucket") != "" &&
            h.settingUC.Get(r.Context(), "s3_access_key") != ""

        // 检查当前存储类型
        currentType := h.settingUC.Get(r.Context(), "storage_type")
        if currentType == "" {
            currentType = "local"
        }

        server.OK(gc, gin.H{
            "current_type":   currentType,
            "available_types": []string{"local"},
            "s3_configured":  s3Configured,
            "s3_available":   s3Configured, // S3 配置完整才可用
            "hybrid_available": s3Configured, // Hybrid 需要 S3
        })
    }
}
```

路由注册:

```go
// system_handler.go - registerSettings
settings.GET("/storage/capabilities", server.HTTPToHandlerFunc(h.getStorageCapabilities()))
```

#### 5.3.3 后端: NewStorageConfig 从 DB 读取

```go
func NewStorageConfig(settingUC *systembiz.SettingUseCase) *config.StorageConfig {
    cfg := config.DefaultStorageConfig()

    // 读取 base_path
    if basePath := settingUC.Get(context.Background(), "storage_base_path"); basePath != "" {
        cfg.BasePath = basePath
    }

    // 读取 storage_type
    if storageType := settingUC.Get(context.Background(), "storage_type"); storageType != "" {
        cfg.Type = config.StorageType(storageType)
    }

    // 读取 S3 配置
    if endpoint := settingUC.Get(context.Background(), "s3_endpoint"); endpoint != "" {
        cfg.S3.Endpoint = endpoint
    }
    if region := settingUC.Get(context.Background(), "s3_region"); region != "" {
        cfg.S3.Region = region
    }
    if bucket := settingUC.Get(context.Background(), "s3_bucket"); bucket != "" {
        cfg.S3.Bucket = bucket
    }
    if accessKey := settingUC.Get(context.Background(), "s3_access_key"); accessKey != "" {
        cfg.S3.AccessKey = accessKey
    }
    if secretKey := settingUC.Get(context.Background(), "s3_secret_key"); secretKey != "" {
        cfg.S3.SecretKey = secretKey
    }
    if usePathStyle := settingUC.GetBool(context.Background(), "s3_use_path_style"); usePathStyle {
        cfg.S3.UsePathStyle = true
    }

    return cfg
}
```

#### 5.3.4 前端: 修正存储类型选项

```tsx
// 存储类型选择器 - 新设计
const [storageCaps, setStorageCaps] = useState({
    s3_available: false,
    hybrid_available: false,
    current_type: 'local',
});

useEffect(() => {
    api.get('/system/settings/storage/capabilities').then(setStorageCaps);
}, []);

<select
    value={formData.storage_type}
    onChange={(e) => handleInputChange('storage_type', e.target.value)}
>
    <option value="local">本地存储</option>
    <option value="s3" disabled={!storageCaps.s3_available}>
        S3 对象存储 {!storageCaps.s3_available && '(未配置)'}
    </option>
    <option value="hybrid" disabled={!storageCaps.hybrid_available}>
        混合存储 {!storageCaps.hybrid_available && '(需要S3)'}
    </option>
</select>

{/* S3 配置区域 - 仅当 storage_type 为 s3 或 hybrid 时显示 */}
{(formData.storage_type === 's3' || formData.storage_type === 'hybrid') && (
    <Card>
        <CardHeader>
            <CardTitle>S3/MinIO 配置</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
            <Input label="Endpoint URL" ... />
            <Input label="Region" ... />
            <Input label="Bucket" ... />
            <Input label="Access Key" type="password" ... />
            <Input label="Secret Key" type="password" ... />
            <Switch label="Path-Style URLs (MinIO)" ... />
        </CardContent>
    </Card>
)}
```

### 5.4 涉及文件

| 文件 | 改动 |
|------|------|
| `internal/features/system/biz/setting.go` | DefaultSettings 添加 storage_type, s3_* |
| `internal/features/system/service/system_handler.go` | 添加 getStorageCapabilities 端点 |
| `cmd/server/wire.go` | NewStorageConfig 接受 ConfigProvider |
| `web/src/lib/api/system.ts` | 添加 storageCapabilities API |
| `web/src/pages/admin/Settings.tsx` | 修正存储类型选项, 动态禁用, 条件显示 S3 配置 |

---

## 6. 问题4: 站点信息未展示

### 6.1 根因

- Portal Header 硬编码 `"OrigStudio"`，不从配置读取
- HTML `<title>` 和 `<meta>` 标签不使用站点设置
- `usePortalConfig()` hook 已定义但从未被使用

### 6.2 当前 site_name 使用情况

| 组件 | 是否使用 site_name |
|------|-------------------|
| DocHeader | ✅ `site.site_name` |
| HeroSection | ✅ `site.site_name` + `site.site_description` |
| WelcomeLayout | ✅ `site.site_name` + `site.site_description` |
| **Portal Header** | ❌ 硬编码 `"OrigStudio"` |
| **HTML `<title>`** | ❌ 未设置 |
| **Meta 标签** | ❌ 未设置 |

### 6.3 修复方案

#### 6.3.1 前端: Portal Header 使用 site_name

```tsx
// components/portal/Header.tsx
const { site } = useModuleState();

<Link to="/" className="flex items-center gap-2 shrink-0">
    <img src="/logo.svg" alt={site.site_name || 'OrigStudio'} className="h-8 w-8" />
    <span className="text-lg font-bold text-foreground hidden sm:inline">
        {site.site_name || 'OrigStudio'}
    </span>
</Link>
```

#### 6.3.2 前端: HTML Head 管理

使用 TanStack Router 的 `head` 机制或 React Helmet:

**方案A: 在 _portal/route.tsx 中设置 (推荐)**

```tsx
// routes/_portal/route.tsx
export const Route = createFileRoute('/_portal')({
    component: PortalRouteComponent,
});

function PortalRouteComponent() {
    const { site } = useModuleState();

    useEffect(() => {
        document.title = site.site_name
            ? `${site.site_name} - ${site.site_description || 'CMS'}`
            : 'OrigStudio';

        // 设置 meta description
        let metaDesc = document.querySelector('meta[name="description"]');
        if (!metaDesc) {
            metaDesc = document.createElement('meta');
            metaDesc.setAttribute('name', 'description');
            document.head.appendChild(metaDesc);
        }
        if (site.site_description) {
            metaDesc.setAttribute('content', site.site_description);
        }
    }, [site.site_name, site.site_description]);

    return (
        <ModuleConfigProvider>
            <LayoutSwitcher />
        </ModuleConfigProvider>
    );
}
```

**方案B: 在 Admin Layout 中也设置**

```tsx
// routes/_authenticated/admin/route.tsx
// Admin 页面 title: "站点名 - 管理后台"
```

### 6.4 涉及文件

| 文件 | 改动 |
|------|------|
| `web/src/components/portal/Header.tsx` | 使用 site.site_name 替代硬编码 |
| `web/src/routes/_portal/route.tsx` | 添加 document.title 和 meta 标签管理 |
| `web/src/routes/_authenticated/route.tsx` | Admin 页面 title 管理 |

---

## 7. 问题5: 站点URL多地址支持

### 7.1 参考: MediaCMS 的做法

```python
# MediaCMS settings.py
ALLOWED_HOSTS = ["*", "mediacms.io", "127.0.0.1", "localhost"]
FRONTEND_HOST = "http://localhost"  # 主 URL (邮件/SEO)
SSL_FRONTEND_HOST = FRONTEND_HOST.replace("http", "https")

# 前端通过 request.build_absolute_uri('/') 动态获取当前 URL
```

### 7.2 设计方案

#### 7.2.1 后端: 设置变更

将 `base_url` (string) 替换为:

| Key | Type | Description |
|-----|------|-------------|
| `base_urls` | json | 允许访问的站点 URL 列表, `["https://example.com", "https://www.example.com"]` |
| `primary_url` | string | 主站点 URL, 用于邮件链接、SEO canonical URL |

```go
// DefaultSettings 新增/替换
{
    Key:           "base_urls",
    Value:         `[""]`,
    Type:          setting.TypeJSON,
    Category:      setting.CategoryGeneral,
    Description:   "Allowed site URLs (JSON array, for CORS and validation)",
    FallbackValue: `[""]`,
    IsBuiltin:     true,
},
{
    Key:           "primary_url",
    Value:         "",
    Type:          setting.TypeString,
    Category:      setting.CategoryGeneral,
    Description:   "Primary site URL for emails, SEO canonical URLs",
    FallbackValue: "",
    IsBuiltin:     true,
},
```

> **迁移**: 保留 `base_url` 设置以兼容，但在 `getPortalConfig` 中优先使用 `primary_url`。

#### 7.2.2 后端: Portal Config API 扩展

```go
type portalSiteConfig struct {
    SiteName          string   `json:"site_name"`
    SiteDescription   string   `json:"site_description"`
    PrimaryURL        string   `json:"primary_url"`
    AllowedURLs       []string `json:"allowed_urls"`
    AllowRegistration bool     `json:"allow_registration"`
    AllowUpload       bool     `json:"allow_upload"`
}
```

#### 7.2.3 后端: CORS 联动

在 CORS 中间件配置中读取 `base_urls`:

```go
// server.go - CORS 配置
func buildAllowedOrigins(settingUC *systembiz.SettingUseCase) []string {
    urls := settingUC.Get(context.Background(), "base_urls")
    var allowed []string
    if urls != "" {
        json.Unmarshal([]byte(urls), &allowed)
    }
    // 过滤空字符串
    var result []string
    for _, u := range allowed {
        if u != "" {
            result = append(result, u)
        }
    }
    if len(result) == 0 {
        result = []string{"*"} // 默认允许所有
    }
    return result
}
```

#### 7.2.4 前端: 多 URL 编辑器

```tsx
// 通用设置 - 站点URL区域
<div className="space-y-2">
    <label>站点URL</label>
    <p className="text-xs text-muted-foreground">
        允许访问此站点的URL列表，第一个将作为主URL用于邮件和SEO
    </p>
    {formData.base_urls.map((url, index) => (
        <div key={index} className="flex gap-2">
            <Input
                value={url}
                onChange={(e) => {
                    const newUrls = [...formData.base_urls];
                    newUrls[index] = e.target.value;
                    setFormData(prev => ({...prev, base_urls: newUrls}));
                }}
                placeholder="https://example.com"
            />
            {index === 0 && (
                <Badge variant="default">主URL</Badge>
            )}
            {formData.base_urls.length > 1 && (
                <Button variant="ghost" size="icon"
                    onClick={() => {
                        const newUrls = formData.base_urls.filter((_, i) => i !== index);
                        setFormData(prev => ({...prev, base_urls: newUrls}));
                    }}>
                    <X className="h-4 w-4" />
                </Button>
            )}
        </div>
    ))}
    <Button variant="outline" size="sm"
        onClick={() => setFormData(prev => ({...prev, base_urls: [...prev.base_urls, '']}))}>
        <Plus className="mr-2 h-4 w-4" /> 添加URL
    </Button>
</div>
```

### 7.3 涉及文件

| 文件 | 改动 |
|------|------|
| `internal/features/system/biz/setting.go` | DefaultSettings 添加 base_urls, primary_url |
| `internal/features/system/service/system_handler.go` | getPortalConfig 扩展 |
| `internal/server/server.go` | CORS 从设置读取 |
| `web/src/pages/admin/Settings.tsx` | 多URL编辑器 |
| `web/src/lib/api/portal.ts` | ModulePortalConfig 添加 primary_url, allowed_urls |
| `web/src/contexts/ModuleConfigContext.tsx` | 类型更新 |

---

## 8. 问题6: 媒体设置无效

### 8.1 根因

1. 前端媒体 TAB 字段与后端设置键名不匹配
2. 大量媒体设置在后端 DefaultSettings 中不存在
3. 后端 UploadUseCase 不从 ConfigProvider 读取设置

### 8.2 字段映射问题

| 前端字段 | 当前映射 | 正确映射 | 后端是否存在 |
|---------|---------|---------|------------|
| auto_transcode | → `auto_approve` | → `auto_transcode` (新) | ❌ 需新增 |
| transcode_method | 无 | → `transcode_method` (新) | ❌ 需新增 |
| video_formats | 无 | → `allowed_video_formats` (新) | ❌ 需新增 |
| image_formats | 无 | → `allowed_image_formats` (新) | ❌ 需新增 |
| max_duration | 无 | → `max_video_duration` (新) | ❌ 需新增 |
| max_file_size2 | 无 | → `max_upload_size_video` (已存在) | ✅ 重复 |

### 8.3 修复方案

#### 8.3.1 后端: 新增媒体设置

```go
// DefaultSettings 新增
{
    Key:           "auto_transcode",
    Value:         "true",
    Type:          setting.TypeBool,
    Category:      setting.CategoryUpload,
    Description:   "Automatically transcode uploaded videos",
    FallbackValue: "true",
    IsBuiltin:     true,
},
{
    Key:           "transcode_method",
    Value:         "ffmpeg",
    Type:          setting.TypeString,
    Category:      setting.CategoryUpload,
    Description:   "Transcode engine: ffmpeg",
    FallbackValue: "ffmpeg",
    IsBuiltin:     true,
},
{
    Key:           "allowed_video_formats",
    Value:         "mp4,webm,mkv,avi,mov",
    Type:          setting.TypeString,
    Category:      setting.CategoryUpload,
    Description:   "Allowed video file extensions (comma-separated)",
    FallbackValue: "mp4,webm,mkv,avi,mov",
    IsBuiltin:     true,
},
{
    Key:           "allowed_image_formats",
    Value:         "jpg,png,gif,webp",
    Type:          setting.TypeString,
    Category:      setting.CategoryUpload,
    Description:   "Allowed image file extensions (comma-separated)",
    FallbackValue: "jpg,png,gif,webp",
    IsBuiltin:     true,
},
{
    Key:           "max_video_duration",
    Value:         "7200",
    Type:          setting.TypeInt,
    Category:      setting.CategoryUpload,
    Description:   "Maximum video duration in seconds (0=unlimited)",
    FallbackValue: "7200",
    IsBuiltin:     true,
},
```

#### 8.3.2 后端: UploadUseCase 消费设置

```go
// media/biz/upload.go - 添加 ConfigProvider
type UploadUseCase struct {
    // ... existing fields
    configProvider systembiz.ConfigProvider
}

func NewUploadUseCase(
    // ... existing params
    configProvider systembiz.ConfigProvider,
) *UploadUseCase {
    // ...
}

// 在上传验证中使用设置
func (uc *UploadUseCase) validateUpload(ctx context.Context, req *UploadRequest) error {
    // 检查是否允许上传
    if !uc.configProvider.GetBool(ctx, "allow_upload") {
        return ErrUploadNotAllowed
    }

    // 检查文件大小
    maxSize := uc.configProvider.GetInt(ctx, "max_upload_size_video")
    if maxSize > 0 && req.Size > int64(maxSize) {
        return ErrFileTooLarge
    }

    // 检查文件格式
    allowedFormats := uc.configProvider.Get(ctx, "allowed_video_formats")
    if !isFormatAllowed(req.Filename, allowedFormats) {
        return ErrFormatNotAllowed
    }

    // 检查视频时长
    maxDuration := uc.configProvider.GetInt(ctx, "max_video_duration")
    if maxDuration > 0 && req.Duration > maxDuration {
        return ErrDurationTooLong
    }

    return nil
}
```

#### 8.3.3 前端: 修正字段映射

```typescript
// Settings.tsx - fetchSettings 映射
auto_transcode: getSettingValue('auto_transcode') || prev.auto_transcode,
transcode_method: getSettingValue('transcode_method') || prev.transcode_method,
video_formats: getSettingValue('allowed_video_formats') || prev.video_formats,
image_formats: getSettingValue('allowed_image_formats') || prev.image_formats,
max_duration: getSettingValue('max_video_duration') || prev.max_duration,

// Settings.tsx - handleSave 映射
{key: 'auto_transcode', value: formData.auto_transcode},
{key: 'transcode_method', value: formData.transcode_method},
{key: 'allowed_video_formats', value: formData.video_formats},
{key: 'allowed_image_formats', value: formData.image_formats},
{key: 'max_video_duration', value: formData.max_duration},
```

#### 8.3.4 前端: 媒体 TAB 重新设计

将媒体 TAB 拆分为更合理的分组:

- **转码设置**: auto_transcode, transcode_method
- **上传限制**: max_upload_size_video, max_upload_size_image, max_video_duration, allowed_video_formats, allowed_image_formats
- **缩略图设置**: thumbnail_quality, thumbnail_resolution, thumbnail_position (已生效)
- **雪碧图设置**: sprite_frame_interval, sprite_columns, sprite_frame_width, sprite_frame_height, sprite_max_frames (已生效)

### 8.4 涉及文件

| 文件 | 改动 |
|------|------|
| `internal/features/system/biz/setting.go` | DefaultSettings 添加媒体设置 |
| `internal/features/media/biz/upload.go` | 添加 ConfigProvider, 上传验证 |
| `cmd/server/wire.go` | NewUploadUseCase 传入 ConfigProvider |
| `web/src/pages/admin/Settings.tsx` | 修正字段映射, 重新设计媒体 TAB |

---

## 9. 问题7: 邮件发送缺失

### 9.1 根因

- SMTP 设置存在于数据库但零消费
- 没有邮件发送模块
- 没有邮件模板系统
- "发送测试邮件"按钮无后端实现

### 9.2 设计方案

#### 9.2.1 后端: 邮件发送模块

新增 `internal/features/system/biz/email.go`:

```go
package biz

import (
    "context"
    "crypto/tls"
    "fmt"
    "net/smtp"
    "strings"

    "gopkg.in/mail.v2"
)

type EmailUseCase struct {
    settingUC *SettingUseCase
}

func NewEmailUseCase(settingUC *SettingUseCase) *EmailUseCase {
    return &EmailUseCase{settingUC: settingUC}
}

// SendEmail 发送邮件
func (uc *EmailUseCase) SendEmail(ctx context.Context, to, subject, body string) error {
    host := uc.settingUC.Get(ctx, "smtp_host")
    if host == "" {
        return fmt.Errorf("SMTP not configured")
    }

    port := uc.settingUC.GetInt(ctx, "smtp_port")
    user := uc.settingUC.Get(ctx, "smtp_user")
    password := uc.settingUC.Get(ctx, "smtp_password")

    m := mail.NewMessage()
    m.SetHeader("From", fmt.Sprintf("%s <%s>", uc.settingUC.Get(ctx, "smtp_sender_name"), user))
    m.SetHeader("To", to)
    m.SetHeader("Subject", subject)
    m.SetBody("text/html", body)

    d := mail.NewDialer(host, port, user, password)
    if port == 465 {
        d.SSL = true
    } else {
        d.StartTLSPolicy = mail.MandatoryStartTLS
    }

    return d.DialAndSend(m)
}

// SendTestEmail 发送测试邮件
func (uc *EmailUseCase) SendTestEmail(ctx context.Context, to string) error {
    subject := "OrigStudio - 邮件测试"
    body := `<h2>邮件测试</h2><p>如果您收到此邮件，说明SMTP配置正确。</p>`
    return uc.SendEmail(ctx, to, subject, body)
}

// IsConfigured 检查SMTP是否已配置
func (uc *EmailUseCase) IsConfigured(ctx context.Context) bool {
    return uc.settingUC.Get(ctx, "smtp_host") != "" &&
        uc.settingUC.Get(ctx, "smtp_user") != ""
}
```

#### 9.2.2 后端: 邮件模板系统

新增 `internal/features/system/biz/email_template.go`:

```go
package biz

import "fmt"

// EmailTemplate 邮件模板
type EmailTemplate struct {
    Name     string
    Subject  string
    BodyHTML string
}

// 内置邮件模板
var builtinTemplates = map[string]EmailTemplate{
    "welcome": {
        Name:    "欢迎邮件",
        Subject: "欢迎加入 {{.SiteName}}",
        BodyHTML: `<h2>欢迎加入 {{.SiteName}}!</h2>
            <p>您好 {{.Username}},</p>
            <p>您的账户已成功创建。</p>
            <p><a href="{{.SiteURL}}">开始使用</a></p>`,
    },
    "email_verify": {
        Name:    "邮箱验证",
        Subject: "{{.SiteName}} - 验证您的邮箱",
        BodyHTML: `<h2>验证您的邮箱</h2>
            <p>请点击以下链接验证您的邮箱:</p>
            <p><a href="{{.VerifyURL}}">验证邮箱</a></p>
            <p>此链接将在24小时后失效。</p>`,
    },
    "password_reset": {
        Name:    "密码重置",
        Subject: "{{.SiteName}} - 重置密码",
        BodyHTML: `<h2>重置密码</h2>
            <p>请点击以下链接重置密码:</p>
            <p><a href="{{.ResetURL}}">重置密码</a></p>
            <p>此链接将在1小时后失效。</p>`,
    },
}

// RenderTemplate 渲染邮件模板
func (uc *EmailUseCase) RenderTemplate(templateName string, data map[string]string) (string, string, error) {
    tmpl, ok := builtinTemplates[templateName]
    if !ok {
        return "", "", fmt.Errorf("template not found: %s", templateName)
    }

    subject := tmpl.Subject
    body := tmpl.BodyHTML
    for k, v := range data {
        subject = strings.ReplaceAll(subject, "{{."+k+"}}", v)
        body = strings.ReplaceAll(body, "{{."+k+"}}", v)
    }
    return subject, body, nil
}
```

#### 9.2.3 后端: API 端点

```go
// system_handler.go - 新增

// POST /system/settings/email/test - 发送测试邮件
func (h *SystemHandler) sendTestEmail() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        gc := ginadapter.GetGinContext(r)
        var req struct {
            To string `json:"to" binding:"required,email"`
        }
        if err := gc.ShouldBindJSON(&req); err != nil {
            server.Fail(gc, server.ErrBadRequest, err.Error())
            return
        }
        if err := h.emailUC.SendTestEmail(r.Context(), req.To); err != nil {
            server.Fail(gc, server.ErrInternal, "Failed to send test email: "+err.Error())
            return
        }
        server.OK(gc, gin.H{"message": "Test email sent"})
    }
}

// GET /system/settings/email/status - 检查邮件配置状态
func (h *SystemHandler) getEmailStatus() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        gc := ginadapter.GetGinContext(r)
        server.OK(gc, gin.H{
            "configured": h.emailUC.IsConfigured(r.Context()),
        })
    }
}
```

#### 9.2.4 后端: DefaultSettings 补充

```go
{
    Key:           "smtp_sender_name",
    Value:         "OrigStudio",
    Type:          setting.TypeString,
    Category:      setting.CategoryEmail,
    Description:   "Sender display name for outgoing emails",
    FallbackValue: "OrigStudio",
    IsBuiltin:     true,
},
{
    Key:           "smtp_use_tls",
    Value:         "true",
    Type:          setting.TypeBool,
    Category:      setting.CategoryEmail,
    Description:   "Use TLS for SMTP connection",
    FallbackValue: "true",
    IsBuiltin:     true,
},
```

#### 9.2.5 前端: 邮件 TAB 增强

- "发送测试邮件" 按钮连接到 `POST /system/settings/email/test`
- 添加邮件配置状态指示（已配置/未配置）
- 邮件模板管理（后续迭代，当前仅显示内置模板列表）

### 9.3 依赖

```
go get gopkg.in/mail.v2
```

### 9.4 涉及文件

| 文件 | 改动 |
|------|------|
| `internal/features/system/biz/email.go` | **新建** - 邮件发送模块 |
| `internal/features/system/biz/email_template.go` | **新建** - 邮件模板 |
| `internal/features/system/service/system_handler.go` | 添加邮件测试端点 |
| `internal/features/system/biz/setting.go` | DefaultSettings 添加 smtp_sender_name, smtp_use_tls |
| `cmd/server/wire.go` | 添加 EmailUseCase 到 ProviderSet |
| `web/src/lib/api/system.ts` | 添加 emailTest API |
| `web/src/pages/admin/Settings.tsx` | 邮件 TAB 连接测试按钮 |

---

## 10. 问题8: 安全设置无效

### 10.1 根因

| 设置 | 问题 |
|------|------|
| `allow_registration` | 存在于 DB，但注册端点不检查 |
| `min_password_len` | 不存在于 DB，注册端点硬编码 `min=6` |
| `jwt_expiry` | 不存在于 DB，来自 YAML 配置 |
| `require_email_verify` | 不存在于 DB，无验证流程 |
| `enable_rest_api` | 不存在于 DB，无消费代码 |
| `rate_limit` | 不存在于 DB，无速率限制中间件 |

### 10.2 修复方案

#### 10.2.1 后端: 新增安全设置

```go
// DefaultSettings 新增
{
    Key:           "min_password_length",
    Value:         "8",
    Type:          setting.TypeInt,
    Category:      setting.CategoryGeneral,
    Description:   "Minimum password length for registration",
    FallbackValue: "8",
    IsBuiltin:     true,
},
{
    Key:           "require_email_verification",
    Value:         "false",
    Type:          setting.TypeBool,
    Category:      setting.CategoryGeneral,
    Description:   "Require email verification after registration",
    FallbackValue: "false",
    IsBuiltin:     true,
},
{
    Key:           "api_rate_limit",
    Value:         "60",
    Type:          setting.TypeInt,
    Category:      setting.CategoryGeneral,
    Description:   "API rate limit (requests per minute per IP, 0=unlimited)",
    FallbackValue: "60",
    IsBuiltin:     true,
},
```

#### 10.2.2 后端: AuthHandler 检查 allow_registration

```go
// auth_handler.go - registerUser
func (h *AuthHandler) registerUser() http2.HandlerFunc {
    return func(ctx http2.Context) error {
        // 检查是否允许注册
        if h.settingUC != nil && !h.settingUC.GetBool(ctx.Request().Context(), "allow_registration") {
            http2.Fail(ctx, server.ErrForbidden, "registration is disabled")
            return nil
        }

        // 检查密码最小长度
        minLen := 6 // 默认
        if h.settingUC != nil {
            if v := h.settingUC.GetInt(ctx.Request().Context(), "min_password_length"); v > 0 {
                minLen = v
            }
        }

        var req struct {
            Username string `json:"username" binding:"required,min=3,max=64"`
            Password string `json:"password" binding:"required,max=128"`
            Email    string `json:"email"    binding:"omitempty,email"`
            Nickname string `json:"nickname"`
        }
        // ... 绑定 JSON ...

        // 动态密码长度验证
        if len(req.Password) < minLen {
            http2.Fail(ctx, server.ErrBadRequest,
                fmt.Sprintf("password must be at least %d characters", minLen))
            return nil
        }

        // ... 其余注册逻辑 ...
    }
}
```

> **注意**: `AuthHandler` 需要注入 `ConfigProvider` (通过 wire)。

#### 10.2.3 后端: 速率限制中间件

新增 `internal/middleware/ratelimit.go`:

```go
package middleware

import (
    "net/http"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    visitors map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}

func NewRateLimiter(rpm int) *RateLimiter {
    if rpm <= 0 {
        rpm = 60 // 默认 60/分钟
    }
    return &RateLimiter{
        visitors: make(map[string]*rate.Limiter),
        rate:     rate.Every(time.Minute / time.Duration(rpm)),
        burst:    rpm,
    }
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    limiter, exists := rl.visitors[ip]
    if !exists {
        limiter = rate.NewLimiter(rl.rate, rl.burst)
        rl.visitors[ip] = limiter
    }
    return limiter
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
    // 定期清理过期 visitors
    go func() {
        for {
            time.Sleep(10 * time.Minute)
            rl.mu.Lock()
            rl.visitors = make(map[string]*rate.Limiter)
            rl.mu.Unlock()
        }
    }()

    return func(c *gin.Context) {
        ip := c.ClientIP()
        limiter := rl.getVisitor(ip)
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "code":    429,
                "message": "rate limit exceeded",
            })
            c.Abort()
            return
        }
        c.Next()
    }
}
```

在 `server.go` 中注册:

```go
// 从设置读取速率限制
rpm := settingUC.GetInt(context.Background(), "api_rate_limit")
if rpm > 0 {
    r.Use(NewRateLimiter(rpm).Middleware())
}
```

#### 10.2.4 前端: 安全 TAB 修正

```typescript
// FormData 修正
interface FormData {
    // ... 其他字段 ...
    allow_registration: string;   // 原 enable_register
    require_email_verify: string; // 映射到 require_email_verification
    min_password_len: string;     // 映射到 min_password_length
    api_rate_limit: string;       // 映射到 api_rate_limit
    // 删除: jwt_expiry (来自YAML, 不在DB设置中)
    // 删除: enable_rest_api (无后端消费)
}

// fetchSettings 映射
allow_registration: getSettingValue('allow_registration') || prev.allow_registration,
require_email_verify: getSettingValue('require_email_verification') || prev.require_email_verify,
min_password_len: getSettingValue('min_password_length') || prev.min_password_len,
api_rate_limit: getSettingValue('api_rate_limit') || prev.api_rate_limit,

// handleSave 映射
{key: 'allow_registration', value: formData.allow_registration},
{key: 'require_email_verification', value: formData.require_email_verify},
{key: 'min_password_length', value: formData.min_password_len},
{key: 'api_rate_limit', value: formData.api_rate_limit},
```

**UI 调整**:
- 移除 `JWT 过期时间` 字段（由 YAML 配置管理，不适合在 UI 修改）
- 移除 `启用 REST API` 字段（无实际意义）
- 将 `速率限制` 移到安全 TAB

### 10.3 涉及文件

| 文件 | 改动 |
|------|------|
| `internal/features/system/biz/setting.go` | DefaultSettings 添加安全设置 |
| `internal/features/auth/service/auth_handler.go` | 注入 ConfigProvider, 检查注册/密码 |
| `internal/middleware/ratelimit.go` | **新建** - 速率限制中间件 |
| `internal/server/server.go` | 注册速率限制中间件 |
| `cmd/server/wire.go` | AuthHandler 注入 ConfigProvider |
| `web/src/pages/admin/Settings.tsx` | 修正安全 TAB 字段映射 |

---

## 11. 数据库Schema变更

### 11.1 Category 枚举

当前: `general | upload | review | email | module`

**不需要变更**。新增的设置项都归入现有 category:
- `storage_*`, `s3_*` → `upload`
- `min_password_length`, `api_rate_limit`, `require_email_verification` → `general`
- `base_urls`, `primary_url` → `general`

### 11.2 新增设置项汇总

| Key | Type | Category | Sensitive |
|-----|------|----------|-----------|
| storage_base_path | string | upload | no |
| storage_type | string | upload | no |
| s3_endpoint | string | upload | no |
| s3_region | string | upload | no |
| s3_bucket | string | upload | no |
| s3_access_key | string | upload | **yes** |
| s3_secret_key | string | upload | **yes** |
| s3_use_path_style | bool | upload | no |
| base_urls | json | general | no |
| primary_url | string | general | no |
| auto_transcode | bool | upload | no |
| transcode_method | string | upload | no |
| allowed_video_formats | string | upload | no |
| allowed_image_formats | string | upload | no |
| max_video_duration | int | upload | no |
| smtp_sender_name | string | email | no |
| smtp_use_tls | bool | email | no |
| min_password_length | int | general | no |
| require_email_verification | bool | general | no |
| api_rate_limit | int | general | no |

**总计新增: 20 项** (从 25 → 45 项)

---

## 12. DefaultSettings完整清单

修复后的完整 DefaultSettings 列表 (45 项):

### general (8项)
1. site_name (string)
2. site_description (string)
3. base_url (string) — 保留兼容
4. base_urls (json) — 新增
5. primary_url (string) — 新增
6. allow_registration (bool)
7. min_password_length (int) — 新增
8. require_email_verification (bool) — 新增
9. api_rate_limit (int) — 新增

### upload (21项)
10. allow_upload (bool)
11. storage_base_path (string) — 新增
12. storage_type (string) — 新增
13. s3_endpoint (string) — 新增
14. s3_region (string) — 新增
15. s3_bucket (string) — 新增
16. s3_access_key (string, sensitive) — 新增
17. s3_secret_key (string, sensitive) — 新增
18. s3_use_path_style (bool) — 新增
19. max_upload_size_video (int)
20. max_upload_size_image (int)
21. auto_transcode (bool) — 新增
22. transcode_method (string) — 新增
23. allowed_video_formats (string) — 新增
24. allowed_image_formats (string) — 新增
25. max_video_duration (int) — 新增
26. sprite_frame_interval (int)
27. sprite_columns (int)
28. sprite_frame_width (int)
29. sprite_frame_height (int)
30. sprite_max_frames (int)
31. thumbnail_quality (int)
32. thumbnail_resolution (string)
33. thumbnail_position (string)

### review (2项)
34. auto_approve (bool)
35. require_review (bool)

### email (6项)
36. smtp_host (string)
37. smtp_port (int)
38. smtp_user (string)
39. smtp_password (string, sensitive)
40. smtp_sender_name (string) — 新增
41. smtp_use_tls (bool) — 新增

### module (4项)
42. module_articles (bool)
43. module_videos (bool)
44. module_music (bool)
45. homepage_layout (string)

---

## 13. 实施顺序

### Phase 1: 基础修复 (前端Bug, 无后端改动)

| 步骤 | 任务 | 预估复杂度 |
|------|------|-----------|
| 1.1 | 修复系统TAB: fetchSystemInfo 使用 api 封装 | 低 |
| 1.2 | 修正存储类型选项: local/s3/hybrid | 低 |
| 1.3 | Portal Header 使用 site_name | 低 |

### Phase 2: 后端设置扩展 (DefaultSettings + 消费)

| 步骤 | 任务 | 预估复杂度 |
|------|------|-----------|
| 2.1 | DefaultSettings 添加全部 20 个新设置项 | 中 |
| 2.2 | NewStorageConfig 从 DB 读取 | 中 |
| 2.3 | Storage Capabilities API | 中 |
| 2.4 | AuthHandler 注入 ConfigProvider, 检查注册/密码 | 中 |
| 2.5 | UploadUseCase 注入 ConfigProvider, 验证上传 | 中 |
| 2.6 | 速率限制中间件 | 中 |
| 2.7 | Wire 依赖注入更新 | 中 |

### Phase 3: 前端全面对接

| 步骤 | 任务 | 预估复杂度 |
|------|------|-----------|
| 3.1 | Settings.tsx FormData 重构, 键名对齐 | 高 |
| 3.2 | 存储TAB重新设计 (BasePath + S3条件显示) | 中 |
| 3.3 | 媒体TAB重新设计 (分组 + 正确映射) | 中 |
| 3.4 | 安全TAB修正 (移除无效字段 + 正确映射) | 中 |
| 3.5 | 多URL编辑器 | 中 |
| 3.6 | HTML title/meta 管理 | 低 |

### Phase 4: 邮件模块 (独立功能)

| 步骤 | 任务 | 预估复杂度 |
|------|------|-----------|
| 4.1 | EmailUseCase 实现 | 中 |
| 4.2 | 邮件模板系统 | 中 |
| 4.3 | 测试邮件 API | 低 |
| 4.4 | 前端邮件TAB增强 | 中 |

### Phase 5: 验证与清理

| 步骤 | 任务 | 预估复杂度 |
|------|------|-----------|
| 5.1 | 删除 usePortalConfig 死代码 | 低 |
| 5.2 | 修复 comment_moderation 键名不匹配 | 低 |
| 5.3 | 移除 MediaReportUseCase 未使用的 ConfigProvider | 低 |
| 5.4 | 端到端测试 | 中 |
