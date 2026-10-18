package avatarimage

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"qmanga/internal/identity/application"
	"testing"
)

func TestAvatarSanitizesAndCenterCrops(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 400, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 400; x++ {
			src.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	var input bytes.Buffer
	if err := png.Encode(&input, src); err != nil {
		t.Fatal(err)
	}
	raw, err := New().ProcessAvatar(&input)
	if err != nil {
		t.Fatal(err)
	}
	result, err := jpeg.Decode(bytes.NewReader(raw))
	if err != nil || result.Bounds().Dx() != 256 || result.Bounds().Dy() != 256 {
		t.Fatalf("bad output %v %v", result, err)
	}
}
func TestAvatarRejectsUntrustedSources(t *testing.T) {
	for _, raw := range [][]byte{[]byte("<svg>fake</svg>"), make([]byte, application.MaxAvatarBytes+1), []byte("GIF89a")} {
		if _, err := New().ProcessAvatar(bytes.NewReader(raw)); err == nil {
			t.Fatal("untrusted avatar accepted")
		}
	}
}
