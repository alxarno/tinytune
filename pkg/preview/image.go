package preview

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/davidbyttow/govips/v2/vips"
)

const (
	maxWidthHeight    = 256
	vipsMaxCacheMem   = 16 * 1024 * 1024
	vipsMaxCacheSize  = 16 * 1024 * 1024
	vipsMaxCacheFiles = 128

	imageDefault = iota
	imageCollage
)

var (
	ErrVipsNewImage = errors.New("failed init new image")
	ErrImageResize  = errors.New("failed to resize the image")
	ErrImageExport  = errors.New("failed export the image")
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

func imagePreview(path string) (data, error) {
	preview := data{}

	image, err := vips.NewImageFromFile(path)
	if err != nil {
		return preview, fmt.Errorf("%w: %w", ErrVipsNewImage, err)
	}
	defer image.Close()

	preview.width, preview.height = image.Width(), image.Height()

	preview.data, err = downScale(image, imageDefault)

	return preview, err
}

func downScale(image *vips.ImageRef, imageType int) ([]byte, error) {
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

	if err := image.Resize(scale, vips.KernelLanczos2); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrImageResize, err)
	}

	ep := vips.NewWebpExportParams()

	bytes, _, err := image.ExportWebp(ep)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrImageExport, err)
	}

	return bytes, nil
}
