package ffmpeg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DEPRECATED: MP4Mux — Bento4 mp4mux is no longer used.
// All muxing is handled by ffmpeg (TranscodeToHLS outputs HLS directly).

// VariantInfo holds metadata about a transcoded variant for master playlist generation.
type VariantInfo struct {
	Path       string // relative path to variant's index.m3u8 (e.g., "h264_720/index.m3u8")
	Bandwidth  int    // bandwidth in bits per second
	Resolution string // e.g., "1280x720" or empty
	Name       string // profile name (e.g., "h264_720")
}

// DEPRECATED: MP4HLS — old two-step pipeline (MP4 → HLS segments).
// Replaced by TranscodeToHLS which outputs directly to HLS in one pass + GenerateMasterPlaylist.
// Kept for reference; delete when fully migrated.
func MP4HLS(_ interface{}, _ string, _ []string, _ []VariantInfo) error {
	return fmt.Errorf("MP4HLS is deprecated; use TranscodeToHLS + GenerateMasterPlaylist")
}

// GenerateMasterPlaylist creates an HLS master playlist that references all
// successful variant playlists. Each variant is expected to have been produced
// by TranscodeToHLS into its own subdirectory under hlsBaseDir.
//
// Expected directory structure:
//
//	hlsBaseDir/ (e.g., hls/{uuid}/)
//	  master.m3u8          ← this file is created here
//	  h264_720/
//	    index.m3u8         ← referenced by master
//	  h264_1080/
//	    index.m3u8         ← referenced by master
//	  ...
func GenerateMasterPlaylist(hlsBaseDir string, variants []VariantInfo) error {
	if len(variants) == 0 {
		return fmt.Errorf("no variants provided for master playlist generation")
	}

	var entries []string
	for _, v := range variants {
		bandwidth := v.Bandwidth
		if bandwidth <= 0 {
			bandwidth = 1000000 // default 1 Mbps
		}
		entry := fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d", bandwidth)
		if v.Resolution != "" {
			entry += fmt.Sprintf(",RESOLUTION=%s", v.Resolution)
		}
		// Path is relative to the master playlist location (hlsBaseDir)
		// e.g., "h264_720/index.m3u8"
		entry += "\n" + filepath.ToSlash(v.Path)
		entries = append(entries, entry)
	}

	lines := []string{
		"#EXTM3U",
		"#EXT-X-VERSION=3",
	}
	lines = append(lines, entries...)
	masterContent := strings.Join(lines, "\n") + "\n"

	masterPath := filepath.Join(hlsBaseDir, "master.m3u8")
	if err := os.WriteFile(masterPath, []byte(masterContent), 0644); err != nil {
		return fmt.Errorf("failed to write master playlist at %s: %w", masterPath, err)
	}

	return nil
}
