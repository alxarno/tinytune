package preview

import (
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"
)

var (
	ErrVideoPreview = errors.New("failed create preview for video")
	ErrImagePreview = errors.New("failed create preview for image")
)

type Source interface {
	IsImage() bool
	IsVideo() bool
	IsAnimatedImage() bool
	Path() string
	Size() int64
}

type Previewer struct {
	image         bool
	video         bool
	acceleration  bool
	maxImages     int64
	maxVideos     int64
	excludedFiles map[string]struct{}
	videoParams   VideoParams
	maxFileSize   int64
	timeout       time.Duration
}

func NewPreviewer(opts ...Option) (*Previewer, error) {
	preview := &Previewer{
		maxImages:     -1,
		maxVideos:     -1,
		excludedFiles: map[string]struct{}{},
		image:         true,
		video:         true,
		maxFileSize:   -1,
	}

	for _, opt := range opts {
		opt(preview)
	}

	if preview.video {
		if err := processorProbe(); err != nil {
			return nil, err
		}

		preview.videoParams.timeout = preview.timeout
	}

	return preview, nil
}

func ifMaxPass(maxNewItems *int64) bool {
	if atomic.LoadInt64(maxNewItems) == -1 {
		return true
	}

	if atomic.LoadInt64(maxNewItems) == 0 {
		return false
	}

	atomic.AddInt64(maxNewItems, -1)

	return true
}

//nolint:cyclop,ireturn,nolintlint //it's very simple method...
func (p Previewer) Pull(src Source) (Data, error) {
	defaultPreview := data{}

	biggestThenMaxFileSize := p.maxFileSize != -1 && src.Size() > p.maxFileSize
	toImage := src.IsImage() && p.image && ifMaxPass(&p.maxImages)
	toVideo := src.IsVideo() && p.video && ifMaxPass(&p.maxVideos)

	if biggestThenMaxFileSize {
		return defaultPreview, nil
	}

	if toImage {
		preview, err := imagePreview(src.Path())
		if err != nil {
			return defaultPreview, fmt.Errorf("%w: %w", ErrImagePreview, err)
		}

		return preview, nil
	}

	if toVideo {
		timer := time.AfterFunc(time.Minute, func() {
			slog.Warn("File media processing run for more than a minute", slog.String("file", src.Path()))
		})
		defer timer.Stop()

		preview, err := videoPreview(src.Path(), p.videoParams)
		if err != nil || preview.Duration() == 0 {
			return defaultPreview, fmt.Errorf("%w: %w", ErrVideoPreview, err)
		}

		return preview, nil
	}

	if src.IsVideo() {
		// default resolution for video player
		defaultPreview.width = 1280
		defaultPreview.height = 720
	}

	return defaultPreview, nil
}
