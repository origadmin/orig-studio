package ffmpeg

import "strings"

// ResolutionToSize converts short-form resolution strings like "720" to
// WxH format required by ffmpeg's -s parameter, e.g., "1280x720".
// For values already in WxH format or non-standard values, returns as-is.
func ResolutionToSize(resolution string) string {
	sizes := map[string]string{
		"240":  "426x240",
		"360":  "640x360",
		"480":  "854x480",
		"720":  "1280x720",
		"1080": "1920x1080",
		"1440": "2560x1440",
		"2160": "3840x2160",
	}
	if s, ok := sizes[resolution]; ok {
		return s
	}
	// Already in WxH format
	if strings.Contains(resolution, "x") {
		return resolution
	}
	return resolution
}

// IsSkipResolution returns true for sentinel values that indicate
// the profile should not perform standard video transcoding.
func IsSkipResolution(resolution string) bool {
	return resolution == "" || resolution == "-"
}
