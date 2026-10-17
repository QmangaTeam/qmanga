package imagemagick

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os/exec"
	"testing"
	"time"

	"qmanga/internal/publishing/application"
)

func TestRealImageProcessingAndTruncatedRejection(testingTarget *testing.T) {
	if _, failure := exec.LookPath("magick"); failure != nil {
		testingTarget.Skip("ImageMagick available in tools image")
	}
	var source bytes.Buffer
	if failure := png.Encode(&source, image.NewRGBA(image.Rect(0, 0, 64, 5000))); failure != nil {
		testingTarget.Fatal(failure)
	}
	content := source.Bytes()
	checksum := fmt.Sprintf("%x", sha256.Sum256(content))
	page := application.Page{ID: "page", Width: 64, Height: 5000, Bytes: int64(len(content)), Checksum: checksum, Format: "png"}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	batch, failure := New().Process(ctx, bytes.NewReader(content), page)
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	defer batch.Close()
	if len(batch.Items) != 10 {
		testingTarget.Fatalf("unexpected fragments: %d", len(batch.Items))
	}
	for _, item := range batch.Items {
		if item.Width > 64 || item.Tier != "original" && item.Height > 2048 {
			testingTarget.Fatal("upscaled or unbounded image")
		}
	}
	broken := content[:64]
	page.Bytes = int64(len(broken))
	page.Checksum = fmt.Sprintf("%x", sha256.Sum256(broken))
	if bad, failure := New().Process(ctx, bytes.NewReader(broken), page); failure == nil {
		bad.Close()
		testingTarget.Fatal("truncated image accepted")
	}
}

func TestRealJPEGProcessing(testingTarget *testing.T) {
	if _, failure := exec.LookPath("magick"); failure != nil {
		testingTarget.Skip("ImageMagick available in tools image")
	}
	var source bytes.Buffer
	if failure := jpeg.Encode(&source, image.NewRGBA(image.Rect(0, 0, 64, 5000)), nil); failure != nil {
		testingTarget.Fatal(failure)
	}
	content := source.Bytes()
	page := application.Page{ID: "jpeg", Width: 64, Height: 5000, Bytes: int64(len(content)), Checksum: fmt.Sprintf("%x", sha256.Sum256(content)), Format: "jpeg"}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	batch, failure := New().Process(ctx, bytes.NewReader(content), page)
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	batch.Close()
}
