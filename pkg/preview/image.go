package preview

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
)

const (
	maxWidthHeight    = 256
	vipsMaxCacheMem   = 16 * 1024 * 1024
	vipsMaxCacheSize  = 16 * 1024 * 1024
	vipsMaxCacheFiles = 4
	jpegShrinkFactor  = 8

	imageDefault = iota
	imageCollage
)

var (
	ErrVipsLoadImage   = errors.New("failed load image")
	ErrVipsResizeImage = errors.New("failed resize image")
	ErrImageDownScale  = errors.New("failed to downscale the image")
	ErrImageExport     = errors.New("failed export the image")
)

//nolint:gochecknoinits
func init() {
	vips.LoggingSettings(func(domain string, level vips.LogLevel, msg string) {
		slog.Info(msg, slog.String("domain", domain), slog.Int("level", int(level)), slog.String("source", "VIPS"))
	}, vips.LogLevelError)
	vips.Startup(&vips.Config{
		ConcurrencyLevel: runtime.NumCPU(),
		MaxCacheMem:      vipsMaxCacheMem,
		MaxCacheSize:     vipsMaxCacheSize,
		MaxCacheFiles:    vipsMaxCacheFiles,
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

	if err := image.Resize(scale, vips.KernelNearest); err != nil {
		return fmt.Errorf("%w: %w", ErrVipsResizeImage, err)
	}

	return nil
}
