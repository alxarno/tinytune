package internal

import (
	"path/filepath"
	"slices"
	"time"

	"github.com/alxarno/tinytune/pkg/index"
)

func getSupportedVideoFormats() []string {
	return []string{"3gp", "avi", "f4v", "flv", "gif", "hevc", "m4v", "mlv", "mov", "mp4", "m4a", "3g2", "mj2", "mpeg", "ogv", "webm"}
}

func getSupportedImageFormats() []string {
	return []string{"jpeg", "png", "jpg", "webp", "bmp"}
}

func GeneratePreview(path string) (time.Duration, int, []byte, error) {
	ext := filepath.Ext(path)
	if len(ext) < 2 {
		return 0, index.ContentTypeOther, nil, nil
	}
	if slices.Contains(getSupportedImageFormats(), ext[1:]) {
		data, err := ImagePreview(path)
		return 0, index.ContentTypeImage, data, err
	}
	if slices.Contains(getSupportedVideoFormats(), ext[1:]) {
		data, err := VideoPreview(path, time.Minute*10)
		return 0, index.ContentTypeVideo, data, err
	}
	return 0, index.ContentTypeOther, nil, nil
}
