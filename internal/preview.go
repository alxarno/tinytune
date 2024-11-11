package internal

import (
	"path/filepath"
	"slices"
	"time"

	"github.com/alxarno/tinytune/pkg/index"
	"github.com/alxarno/tinytune/pkg/preview"
)

type previewer struct {
	image        bool
	video        bool
	acceleration bool
	videoParams  videoParams
	videoFormats []string
	imageFormats []string
}

type PreviewerOption func(*previewer)

func WithImagePreview(param bool) PreviewerOption {
	return func(p *previewer) {
		p.image = param
	}
}

func WithVideoPreview(param bool) PreviewerOption {
	return func(p *previewer) {
		p.video = param
	}
}

func WithAcceleration(param bool) PreviewerOption {
	return func(p *previewer) {
		p.acceleration = param
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
	if preview.video {
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
	if !preview.acceleration {
		preview.videoParams.device = ""
	}
	return preview, nil
}

func (p previewer) Pull(path string) (preview.PreviewData, error) {
	contentType := p.ContentType(path)
	defaultPreview := preview.PreviewData{ContentType: index.ContentTypeOther}

	if contentType == index.ContentTypeImage && p.image {
		preview, err := ImagePreview(path)
		if err != nil {
			return defaultPreview, err
		}
		preview.ContentType = contentType
		return preview, nil
	}
	if contentType == index.ContentTypeVideo && p.video {
		preview, err := VideoPreview(path, p.videoParams)
		if err != nil {
			return defaultPreview, err
		}
		preview.ContentType = contentType
		return preview, nil
	}
	return defaultPreview, nil
}

func (p previewer) ContentType(path string) int {
	ext := filepath.Ext(path)
	if len(ext) < 2 {
		return index.ContentTypeOther
	}
	if slices.Contains(p.imageFormats, ext[1:]) {
		return index.ContentTypeImage
	}
	if slices.Contains(p.videoFormats, ext[1:]) {
		return index.ContentTypeVideo
	}
	return index.ContentTypeOther
}
