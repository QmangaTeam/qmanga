package bootstrap

import (
	"context"
	"errors"
	"io"
	"qmanga/internal/catalog/application"
	"qmanga/internal/catalog/domain"
	"qmanga/internal/publishing/adapters/imagemagick"
	publishing "qmanga/internal/publishing/domain"
)

type titleArtwork struct{ processor *imagemagick.Processor }

func (a titleArtwork) ProcessChangeImage(ctx context.Context, reader io.Reader, purpose string) (application.ProcessedChangeImage, error) {
	image, err := a.processor.ProcessArtwork(ctx, reader, purpose)
	if errors.Is(err, publishing.ErrInvalidImage) {
		return application.ProcessedChangeImage{}, domain.ErrInvalid
	}
	if err != nil {
		return application.ProcessedChangeImage{}, err
	}
	return application.ProcessedChangeImage{Asset: application.TitleChangeAsset{Mime: image.Mime, Checksum: image.Checksum, Width: image.Width, Height: image.Height, Bytes: image.Bytes}, Content: image.Content, Close: image.Close}, nil
}
