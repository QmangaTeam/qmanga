package imagemagick

import (
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"time"

	"qmanga/internal/publishing/application"
	"qmanga/internal/publishing/domain"
)

// ProcessCover creates one metadata-free, bounded WebP independent of chapter pages.
func (processor *Processor) ProcessCover(ctx context.Context, reader io.Reader) (application.ProcessedCover, error) {
	return processor.processArtwork(ctx, reader, 720, 1200, 25<<20)
}

// ProcessArtwork preserves image proportions and strips metadata for each bounded use.
func (processor *Processor) ProcessArtwork(ctx context.Context, reader io.Reader, purpose string) (application.ProcessedCover, error) {
	switch purpose {
	case "cover":
		return processor.ProcessCover(ctx, reader)
	case "banner":
		return processor.processArtwork(ctx, reader, 1600, 800, 25<<20)
	case "evidence":
		return processor.processArtwork(ctx, reader, 2000, 3000, 5<<20)
	default:
		return application.ProcessedCover{}, domain.ErrInvalidImage
	}
}
func (processor *Processor) processArtwork(ctx context.Context, reader io.Reader, maxWidth, maxHeight int, maxBytes int64) (application.ProcessedCover, error) {
	directory, failure := os.MkdirTemp("", "qmanga-cover-")
	if failure != nil {
		return application.ProcessedCover{}, failure
	}
	defer func() {
		if failure != nil {
			_ = os.RemoveAll(directory)
		}
	}()
	sourcePath := filepath.Join(directory, "source")
	source, failure := os.Create(sourcePath)
	if failure != nil {
		return application.ProcessedCover{}, failure
	}
	size, copyFailure := io.Copy(source, io.LimitReader(reader, maxBytes+1))
	closeFailure := source.Close()
	if copyFailure != nil || closeFailure != nil {
		failure = copyFailure
		if failure == nil {
			failure = closeFailure
		}
		return application.ProcessedCover{}, failure
	}
	if size == 0 || size > maxBytes {
		failure = domain.ErrInvalidImage
		return application.ProcessedCover{}, failure
	}
	source, failure = os.Open(sourcePath)
	if failure != nil {
		return application.ProcessedCover{}, failure
	}
	configuration, format, decodeFailure := image.DecodeConfig(source)
	_ = source.Close()
	if decodeFailure != nil || (format != "jpeg" && format != "png" && format != "webp") || configuration.Width <= 0 || configuration.Height <= 0 || int64(configuration.Width)*int64(configuration.Height) > 80000000 {
		failure = domain.ErrInvalidImage
		return application.ProcessedCover{}, failure
	}
	path := filepath.Join(directory, "cover.webp")
	processing, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	if failure = run(processing, []string{sourcePath + "[0]", "-auto-orient", "-strip", "-resize", fmt.Sprintf("%dx%d>", maxWidth, maxHeight), "-quality", "82", path}); failure != nil {
		failure = domain.ErrInvalidImage
		return application.ProcessedCover{}, failure
	}
	content, failure := os.Open(path)
	if failure != nil {
		return application.ProcessedCover{}, failure
	}
	configuration, format, failure = image.DecodeConfig(content)
	if failure != nil || format != "webp" || configuration.Width <= 0 || configuration.Width > maxWidth || configuration.Height <= 0 || configuration.Height > maxHeight {
		_ = content.Close()
		failure = domain.ErrInvalidImage
		return application.ProcessedCover{}, failure
	}
	if _, failure = content.Seek(0, io.SeekStart); failure != nil {
		_ = content.Close()
		return application.ProcessedCover{}, failure
	}
	hash := sha256.New()
	bytes, failure := io.Copy(hash, content)
	if failure != nil || bytes == 0 || bytes > maxBytes {
		_ = content.Close()
		failure = domain.ErrInvalidImage
		return application.ProcessedCover{}, failure
	}
	if _, failure = content.Seek(0, io.SeekStart); failure != nil {
		_ = content.Close()
		return application.ProcessedCover{}, failure
	}
	return application.ProcessedCover{
		CoverAsset: application.CoverAsset{Mime: "image/webp", Checksum: fmt.Sprintf("%x", hash.Sum(nil)), Width: configuration.Width, Height: configuration.Height, Bytes: bytes},
		Content:    content,
		Close:      func() { _ = content.Close(); _ = os.RemoveAll(directory) },
	}, nil
}
