package data

import (
	"context"
	"log/slog"

	"origadmin/application/origcms/internal/data/entity"
)

// SeedEncodeProfiles ensures default encoding profiles exist in the system.
func SeedEncodeProfiles(ctx context.Context, client *entity.Client) error {
	count, err := client.EncodeProfile.Query().Count(ctx)
	if err != nil {
		return err
	}

	if count > 0 {
		slog.Info("Encode profiles already seeded, skipping")
		return nil
	}

	// Exactly 22 profiles matching the MotoPlayer requirement table
	// RESOLUTION and CODEC are stored EXACTLY as provided for Bento4 usage
	profiles := []struct {
		Name        string
		Ext         string
		Res         string
		Codec       string
		Active      bool
		BentoParams string
	}{
		{"h265-240", "mp4", "240", "h265", true, "--video-bitrate 400k --audio-bitrate 64k"},
		{"vp9-240", "webm", "240", "vp9", false, "--video-bitrate 350k --audio-bitrate 64k"},
		{"h264-240", "mp4", "240", "h264", true, "--video-bitrate 500k --audio-bitrate 64k"},

		{"h265-360", "mp4", "360", "h265", true, "--video-bitrate 800k --audio-bitrate 96k"},
		{"vp9-360", "webm", "360", "vp9", false, "--video-bitrate 700k --audio-bitrate 96k"},
		{"h264-360", "mp4", "360", "h264", true, "--video-bitrate 1000k --audio-bitrate 96k"},

		{"h265-480", "mp4", "480", "h265", true, "--video-bitrate 1500k --audio-bitrate 128k"},
		{"vp9-480", "webm", "480", "vp9", false, "--video-bitrate 1200k --audio-bitrate 128k"},
		{"h264-480", "mp4", "480", "h264", true, "--video-bitrate 2000k --audio-bitrate 128k"},

		{"h265-720", "mp4", "720", "h265", true, "--video-bitrate 3500k --audio-bitrate 128k"},
		{"vp9-720", "webm", "720", "vp9", false, "--video-bitrate 3000k --audio-bitrate 128k"},
		{"h264-720", "mp4", "720", "h264", true, "--video-bitrate 4000k --audio-bitrate 128k"},

		{"h265-1080", "mp4", "1080", "h265", true, "--video-bitrate 6500k --audio-bitrate 192k"},
		{"vp9-1080", "webm", "1080", "vp9", false, "--video-bitrate 6000k --audio-bitrate 192k"},
		{"h264-1080", "mp4", "1080", "h264", true, "--video-bitrate 8000k --audio-bitrate 192k"},

		{"h265-1440", "mp4", "1440", "h265", false, "--video-bitrate 12000k --audio-bitrate 256k"},
		{"vp9-1440", "webm", "1440", "vp9", false, "--video-bitrate 10000k --audio-bitrate 256k"},
		{"h264-1440", "mp4", "1440", "h264", false, "--video-bitrate 15000k --audio-bitrate 256k"},

		{"vp9-2160", "webm", "2160", "vp9", false, "--video-bitrate 20000k --audio-bitrate 320k"},
		{"h264-2160", "mp4", "2160", "h264", false, "--video-bitrate 25000k --audio-bitrate 320k"},
		{"h265-2160", "mp4", "2160", "h265", false, "--video-bitrate 22000k --audio-bitrate 320k"},

		{"preview", "gif", "-", "-", true, "--fps 10 --scale 320"},
	}

	for _, p := range profiles {
		_, err := client.EncodeProfile.Create().
			SetName(p.Name).
			SetDescription("MotoPlayer specific profile").
			SetResolution(p.Res).
			SetExtension(p.Ext).
			SetVideoCodec(p.Codec).
			SetVideoBitrate("auto").
			SetAudioCodec("aac").
			SetAudioBitrate("128k").
			SetIsActive(p.Active).
			// SetBentoParameters(p.BentoParams).
			Save(ctx)
		if err != nil {
			slog.Error("failed to seed encode profile", "name", p.Name, "err", err)
			return err
		}
	}

	slog.Info("Successfully seeded 22 Bento4-ready profiles")
	return nil
}
