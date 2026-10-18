package avatarimage

import (
	"bytes"
	"fmt"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"qmanga/internal/identity/application"
)

type Processor struct{}

func New() *Processor { return &Processor{} }
func (*Processor) ProcessAvatar(source io.Reader) ([]byte, error) {
	raw, err := io.ReadAll(io.LimitReader(source, application.MaxAvatarBytes+1))
	if err != nil || int64(len(raw)) > application.MaxAvatarBytes {
		return nil, fmt.Errorf("avatar size")
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || (format != "jpeg" && format != "png" && format != "webp") || config.Width < 1 || config.Height < 1 || config.Width > 8192 || config.Height > 8192 || int64(config.Width)*int64(config.Height) > 8_000_000 {
		return nil, fmt.Errorf("avatar dimensions or format")
	}
	src, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	side := min(config.Width, config.Height)
	left := (config.Width - side) / 2
	top := (config.Height - side) / 2
	target := image.NewRGBA(image.Rect(0, 0, 256, 256))
	draw.CatmullRom.Scale(target, target.Bounds(), src, image.Rect(left, top, left+side, top+side), draw.Src, nil)
	var output bytes.Buffer
	err = jpeg.Encode(&output, target, &jpeg.Options{Quality: 85})
	return output.Bytes(), err
}
