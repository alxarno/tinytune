package internal

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/alxarno/tinytune/pkg/index"
	"github.com/alxarno/tinytune/pkg/preview"
	"github.com/davidbyttow/govips/v2/vips"
)

const (
	maxWidthHeight    = 512
	vipsMaxCacheMem   = 16 * 1024 * 1024
	vipsMaxCacheSize  = 16 * 1024 * 1024
	vipsMaxCacheFiles = 128
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

func ImagePreview(path string) (preview.Data, error) {
	preview := preview.Data{Resolution: "0x0", ContentType: index.ContentTypeImage}

	image, err := vips.NewImageFromFile(path)
	if err != nil {
		return preview, fmt.Errorf("%w: %w", ErrVipsNewImage, err)
	}
	defer image.Close()

	scale := 1.0
	preview.Resolution = fmt.Sprintf("%dx%d", image.Width(), image.Height())

	if image.Width() > maxWidthHeight || image.Height() > maxWidthHeight {
		scale = float64(maxWidthHeight) / float64(max(image.Width(), image.Height()))
	}

	if err = image.Resize(scale, vips.KernelLanczos2); err != nil {
		return preview, fmt.Errorf("%w:%w", ErrImageResize, err)
	}

	ep := vips.NewWebpExportParams()

	bytes, _, err := image.ExportWebp(ep)
	if err != nil {
		return preview, fmt.Errorf("%w:%w", ErrImageExport, err)
	}

	preview.Data = bytes

	return preview, nil
}
