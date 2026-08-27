package service

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/conf"
	"origadmin/application/origstudio/internal/features/media/biz"
)

type StorageProxyService struct {
	storage biz.Storage
	paths   *conf.StoragePaths
	logger  *log.Helper
}

func NewStorageProxyService(storage biz.Storage, paths *conf.StoragePaths, logger log.Logger) *StorageProxyService {
	return &StorageProxyService{
		storage: storage,
		paths:   paths,
		logger:  log.NewHelper(log.With(logger, "module", "storage.proxy")),
	}
}

func (s *StorageProxyService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/files/")
	if key == "" {
		http.Error(w, "missing file key", http.StatusBadRequest)
		return
	}

	key = path.Clean(key)
	if strings.Contains(key, "..") {
		http.Error(w, "invalid key", http.StatusBadRequest)
		return
	}

	reader, err := s.storage.Download(r.Context(), key)
	if err != nil {
		s.logger.Warnf("storage download failed: key=%s err=%v", key, err)
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer reader.Close()

	contentType := detectContentType(key)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	var fileSize int64 = -1
	var seeker io.ReadSeeker
	if s, ok := reader.(io.ReadSeeker); ok {
		seeker = s
		if fi, err := seeker.Seek(0, io.SeekEnd); err == nil {
			fileSize = fi
			_, _ = seeker.Seek(0, io.SeekStart)
		}
	}

	if fileSize > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))

		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" && seeker != nil {
			rangeStart, rangeEnd, ok := parseRange(rangeHeader, fileSize)
			if ok {
				if _, err := seeker.Seek(rangeStart, io.SeekStart); err != nil {
					http.Error(w, "range not satisfiable", http.StatusRequestedRangeNotSatisfiable)
					return
				}
				w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", rangeStart, rangeEnd, fileSize))
				w.Header().Set("Content-Length", strconv.FormatInt(rangeEnd-rangeStart+1, 10))
				w.WriteHeader(http.StatusPartialContent)
				if r.Method == http.MethodHead {
					return
				}
				_, _ = io.CopyN(w, seeker, rangeEnd-rangeStart+1)
				return
			}
		}
	}

	if r.Method == http.MethodHead {
		return
	}
	_, _ = io.Copy(w, reader)
}

func parseRange(rangeHeader string, fileSize int64) (int64, int64, bool) {
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return 0, 0, false
	}
	rangeSpec := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.SplitN(rangeSpec, "-", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}

	startStr := strings.TrimSpace(parts[0])
	endStr := strings.TrimSpace(parts[1])

	if startStr == "" {
		suffixLength, err := strconv.ParseInt(endStr, 10, 64)
		if err != nil || suffixLength <= 0 || suffixLength > fileSize {
			return 0, 0, false
		}
		return fileSize - suffixLength, fileSize - 1, true
	}

	start, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil || start < 0 || start >= fileSize {
		return 0, 0, false
	}

	var end int64
	if endStr == "" {
		end = fileSize - 1
	} else {
		end, err = strconv.ParseInt(endStr, 10, 64)
		if err != nil || end < start || end >= fileSize {
			return 0, 0, false
		}
	}

	return start, end, true
}

func detectContentType(key string) string {
	ext := strings.ToLower(path.Ext(key))
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".m3u8":
		return "application/vnd.apple.mpegurl"
	case ".ts":
		return "video/mp2t"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".vtt":
		return "text/vtt; charset=utf-8"
	case ".webm":
		return "video/webm"
	case ".mp3":
		return "audio/mpeg"
	case ".aac":
		return "audio/aac"
	case ".ogg":
		return "audio/ogg"
	case ".wav":
		return "audio/wav"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".json":
		return "application/json"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	default:
		return "application/octet-stream"
	}
}
