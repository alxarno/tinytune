package preview

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
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

//nolint:gochecknoinits
func init() {
	os.Setenv("MALLOC_ARENA_MAX", "2")
	vips.LoggingSettings(func(domain string, level vips.LogLevel, msg string) {
		domainSlog := slog.String("source", domain)

		switch level {
		case vips.LogLevelCritical:
		case vips.LogLevelError:
			slog.Error(msg, domainSlog)
		case vips.LogLevelDebug:
			slog.Debug(msg, domainSlog)
		case vips.LogLevelWarning:
			slog.Warn(msg, domainSlog)
		case vips.LogLevelInfo:
		case vips.LogLevelMessage:
		default:
			slog.Info(msg, domainSlog)
		}
	}, vips.LogLevelError)
	vips.Startup(&vips.Config{
		MaxCacheFiles: 1,
	})
}

func imagePreview(path string, size int64) (data, error) {
	preview := data{}
	params := &vips.ImportParams{}

	// for big jpegs (>500kb), use shrink load
	if size > 1<<19 {
		params.JpegShrinkFactor.Set(jpegShrinkFactor)
	}

	image, err := vips.LoadImageFromFile(path, params)
	if err != nil {
		return preview, fmt.Errorf("%w: %w", ErrVipsLoadImage, err)
	}
	defer image.Close()

	preview.width, preview.height = image.Width(), image.Height()

	if err := downScale(image, imageDefault); err != nil {
		return preview, fmt.Errorf("%w: %w", ErrImageDownScale, err)
	}

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
