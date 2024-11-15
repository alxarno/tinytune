package preview

import (
	"time"
)

type Source interface {
	IsImage() bool
	IsVideo() bool
	Path() string
}

type Previewer struct {
	image        bool
	video        bool
	acceleration bool
	videoParams  VideoParams
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
	preview := &Previewer{}

	for _, opt := range opts {
		opt(preview)
	}

	if preview.video {
		if err := processorProbe(); err != nil {
			return nil, err
		}

		params, err := pullVideoParams()
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

//nolint:ireturn
func (p Previewer) Pull(src Source) (Data, error) {
	defaultPreview := data{resolution: "0x0"}

	if src.IsImage() && p.image {
		preview, err := imagePreview(src.Path())
		if err != nil {
			return defaultPreview, err
		}

		return preview, nil
	}

	if src.IsVideo() && p.video {
		preview, err := videoPreview(src.Path(), p.videoParams)
		if err != nil || preview.Duration() == 0 {
			return defaultPreview, err
		}

		return preview, nil
	}

	if src.IsVideo() {
		// default resolution for video player
		defaultPreview.resolution = "1280x720"
	}

	return defaultPreview, nil
}
