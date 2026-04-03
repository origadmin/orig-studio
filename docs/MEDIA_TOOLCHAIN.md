# 媒体处理工具链文档 (Media Toolchain)

本项目使用 `FFmpeg` 和 `Bento4` 作为视频转码和流媒体封装的核心工具。

## 1. 核心依赖

| 工具 | 用途 | 对应二进制文件 |
|------|------|---------------|
| **FFmpeg** | 视频/音频编码、缩略图提取 | `ffmpeg.exe` |
| **FFprobe** | 媒体信息提取（时长、分辨率等） | `ffprobe.exe` |
| **Bento4** | HLS/fMP4 封装、Master Playlist 生成 | `mp4mux.exe`, `mp4hls.exe`, `mp4info.exe` |

---

## 2. 目录结构与路径探测

系统对外部工具的路径探测遵循以下优先级逻辑：

1. **Backend Settings (未来)**：从数据库 `settings` 表中读取自定义路径（优先级最高）。
2. **Environment Variables**：检查环境变量 `FFMPEG_PATH` 或 `BENTO4_PATH`。
3. **Local Tools Directory**：检查项目根目录下的 `tools/bin/` 目录。
4. **System PATH**：直接调用系统已安装的工具。

### 推荐目录结构 (Windows)
```text
D:\workspace\...\orig-cms\
└── tools\
    └── bin\
        ├── ffmpeg.exe
        ├── ffprobe.exe
        ├── mp4mux.exe
        ├── mp4hls.exe
        └── mp4info.exe
```

---

## 3. 配置文件配置项

在 `configs/config.yaml` 中，可以预留以下配置：

```yaml
media:
  tools:
    # 留空则使用上述自动探测逻辑
    ffmpeg_path: "D:/tools/bin/ffmpeg.exe"
    ffprobe_path: "D:/tools/bin/ffprobe.exe"
    bento4_bin_dir: "D:/tools/bin/"
```

---

## 4. 后台动态识别方案

系统启动时，会运行 `internal/helpers/ffmpeg` 中的初始化逻辑：
1. 扫描 `tools/bin` 目录下的所有支持的二进制文件。
2. 校验文件执行权限和版本信息。
3. 如果识别失败，系统将限制部分转码功能，但依然支持普通文件上传。

---

*文档版本：v1.0 | 2026-04-02*
