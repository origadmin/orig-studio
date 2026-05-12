# F001: R1 — Data Model

## New: ffmpeg.MediaInfo struct

File: `internal/helpers/ffmpeg/ffmpeg.go`

```go
type MediaInfo struct {
    Duration     float64 // seconds
    Size         int64   // bytes
    BitRate      int64   // bps (format-level)
    FormatName   string  // e.g. "mov,mp4,m4a"
    Width        int     // video stream
    Height       int     // video stream
    VideoCodec   string  // e.g. "h264"
    VideoBitRate int64   // bps (stream-level)
    FPS          float64 // e.g. 29.97
    AudioCodec   string  // e.g. "aac"
    AudioBitRate int64   // bps
    Channels     int     // e.g. 2
    SampleRate   int     // e.g. 48000
}
```

## ffprobe JSON schema (internal)

```go
type ffprobeOutput struct {
    Streams []ffprobeStream `json:"streams"`
    Format  ffprobeFormat   `json:"format"`
}
type ffprobeStream struct {
    CodecType  string `json:"codec_type"`  // "video" | "audio"
    CodecName  string `json:"codec_name"`
    Width      int    `json:"width"`
    Height     int    `json:"height"`
    RFrameRate string `json:"r_frame_rate"` // "30000/1001"
    BitRate    string `json:"bit_rate"`
    Channels   int    `json:"channels"`
    SampleRate string `json:"sample_rate"`
}
type ffprobeFormat struct {
    Duration   string `json:"duration"`
    Size       string `json:"size"`
    BitRate    string `json:"bit_rate"`
    FormatName string `json:"format_name"`
}
```

## Media entity fields populated

| Field | Source | Previously |
|-------|--------|-----------|
| width (int32) | MediaInfo.Width | unfilled |
| height (int32) | MediaInfo.Height | unfilled |
| duration (int32) | MediaInfo.Duration | ffmpeg.GetVideoDuration |
| size (string) | session.FileSize | (unchanged) |
| extension (string) | filepath.Ext(session.Filename) | unfilled |
| mime_type (string) | session.ContentType | (unchanged) |
| url (string) | PromoteToOriginal result | was temp/... |

## UploadSession changes

- StoragePath now records promoted (originals) path instead of temp path