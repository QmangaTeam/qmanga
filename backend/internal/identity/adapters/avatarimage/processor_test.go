package avatarimage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
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
	f.Add([]byte("\x89PNG\r\n\x1a\n\xff\xff\xff\xffeXIf"))

	f.Fuzz(func(t *testing.T, raw []byte) {
		for _, parser := range []func([]byte) uint16{pngOrientation} {
			if got := parser(raw); got < 1 || got > 8 {
				t.Fatalf("orientation %d", got)
			}
		}
		if orientation := jpegOrientation(raw); orientation < 1 || orientation > 8 {
			t.Fatalf("orientation %d", orientation)
		}
	})
}

func TestAvatarMatchesPNGAndWebPBrowserOrientation(t *testing.T) {
	// Reuse the generated JPEG's quadrants and TIFF metadata, converting only
	// the container. WebP fixture is our own 400x200 four-color artwork encoded
	// by ffmpeg libwebp -lossless 1. Chrome preserves its pixels despite EXIF6;
	// PNG/JPEG are oriented. Test samples encode this observed preview contract.
	for _, format := range []string{"png", "webp"} {
		t.Run(format, func(t *testing.T) {
			source := orientedJPEG(t, 6, false)
			tiff := source[12:38]
			var raw []byte
			if format == "png" {
				decoded, err := jpeg.Decode(bytes.NewReader(append([]byte{255, 216}, source[38:]...)))
				if err != nil {
					t.Fatal(err)
				}
				var encoded bytes.Buffer
				if err = png.Encode(&encoded, decoded); err != nil {
					t.Fatal(err)
				}
				payload := append([]byte("eXIf"), tiff...)
				chunk := make([]byte, 4)
				binary.BigEndian.PutUint32(chunk, uint32(len(tiff)))
				chunk = append(chunk, payload...)
				crc := make([]byte, 4)
				binary.BigEndian.PutUint32(crc, crc32.ChecksumIEEE(payload))
				chunk = append(chunk, crc...)
				raw = append(append(append([]byte{}, encoded.Bytes()[:33]...), chunk...), encoded.Bytes()[33:]...)
			} else {
				encoded, err := os.ReadFile("testdata/quadrants.webp")
				if err != nil {
					t.Fatal(err)
				}
				// Extended header: EXIF flag, 400x200 canvas, followed by original VP8L.
				raw = append([]byte("RIFF\x00\x00\x00\x00WEBPVP8X\x0a\x00\x00\x00\x08\x00\x00\x00\x8f\x01\x00\xc7\x00\x00"), encoded[12:]...)
				raw = append(raw, []byte("EXIF\x1a\x00\x00\x00")...)
				raw = append(raw, tiff...)
				binary.LittleEndian.PutUint32(raw[4:8], uint32(len(raw)-8))
			}
			if format == "png" {
				if got := pngOrientation(raw); got != 6 {
					t.Fatalf("container orientation=%d", got)
				}
				for size := 0; size < len(raw); size++ {
					pngOrientation(raw[:size])
				}
				broken := bytes.Clone(raw)
				binary.BigEndian.PutUint32(broken[8:12], 0xffffffff)
				if got := pngOrientation(broken); got != 1 {
					t.Fatalf("unbounded chunk orientation=%d", got)
				}
				broken = bytes.Clone(raw)
				broken[33+8] ^= 1
				if got := pngOrientation(broken); got != 1 {
					t.Fatalf("bad EXIF CRC orientation=%d", got)
				}
			}

			output, err := New().ProcessAvatar(bytes.NewReader(raw))
			if err != nil {
				t.Fatal(err)
			}
			result, err := jpeg.Decode(bytes.NewReader(output))
			if err != nil {
				t.Fatal(err)
			}
			for index, point := range []image.Point{{32, 32}, {224, 32}, {32, 224}, {224, 224}} {
				got := color.RGBAModel.Convert(result.At(point.X, point.Y)).(color.RGBA)
				want := []color.RGBA{{20, 20, 240, 255}, {240, 20, 20, 255}, {240, 240, 20, 255}, {20, 240, 20, 255}}[index]
				if format == "webp" {
					want = []color.RGBA{{240, 20, 20, 255}, {20, 240, 20, 255}, {20, 20, 240, 255}, {240, 240, 20, 255}}[index]
				}
				if absColor(got.R, want.R) > 35 || absColor(got.G, want.G) > 35 || absColor(got.B, want.B) > 35 {
					t.Fatalf("corner %d got%v want%v", index, got, want)
				}
			}
		})
	}
}
func absColor(a, b uint8) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}
