package internal

import (
	"fmt"
	"runtime"

	"github.com/davidbyttow/govips/v2/vips"
)

const MAX_WIDTH_HEIGHT = 512

func init() {
	vips.LoggingSettings(func(domain string, level vips.LogLevel, msg string) {
		fmt.Println(domain, level, msg)
	}, vips.LogLevelError)
	vips.Startup(&vips.Config{
		ConcurrencyLevel: runtime.NumCPU(),
		MaxCacheMem:      16 * 1024 * 1024,
		MaxCacheSize:     16 * 1024 * 1024,
		MaxCacheFiles:    128,
	})
}

func ImagePreview(path string) ([]byte, error) {
	image, err := vips.NewImageFromFile(path)
	if err != nil {
		panic(err)
	}
	defer image.Close()
	scale := 1.0
	if image.Width() > MAX_WIDTH_HEIGHT || image.Height() > MAX_WIDTH_HEIGHT {
		scale = float64(MAX_WIDTH_HEIGHT) / float64(max(image.Width(), image.Height()))
	}
	if err = image.Resize(scale, vips.KernelLanczos2); err != nil {
		return nil, err
	}
	ep := vips.NewWebpExportParams()
	bytes, _, err := image.ExportWebp(ep)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
