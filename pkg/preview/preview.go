package preview

import (
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
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
	maxImages     int64
	maxVideos     int64
	excludedFiles map[string]struct{}
	videoParams   VideoParams
	maxFileSize   int64
	timeout       time.Duration
}

//nolint:gochecknoinits
func init() {
	vips.LoggingSettings(func(domain string, level vips.LogLevel, msg string) {
		domainSlog := slog.String("source", domain)

		switch level {
		case vips.LogLevelCritical, vips.LogLevelError:
			slog.Error(msg, domainSlog)
		case vips.LogLevelDebug:
			slog.Debug(msg, domainSlog)
		case vips.LogLevelWarning:
			slog.Warn(msg, domainSlog)
		case vips.LogLevelMessage, vips.LogLevelInfo:
			slog.Info(msg, domainSlog)
		default:
			slog.Info(msg, domainSlog)
		}
	}, vips.LogLevelError)
	vips.Startup(&vips.Config{
		MaxCacheFiles: 1,
	})
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

func (p Previewer) Close() {
	vips.Shutdown()
}
