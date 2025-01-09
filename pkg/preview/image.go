package preview

import (
	"errors"
	"fmt"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
)

const (
	maxWidthHeight   = 256
	jpegShrinkFactor = 8

	imageDefault = iota
	imageCollage
)

var (
	ErrVipsLoadImage            = errors.New("failed load image")
	ErrVipsResizeImage          = errors.New("failed resize image")
	ErrImageDownScale           = errors.New("failed to downscale the image")
	ErrImageColorSpaceTransform = errors.New("failed to transform the image's color space")
	ErrImageExport              = errors.New("failed export the image")
)

func imagePreview(path string) (data, error) {
	preview := data{}

	image, err := vips.NewThumbnailFromFile(path, maxWidthHeight, maxWidthHeight, vips.InterestingAll)
	if err != nil {
		return preview, fmt.Errorf("%w: %w", ErrVipsLoadImage, err)
	}
	defer image.Close()

	preview.width, preview.height = image.Width(), image.Height()
	preview.data, err = exportWebP(image)

	return preview, err
}

func exportWebP(image *vips.ImageRef) ([]byte, error) {
	ep := vips.NewWebpExportParams()

	bytes, _, err := image.ExportWebp(ep)
	if err != nil {
		// vips return stack traces for some corrupted files, so let's just hide stack trace
		if strings.Contains(err.Error(), "Stack") {
			errorMsg := strings.Split(err.Error(), "\n")
			//nolint:err113
			return nil, fmt.Errorf("%w: %w", ErrImageExport, errors.New(errorMsg[0]))
		}

		return nil, fmt.Errorf("%w: %w", ErrImageExport, err)
	}

	return bytes, nil
}

func downScale(image *vips.ImageRef, imageType int) error {
	scale := 1.0

	switch imageType {
	case imageDefault:
		if image.Width() > maxWidthHeight || image.Height() > maxWidthHeight {
			scale = float64(maxWidthHeight) / float64(max(image.Width(), image.Height()))
		}
	case imageCollage:
		if image.Width() > maxWidthHeight {
			scale = float64(maxWidthHeight) / float64(image.Width())
		}
	}

	if err := image.Resize(scale, vips.KernelAuto); err != nil {
		return fmt.Errorf("%w: %w", ErrVipsResizeImage, err)
	}

	return nil
}
