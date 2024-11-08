package internal

import (
	"path/filepath"
	"slices"
	"time"

	"github.com/alxarno/tinytune/pkg/index"
)

type previewer struct {
	imagePreview bool
	videoPreview bool
	videoParams  videoParams
	videoFormats []string
	imageFormats []string
}

type PreviewerOption func(*previewer)

func WithImagePreview() PreviewerOption {
	return func(p *previewer) {
		p.imagePreview = true
	}
}

func WithVideoPreview() PreviewerOption {
	return func(p *previewer) {
		p.videoPreview = true
	}
}

func NewPreviewer(opts ...PreviewerOption) (*previewer, error) {
	preview := &previewer{
		videoFormats: []string{"3gp", "avi", "f4v", "flv", "gif", "hevc", "m4v", "mlv", "mov", "mp4", "m4a", "3g2", "mj2", "mpeg", "ogv", "webm"},
		imageFormats: []string{"jpeg", "png", "jpg", "webp", "bmp"},
	}
	for _, opt := range opts {
		opt(preview)
	}
	if preview.videoPreview {
		if err := ProcessorProbe(); err != nil {
			return nil, err
		}

		params, err := PullVideoParams()
		if err != nil {
			return nil, err
		}
		preview.videoParams = params
		preview.videoParams.timeout = time.Minute * 10
	}
	return preview, nil
}

func (p previewer) Pull(path string) (time.Duration, int, []byte, error) {
	ext := filepath.Ext(path)
	if len(ext) < 2 {
		return 0, index.ContentTypeOther, nil, nil
	}
	if slices.Contains(p.imageFormats, ext[1:]) && p.imagePreview {
		data, err := ImagePreview(path)
		return 0, index.ContentTypeImage, data, err
	}
	if slices.Contains(p.videoFormats, ext[1:]) && p.videoPreview {
		data, err := VideoPreview(path, p.videoParams)
		return 0, index.ContentTypeVideo, data, err
	}
	return 0, index.ContentTypeOther, nil, nil
}
