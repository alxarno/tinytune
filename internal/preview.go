package internal

import (
	"path/filepath"
	"slices"
	"time"

	"github.com/alxarno/tinytune/pkg/index"
	"github.com/alxarno/tinytune/pkg/preview"
)

type Previewer struct {
	image        bool
	video        bool
	acceleration bool
	videoParams  VideoParams
	videoFormats []string
	imageFormats []string
}

type PreviewerOption func(*Previewer)

func WithImagePreview(param bool) PreviewerOption {
	return func(p *Previewer) {
		p.image = param
	}
}

func WithVideoPreview(param bool) PreviewerOption {
	return func(p *Previewer) {
		p.video = param
	}
}

func WithAcceleration(param bool) PreviewerOption {
	return func(p *Previewer) {
		p.acceleration = param
	}
}

func NewPreviewer(opts ...PreviewerOption) (*Previewer, error) {
	timeoutMinutes := 10
	preview := &Previewer{
		//nolint:lll
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
		preview.videoParams.timeout = time.Minute * time.Duration(timeoutMinutes)
	}

	if !preview.acceleration {
		preview.videoParams.device = ""
	}

	return preview, nil
}

func (p Previewer) Pull(path string) (preview.Data, error) {
	contentType := p.ContentType(path)
	defaultPreview := preview.Data{ContentType: index.ContentTypeOther}

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

func (p Previewer) ContentType(path string) int {
	ext := filepath.Ext(path)
	if slices.Contains(p.imageFormats, ext[1:]) {
		return index.ContentTypeImage
	}

	if slices.Contains(p.videoFormats, ext[1:]) {
		return index.ContentTypeVideo
	}

	return index.ContentTypeOther
}
