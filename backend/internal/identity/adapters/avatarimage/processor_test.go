package avatarimage

import (
	"bytes"
	"fmt"
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

// Generated JPEG fixtures contain four broad color quadrants and a minimal
// standards-shaped EXIF IFD0. Samples stay away from lossy JPEG boundaries.
func orientedJPEG(t *testing.T, orientation byte, little bool) []byte {
	t.Helper()
	src := image.NewRGBA(image.Rect(0, 0, 400, 200))
	colors := []color.RGBA{{240, 20, 20, 255}, {20, 240, 20, 255}, {20, 20, 240, 255}, {240, 240, 20, 255}}
	for y := 0; y < 200; y++ {
		for x := 0; x < 400; x++ {
			src.SetRGBA(x, y, colors[(y/100)*2+x/200])
		}
	}
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, src, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatal(err)
	}
	tiff := []byte{'M', 'M', 0, 42, 0, 0, 0, 8, 0, 1, 1, 18, 0, 3, 0, 0, 0, 1, 0, orientation, 0, 0, 0, 0, 0, 0}
	if little {
		tiff = []byte{'I', 'I', 42, 0, 8, 0, 0, 0, 1, 0, 18, 1, 3, 0, 1, 0, 0, 0, orientation, 0, 0, 0, 0, 0, 0, 0}
	}
	payload := append([]byte("Exif\x00\x00"), tiff...)
	segment := append([]byte{255, 225, byte((len(payload) + 2) >> 8), byte(len(payload) + 2)}, payload...)
	return append(append(append([]byte{}, encoded.Bytes()[:2]...), segment...), encoded.Bytes()[2:]...)
}
func TestAvatarHonorsEXIFOrientationBeforeCrop(t *testing.T) {
	colors := []color.RGBA{{240, 20, 20, 255}, {20, 240, 20, 255}, {20, 20, 240, 255}, {240, 240, 20, 255}}
	cases := [][]int{{0, 1, 2, 3}, {1, 0, 3, 2}, {3, 2, 1, 0}, {2, 3, 0, 1}, {0, 2, 1, 3}, {2, 0, 3, 1}, {3, 1, 2, 0}, {1, 3, 0, 2}}
	for index, want := range cases {
		for _, little := range []bool{false, true} {
			t.Run(fmt.Sprintf("orientation%d/little%t", index+1, little), func(t *testing.T) {
				raw, err := New().ProcessAvatar(bytes.NewReader(orientedJPEG(t, byte(index+1), little)))
				if err != nil {
					t.Fatal(err)
				}
				result, err := jpeg.Decode(bytes.NewReader(raw))
				if err != nil {
					t.Fatal(err)
				}
				if result.Bounds() != image.Rect(0, 0, 256, 256) {
					t.Fatalf("bounds %v", result.Bounds())
				}
				if bytes.Contains(raw, []byte("Exif")) {
					t.Fatal("EXIF leaked to public avatar")
				}
				for corner, point := range []image.Point{{32, 32}, {224, 32}, {32, 224}, {224, 224}} {
					got := color.RGBAModel.Convert(result.At(point.X, point.Y)).(color.RGBA)
					expected := colors[want[corner]]
					for _, pair := range [][2]uint8{{got.R, expected.R}, {got.G, expected.G}, {got.B, expected.B}} {
						if int(pair[0])-int(pair[1]) > 35 || int(pair[1])-int(pair[0]) > 35 {
							t.Fatalf("corner %d color %v want %v", corner, got, expected)
						}
					}
				}
			})
		}
	}
}

func TestAvatarMalformedEXIFKeepsSafeDefault(t *testing.T) {
	for _, orientation := range []byte{0, 9, 255} {
		raw := orientedJPEG(t, orientation, false)
		if got := jpegOrientation(raw); got != 1 {
			t.Fatalf("invalid orientation %d accepted: %d", orientation, got)
		}
		if _, err := New().ProcessAvatar(bytes.NewReader(raw)); err != nil {
			t.Fatal(err)
		}
	}
	// Reject malformed TIFF magic, byte order, out-of-bounds IFD offsets,
	// oversized entry count, wrong field type and external/multi-value fields.
	source := orientedJPEG(t, 6, false)
	mutations := []func([]byte){
		func(raw []byte) { raw[12] = 'X' },
		func(raw []byte) { raw[15] = 43 },
		func(raw []byte) { copy(raw[16:20], []byte{255, 255, 255, 255}) },
		func(raw []byte) { raw[20] = 255; raw[21] = 255 },
		func(raw []byte) { raw[25] = 4 },
		func(raw []byte) { raw[29] = 2 },
		func(raw []byte) { raw[4] = 255; raw[5] = 255 },
	}
	for index, mutate := range mutations {
		raw := bytes.Clone(source)
		mutate(raw)
		if got := jpegOrientation(raw); got != 1 {
			t.Fatalf("malformed case%d orientation=%d", index, got)
		}
	}
	for size := 0; size < len(source); size++ {
		jpegOrientation(source[:size])
	}
}

func FuzzJPEGOrientationBounds(f *testing.F) {
	f.Add([]byte{255, 216, 255, 225, 0, 2, 255, 217})
	f.Add([]byte("Exif\x00\x00MM\x00*\xff\xff\xff\xff"))
	f.Fuzz(func(t *testing.T, raw []byte) {
		if orientation := jpegOrientation(raw); orientation < 1 || orientation > 8 {
			t.Fatalf("orientation %d", orientation)
		}
	})
}
