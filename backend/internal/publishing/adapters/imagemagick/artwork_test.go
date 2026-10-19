package imagemagick

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"io"
	"qmanga/internal/publishing/domain"
	"testing"
)

func TestArtworkBoundsAndPurpose(t *testing.T) {
	var source bytes.Buffer
	if err := png.Encode(&source, image.NewRGBA(image.Rect(0, 0, 2000, 1000))); err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		purpose       string
		width, height int
	}{{"cover", 720, 360}, {"banner", 1600, 800}, {"evidence", 2000, 1000}} {
		processed, err := New().ProcessArtwork(context.Background(), bytes.NewReader(source.Bytes()), check.purpose)
		if err != nil {
			t.Fatal(check.purpose, err)
		}
		content, err := io.ReadAll(processed.Content)
		processed.Close()
		if err != nil {
			t.Fatal(err)
		}
		image, format, err := image.DecodeConfig(bytes.NewReader(content))
		if err != nil || format != "webp" || image.Width != check.width || image.Height != check.height || int64(len(content)) != processed.Bytes {
			t.Fatal(check.purpose, image, format, err)
		}
	}
	if _, err := New().ProcessArtwork(context.Background(), bytes.NewReader(source.Bytes()), "unknown"); !errors.Is(err, domain.ErrInvalidImage) {
		t.Fatal("unknown image purpose accepted", err)
	}
	if _, err := New().ProcessArtwork(context.Background(), io.LimitReader(zeroReader{}, (5<<20)+1), "evidence"); !errors.Is(err, domain.ErrInvalidImage) {
		t.Fatal("oversized evidence accepted", err)
	}
}

type zeroReader struct{}

func (zeroReader) Read(value []byte) (int, error) { clear(value); return len(value), nil }
