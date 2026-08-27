package biz

import (
	"strings"
	"testing"
)

func TestConvertSRTToVTT_Valid(t *testing.T) {
	srt := "1\n00:00:01,000 --> 00:00:04,000\nHello world\n\n2\n00:00:05,500 --> 00:00:08,000\n你好，世界\n"
	out, err := ConvertSRTToVTT([]byte(srt))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := string(out)
	if !strings.HasPrefix(text, "WEBVTT") {
		t.Errorf("missing WEBVTT header: %q", text[:20])
	}
	if !strings.Contains(text, "00:00:01.000 --> 00:00:04.000") {
		t.Errorf("comma millis not converted to dot: %q", text)
	}
	if !strings.Contains(text, "Hello world") || !strings.Contains(text, "你好，世界") {
		t.Errorf("text lines missing: %q", text)
	}
}

func TestConvertSRTToVTT_TimeFormatError(t *testing.T) {
	// 第 2 行时间轴用错误分隔符（单箭头/少毫秒）
	srt := "1\n00:00:01,00 --> 00:00:04,000\nbad\n"
	_, err := ConvertSRTToVTT([]byte(srt))
	if err == nil {
		t.Fatal("expected error for malformed time line")
	}
	if !strings.Contains(err.Error(), "第 2 行") {
		t.Errorf("error should carry line number 2, got: %v", err)
	}
	if !strings.Contains(err.Error(), "HH:MM:SS,mmm") {
		t.Errorf("error should hint expected format, got: %v", err)
	}
}

func TestConvertSRTToVTT_EndBeforeStart(t *testing.T) {
	srt := "1\n00:00:05,000 --> 00:00:02,000\nreversed\n"
	_, err := ConvertSRTToVTT([]byte(srt))
	if err == nil {
		t.Fatal("expected error for end < start")
	}
	if !strings.Contains(err.Error(), "结束时间早于开始时间") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNormalizeSubtitle_VTT(t *testing.T) {
	vtt := "WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nhi\n"
	out, err := NormalizeSubtitle([]byte(vtt), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != vtt {
		t.Errorf("vtt should pass through unchanged")
	}
}

func TestNormalizeSubtitle_VTTMissingHeader(t *testing.T) {
	_, err := NormalizeSubtitle([]byte("00:00:01.000 --> 00:00:02.000\nhi\n"), true)
	if err == nil || !strings.Contains(err.Error(), "WEBVTT") {
		t.Errorf("expected WEBVTT header error, got: %v", err)
	}
}

func TestConvertSRTToVTT_Empty(t *testing.T) {
	out, err := ConvertSRTToVTT([]byte(""))
	if err != nil {
		t.Fatalf("empty srt should convert to bare WEBVTT: %v", err)
	}
	if string(out) != "WEBVTT\n\n" {
		t.Errorf("unexpected empty output: %q", string(out))
	}
}
