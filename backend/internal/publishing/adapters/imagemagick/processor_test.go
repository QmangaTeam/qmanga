package imagemagick

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os/exec"
	"testing"
	"time"

	"qmanga/internal/publishing/application"
	"qmanga/internal/publishing/domain"
)

func TestGeneratedFormatMustMatchDeclaredMIME(testingTarget *testing.T) {
	var source bytes.Buffer
	if failure := png.Encode(&source, image.NewRGBA(image.Rect(0, 0, 64, 32))); failure != nil {
		testingTarget.Fatal(failure)
	}
	fragment := domain.Fragment{Width: 64, Height: 32}
	if failure := validateGenerated(bytes.NewReader(source.Bytes()), "image/webp", fragment); failure == nil {
		testingTarget.Fatal("PNG accepted as WebP")
	}
	if failure := validateGenerated(bytes.NewReader(source.Bytes()), "image/png", fragment); failure != nil {
		testingTarget.Fatal(failure)
	}
}

func verifyGenerated(testingTarget *testing.T, batch application.Batch) {
	testingTarget.Helper()
	seen := map[string]bool{}
	for _, item := range batch.Items {
		content, failure := io.ReadAll(item.Content)
		if failure != nil {
			testingTarget.Fatal(failure)
		}
		decoded, format, failure := image.Decode(bytes.NewReader(content))
		if failure != nil {
			testingTarget.Fatal(failure)
		}
		mime := "image/webp"
		if item.Tier == "original" {
			mime = "image/png"
		}
		if item.Mime != mime || "image/"+format != mime || int64(len(content)) != item.Bytes || fmt.Sprintf("%x", sha256.Sum256(content)) != item.Checksum || decoded.Bounds().Dx() != item.Width || decoded.Bounds().Dy() != item.Height {
			testingTarget.Fatalf("incorrect generated metadata for %s", item.Tier)
		}
		limits := map[string]int{"eco": 720, "standard": 1080, "high": 1600}
		if limit, exists := limits[item.Tier]; exists && (item.Width > limit || item.Height > 2048) {
			testingTarget.Fatalf("unbounded %s variant", item.Tier)
		}
		seen[item.Tier] = true
	}
	for _, tier := range []string{"eco", "standard", "high", "original"} {
		if !seen[tier] {
			testingTarget.Fatalf("missing %s", tier)
		}
	}
}

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
	verifyGenerated(testingTarget, batch)
	broken := content[:64]
	page.Bytes = int64(len(broken))
	page.Checksum = fmt.Sprintf("%x", sha256.Sum256(broken))
	if bad, failure := New().Process(ctx, bytes.NewReader(broken), page); failure == nil {
		bad.Close()
		testingTarget.Fatal("truncated image accepted")
	}
}

func TestSeparateCoverIsSanitizedAndBounded(testingTarget *testing.T) {
	if _, failure := exec.LookPath("magick"); failure != nil {
		testingTarget.Skip("ImageMagick available in tools image")
	}
	var source bytes.Buffer
	if failure := png.Encode(&source, image.NewRGBA(image.Rect(0, 0, 900, 1500))); failure != nil {
		testingTarget.Fatal(failure)
	}
	cover, failure := New().ProcessCover(context.Background(), bytes.NewReader(source.Bytes()))
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	defer cover.Close()
	content, failure := io.ReadAll(cover.Content)
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	configuration, format, failure := image.DecodeConfig(bytes.NewReader(content))
	if failure != nil || format != "webp" || cover.Mime != "image/webp" || configuration.Width > 720 || configuration.Height > 1200 || configuration.Width != cover.Width || configuration.Height != cover.Height || cover.Bytes != int64(len(content)) || cover.Checksum != fmt.Sprintf("%x", sha256.Sum256(content)) {
		testingTarget.Fatalf("invalid independent cover: %+v %v", cover, failure)
	}
	if invalid, failure := New().ProcessCover(context.Background(), bytes.NewReader([]byte("not an image"))); failure == nil {
		invalid.Close()
		testingTarget.Fatal("invalid cover accepted")
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
	defer batch.Close()
	var webpSource []byte
	for _, item := range batch.Items {
		if item.Tier == "high" {
			webpSource, failure = io.ReadAll(item.Content)
			if failure != nil {
				testingTarget.Fatal(failure)
			}
			if _, failure = item.Content.Seek(0, io.SeekStart); failure != nil {
				testingTarget.Fatal(failure)
			}
			page = application.Page{ID: "webp", Width: item.Width, Height: item.Height, Bytes: int64(len(webpSource)), Checksum: item.Checksum, Format: "webp"}
			break
		}
	}
	verifyGenerated(testingTarget, batch)
	webpBatch, failure := New().Process(ctx, bytes.NewReader(webpSource), page)
	if failure != nil {
		testingTarget.Fatal(failure)
	}
	defer webpBatch.Close()
	verifyGenerated(testingTarget, webpBatch)
}
