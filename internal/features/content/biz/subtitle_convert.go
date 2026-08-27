package biz

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Subtitle converter — BUG-186: normalize uploaded .srt/.vtt to WebVTT.
//
// G5 decisions:
//   - single storage format: everything is stored as .vtt
//   - failure reports are line-level ("第 N 行: ...") so the UI can tell the
//     user exactly which line to fix
//   - only .srt / .vtt accepted; .ass/.ssa/.lrc/.sub/.idx/.ttml rejected

var (
	// srtTimeRe matches "00:00:01,000 --> 00:00:04,000" (comma millis).
	srtTimeRe = regexp.MustCompile(`^(\d{2}):(\d{2}):(\d{2}),(\d{3})\s*-->\s*(\d{2}):(\d{2}):(\d{2}),(\d{3})\s*$`)
	// vttTimeRe matches "00:00:01.000 --> 00:00:04.000" (dot millis).
	vttTimeRe = regexp.MustCompile(`^(\d{2}):(\d{2}):(\d{2})\.(\d{3})\s*-->\s*(\d{2}):(\d{2}):(\d{2})\.(\d{3})\s*$`)
)

// NormalizeSubtitle converts uploaded subtitle bytes to a .vtt file.
//   - srt content -> converted to vtt (line-level validation)
//   - vtt content  -> returned as-is (header validated)
// Returns error with line number when malformed.
func NormalizeSubtitle(content []byte, isVTT bool) ([]byte, error) {
	if isVTT {
		return normalizeVTT(content)
	}
	return ConvertSRTToVTT(content)
}

func normalizeVTT(content []byte) ([]byte, error) {
	text := strings.ReplaceAll(string(content), "\r\n", "\n")
	if !strings.HasPrefix(strings.TrimSpace(text), "WEBVTT") {
		return nil, fmt.Errorf("文件不是有效的 VTT 字幕（缺少 WEBVTT 头）")
	}
	// validate every time line uses dot millis
	for i, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "-->") && !vttTimeRe.MatchString(trimmed) {
			return nil, fmt.Errorf("第 %d 行: 时间轴格式错误 '%s'，应为 HH:MM:SS.mmm --> HH:MM:SS.mmm", i+1, trimmed)
		}
	}
	return []byte(text), nil
}

// ConvertSRTToVTT converts SRT content to WebVTT with line-level validation.
func ConvertSRTToVTT(srt []byte) ([]byte, error) {
	lines := strings.Split(strings.ReplaceAll(string(srt), "\r\n", "\n"), "\n")
	var b strings.Builder
	b.WriteString("WEBVTT\n\n")

	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		i++
		if line == "" {
			continue
		}
		// optional index line (digits); if it's a time line directly, handle it
		if _, err := strconv.Atoi(line); err == nil {
			if i >= len(lines) {
				return nil, fmt.Errorf("第 %d 行: 缺少时间轴行（应为 序号 → 时间轴 → 文本）", i)
			}
			tl := strings.TrimSpace(lines[i])
			m := srtTimeRe.FindStringSubmatch(tl)
			if m == nil {
				return nil, fmt.Errorf("第 %d 行: 时间轴格式错误 '%s'，应为 HH:MM:SS,mmm --> HH:MM:SS,mmm", i+1, tl)
			}
			start, end := srtTimeToVTT(m)
			if ms(end) < ms(start) {
				return nil, fmt.Errorf("第 %d 行: 结束时间早于开始时间", i+1)
			}
			b.WriteString(start + " --> " + end + "\n")
			i++
			// text lines until blank
			for i < len(lines) {
				txt := strings.TrimSpace(lines[i])
				if txt == "" {
					break
				}
				b.WriteString(txt + "\n")
				i++
			}
			b.WriteString("\n")
			continue
		}
		if m := srtTimeRe.FindStringSubmatch(line); m != nil {
			// time line without an index line (some SRT variants)
			start, end := srtTimeToVTT(m)
			if ms(end) < ms(start) {
				return nil, fmt.Errorf("第 %d 行: 结束时间早于开始时间", i)
			}
			b.WriteString(start + " --> " + end + "\n")
			for i < len(lines) {
				txt := strings.TrimSpace(lines[i])
				if txt == "" {
					break
				}
				b.WriteString(txt + "\n")
				i++
			}
			b.WriteString("\n")
			continue
		}
		return nil, fmt.Errorf("第 %d 行: 无法解析（应为 序号 / 时间轴 / 文本），内容 '%s'", i, truncate(line, 40))
	}
	return []byte(b.String()), nil
}

func srtTimeToVTT(m []string) (string, string) {
	start := fmt.Sprintf("%s:%s:%s.%s", m[1], m[2], m[3], m[4])
	end := fmt.Sprintf("%s:%s:%s.%s", m[5], m[6], m[7], m[8])
	return start, end
}

// ms converts "HH:MM:SS.mmm" to total milliseconds for ordering checks.
func ms(t string) int64 {
	parts := strings.Split(t, ":")
	if len(parts) != 3 {
		return 0
	}
	h, _ := strconv.ParseInt(parts[0], 10, 64)
	mi, _ := strconv.ParseInt(parts[1], 10, 64)
	sec := strings.Split(parts[2], ".")
	s, _ := strconv.ParseInt(sec[0], 10, 64)
	ml := int64(0)
	if len(sec) > 1 {
		ml, _ = strconv.ParseInt(sec[1], 10, 64)
	}
	return h*3600000 + mi*60000 + s*1000 + ml
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
