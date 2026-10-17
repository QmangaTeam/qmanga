package imagemagick

import (
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	_ "golang.org/x/image/webp"
	"qmanga/internal/publishing/application"
	"qmanga/internal/publishing/domain"
)

type Processor struct{}

func New() *Processor { return &Processor{} }
func (processor *Processor) Process(ctx context.Context, reader io.Reader, page application.Page) (application.Batch, error) {
	directory, failure := os.MkdirTemp("", "qmanga-image-")
	if failure != nil {
		return application.Batch{}, failure
	}
	batch := application.Batch{}
	files := []*os.File{}
	success := false
	batch.Close = func() {
		for _, file := range files {
			file.Close()
		}
		os.RemoveAll(directory)
	}
	defer func() {
		if !success {
			batch.Close()
		}
	}()
	sourcePath := filepath.Join(directory, "source")
	source, failure := os.Create(sourcePath)
	if failure != nil {
		return batch, failure
	}
	hash := sha256.New()
	size, failure := io.Copy(io.MultiWriter(source, hash), io.LimitReader(reader, (25<<20)+1))
	if failure != nil {
		source.Close()
		return batch, failure
	}
	if size != page.Bytes || size > 25<<20 || fmt.Sprintf("%x", hash.Sum(nil)) != page.Checksum {
		source.Close()
		return batch, domain.ErrInvalidImage
	}
	if _, failure = source.Seek(0, io.SeekStart); failure != nil {
		source.Close()
		return batch, failure
	}
	configuration, format, failure := image.DecodeConfig(source)
	source.Close()
	if failure != nil || configuration.Width != page.Width || configuration.Height != page.Height || format != page.Format {
		return batch, domain.ErrInvalidImage
	}
	if _, failure = domain.PlanFragments(page.Width, page.Height, 720); failure != nil {
		return batch, failure
	}
	appendFile := func(path, tier, mime string, fragment domain.Fragment) error {
		file, failure := os.Open(path)
		if failure != nil {
			return failure
		}
		files = append(files, file)
		configuration, _, failure := image.DecodeConfig(file)
		if failure != nil || configuration.Width != fragment.Width || configuration.Height != fragment.Height {
			return domain.ErrInvalidImage
		}
		if _, failure = file.Seek(0, io.SeekStart); failure != nil {
			return failure
		}
		hash := sha256.New()
		size, failure := io.Copy(hash, file)
		if failure != nil {
			return failure
		}
		if size <= 0 || size > 500<<20 {
			return domain.ErrInvalidImage
		}
		if _, failure = file.Seek(0, io.SeekStart); failure != nil {
			return failure
		}
		batch.Items = append(batch.Items, application.Generated{Variant: application.Variant{Tier: tier, Mime: mime, Width: fragment.Width, Height: fragment.Height, Top: fragment.Top, Offset: fragment.Offset, Bytes: size, Checksum: fmt.Sprintf("%x", hash.Sum(nil))}, Content: file})
		return nil
	}
	for _, profile := range []struct {
		tier           string
		width, quality int
	}{{"eco", 720, 70}, {"standard", 1080, 82}, {"high", 1600, 92}} {
		fragments, failure := domain.PlanFragments(page.Width, page.Height, profile.width)
		if failure != nil {
			return batch, failure
		}
		width := fragments[0].Width
		last := fragments[len(fragments)-1]
		height := last.Top + last.Height
		pattern := filepath.Join(directory, profile.tier+"-%03d.webp")
		arguments := []string{"-regard-warnings", sourcePath + "[0]", "-strip", "-background", "white", "-alpha", "remove", "-alpha", "off", "-resize", fmt.Sprintf("%dx%d!", width, height), "-crop", fmt.Sprintf("%dx2048", width), "+repage", "-scene", "0", "-quality", strconv.Itoa(profile.quality), pattern}
		if failure = run(ctx, arguments); failure != nil {
			return batch, failure
		}
		for position, fragment := range fragments {
			if failure = appendFile(fmt.Sprintf(pattern, position), profile.tier, "image/webp", fragment); failure != nil {
				return batch, failure
			}
		}
	}
	originalPath := filepath.Join(directory, "original.png")
	if failure = run(ctx, []string{"-regard-warnings", sourcePath + "[0]", "-strip", "-background", "white", "-alpha", "remove", "-alpha", "off", "-depth", "8", "-define", "png:exclude-chunks=all", originalPath}); failure != nil {
		return batch, failure
	}
	if failure = appendFile(originalPath, "original", "image/png", domain.Fragment{Width: page.Width, Height: page.Height}); failure != nil {
		return batch, failure
	}
	success = true
	return batch, nil
}
func run(ctx context.Context, arguments []string) error {
	command := exec.CommandContext(ctx, "magick", arguments...)
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	return command.Run()
}
