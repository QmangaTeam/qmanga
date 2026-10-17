package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func mediaRange(value string, size int64) (int64, int64, bool) {
	if !strings.HasPrefix(value, "bytes=") || strings.Contains(value, ",") || size < 1 {
		return 0, 0, false
	}
	parts := strings.Split(strings.TrimPrefix(value, "bytes="), "-")
	if len(parts) != 2 {
		return 0, 0, false
	}
	if parts[0] == "" {
		suffix, failure := strconv.ParseInt(parts[1], 10, 64)
		if failure != nil || suffix < 1 {
			return 0, 0, false
		}
		length := min(suffix, size)
		return size - length, length, true
	}
	start, failure := strconv.ParseInt(parts[0], 10, 64)
	if failure != nil || start < 0 || start >= size {
		return 0, 0, false
	}
	end := size - 1
	if parts[1] != "" {
		end, failure = strconv.ParseInt(parts[1], 10, 64)
		if failure != nil || end < start {
			return 0, 0, false
		}
		end = min(end, size-1)
	}
	return start, end - start + 1, true
}
func (handler *PublishingHandler) asset(request *gin.Context) {
	select {
	case handler.media <- struct{}{}:
		defer func() { <-handler.media }()
	default:
		request.Header("Retry-After", "1")
		Error(request, 429, "busy", "Сурет кезегі толы")
		return
	}
	asset, failure := handler.service.Asset(request.Request.Context(), handler.actor(request), request.Param("asset_id"))
	if failure != nil {
		publicationError(request, failure)
		return
	}
	etag := "\"" + asset.Checksum + "\""
	request.Header("ETag", etag)
	request.Header("Last-Modified", asset.ModifiedAt.UTC().Format(http.TimeFormat))
	request.Header("Accept-Ranges", "bytes")
	if asset.Public {
		request.Header("Cache-Control", "public, max-age=0, must-revalidate")
	} else {
		request.Header("Cache-Control", "private, no-store")
	}
	matched := false
	if condition := request.GetHeader("If-None-Match"); condition != "" {
		for _, candidate := range strings.Split(condition, ",") {
			candidate = strings.TrimSpace(candidate)
			if candidate == "*" || strings.TrimPrefix(candidate, "W/") == etag {
				matched = true
			}
		}
	} else if modified, failure := http.ParseTime(request.GetHeader("If-Modified-Since")); failure == nil {
		matched = !asset.ModifiedAt.Truncate(time.Second).After(modified)
	}
	if matched {
		request.Status(304)
		return
	}
	offset, length, status := int64(0), asset.Bytes, 200
	if value := request.GetHeader("Range"); value != "" && request.Request.Method == "GET" {
		apply := true
		if condition := request.GetHeader("If-Range"); condition != "" && condition != etag {
			modified, failure := http.ParseTime(condition)
			apply = failure == nil && !asset.ModifiedAt.Truncate(time.Second).After(modified)
		}
		if apply {
			var valid bool
			offset, length, valid = mediaRange(value, asset.Bytes)
			if !valid {
				request.Header("Content-Range", fmt.Sprintf("bytes */%d", asset.Bytes))
				request.Status(416)
				return
			}
			status = 206
		}
	}
	ctx, cancel := context.WithTimeout(request.Request.Context(), 5*time.Minute)
	defer cancel()
	stream, failure := handler.service.Stream(ctx, asset, offset, length)
	if failure != nil {
		request.Header("Cache-Control", "no-store")
		Error(request, 503, "media_unavailable", "Суретті жүктеу мүмкін болмады")
		return
	}
	defer stream.Close()
	request.Header("Content-Type", asset.Mime)
	request.Header("Content-Length", strconv.FormatInt(length, 10))
	if status == 206 {
		request.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", offset, offset+length-1, asset.Bytes))
	}
	request.Status(status)
	if request.Request.Method == "HEAD" {
		return
	}
	_, _ = io.CopyN(request.Writer, stream, length)
}
