package dal

import (
	"context"
	"fmt"
	"log/slog"
	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/encodeprofile"
	"strings"
)

// SeedEncodeProfiles ensures default encoding profiles exist in the system.
func SeedEncodeProfiles(ctx context.Context, client *entity.Client) error {
	profiles := []struct {
		Name        string
		Ext         string
		Res         string
		Codec       string
		Active      bool
		BentoParams string
	}{
		{"h264-360", "mp4", "360", "h264", true, "--video-bitrate 1000k --audio-bitrate 96k"},
		{"h264-720", "mp4", "720", "h264", true, "--video-bitrate 4000k --audio-bitrate 128k"},
		{"h264-1080", "mp4", "1080", "h264", true, "--video-bitrate 8000k --audio-bitrate 192k"},

		{"h264-240", "mp4", "240", "h264", false, "--video-bitrate 500k --audio-bitrate 64k"},
		{"h264-480", "mp4", "480", "h264", false, "--video-bitrate 2000k --audio-bitrate 128k"},
		{"h264-1440", "mp4", "1440", "h264", false, "--video-bitrate 15000k --audio-bitrate 256k"},
		{"h264-2160", "mp4", "2160", "h264", false, "--video-bitrate 25000k --audio-bitrate 320k"},

		{"h265-360", "mp4", "360", "h265", false, "--video-bitrate 800k --audio-bitrate 96k"},
		{"h265-720", "mp4", "720", "h265", false, "--video-bitrate 3500k --audio-bitrate 128k"},
		{"h265-1080", "mp4", "1080", "h265", false, "--video-bitrate 6500k --audio-bitrate 192k"},
		{"h265-240", "mp4", "240", "h265", false, "--video-bitrate 400k --audio-bitrate 64k"},
		{"h265-480", "mp4", "480", "h265", false, "--video-bitrate 1500k --audio-bitrate 128k"},
		{"h265-1440", "mp4", "1440", "h265", false, "--video-bitrate 12000k --audio-bitrate 256k"},
		{"h265-2160", "mp4", "2160", "h265", false, "--video-bitrate 22000k --audio-bitrate 320k"},

		{"vp9-360", "webm", "360", "vp9", false, "--video-bitrate 700k --audio-bitrate 96k"},
		{"vp9-720", "webm", "720", "vp9", false, "--video-bitrate 3000k --audio-bitrate 128k"},
		{"vp9-1080", "webm", "1080", "vp9", false, "--video-bitrate 6000k --audio-bitrate 192k"},
		{"vp9-240", "webm", "240", "vp9", false, "--video-bitrate 350k --audio-bitrate 64k"},
		{"vp9-480", "webm", "480", "vp9", false, "--video-bitrate 1200k --audio-bitrate 128k"},
		{"vp9-1440", "webm", "1440", "vp9", false, "--video-bitrate 10000k --audio-bitrate 256k"},
		{"vp9-2160", "webm", "2160", "vp9", false, "--video-bitrate 20000k --audio-bitrate 320k"},

		{"preview", "gif", "-", "-", true, "--fps 10 --scale 320"},
	}

	var seeded int
	for _, p := range profiles {
		exists, err := client.EncodeProfile.Query().Where(encodeprofile.NameEQ(p.Name)).Exist(ctx)
		if err != nil {
			return fmt.Errorf("check profile %s: %w", p.Name, err)
		}
		if exists {
			continue
		}
		audioCodec := "aac"
		if p.Ext == "gif" || p.Codec == "-" {
			audioCodec = ""
		}
		_, err = client.EncodeProfile.Create().
			SetName(p.Name).
			SetDescription(profileDescription(p.Name, p.Codec, p.Ext, p.Res)).
			SetResolution(p.Res).
			SetExtension(p.Ext).
			SetVideoCodec(p.Codec).
			SetVideoBitrate(extractBentoBitrate(p.BentoParams, "video")).
			SetAudioCodec(audioCodec).
			SetAudioBitrate(extractBentoBitrate(p.BentoParams, "audio")).
			SetBentoParameters(p.BentoParams).
			SetIsActive(p.Active).
			Save(ctx)
		if err != nil {
			slog.Error("failed to seed encode profile", "name", p.Name, "err", err)
			return err
		}
		seeded++
	}

	if seeded > 0 {
		slog.Info("Seeded encode profiles", "count", seeded)
	} else {
		slog.Info("All encode profiles already exist, skipping seed")
	}
	return nil
}

// extractBentoBitrate extracts bitrate value from Bento4 parameter string.
// e.g., "--video-bitrate 400k --audio-bitrate 64k" → "400k" (kind=video) or "64k" (kind=audio)
func extractBentoBitrate(bentoParams, kind string) string {
	if bentoParams == "" {
		return ""
	}
	parts := strings.Fields(bentoParams)
	prefix := "--" + kind + "-bitrate"
	for i, p := range parts {
		if p == prefix && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func profileDescription(name, codec, ext, res string) string {
	if name == "preview" {
		return "Preview GIF generation profile"
	}
	var codecDisplay string
	switch codec {
	case "h264":
		codecDisplay = "H.264"
	case "h265":
		codecDisplay = "H.265"
	case "vp9":
		codecDisplay = "VP9"
	default:
		codecDisplay = strings.ToUpper(codec)
	}
	var extDisplay string
	switch ext {
	case "mp4":
		extDisplay = "MP4"
	case "webm":
		extDisplay = "WebM"
	default:
		extDisplay = strings.ToUpper(ext)
	}
	return fmt.Sprintf("%s %sp %s encoding profile", codecDisplay, res, extDisplay)
}
