package preview

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
)

const (
	BigVideosQueueSize                 = 4
	BigVideoSizeB                      = 500 * 1024 * 1024
	BigVideoThrottleMaxOccupiedPercent = 0.9
	BigVideoThrottleMaxWaiting         = time.Duration(5) * time.Second
)

var (
	ErrVideoPreview               = errors.New("failed create preview for video")
	ErrImagePreview               = errors.New("failed create preview for image")
	ErrFFmpegCudaDecodersNotFound = errors.New("cuda decoders not found in ffmpeg")
	ErrFFmpegCudaProbe            = errors.New("failed probe cuda decoders in ffmpeg")
)

type Source interface {
	IsImage() bool
	IsVideo() bool
	IsAnimatedImage() bool
	Path() string
	Size() int64
}

type Previewer struct {
	image          bool
	video          bool
	maxImages      int64
	maxVideos      int64
	excludedFiles  map[string]struct{}
	videoParams    VideoParams
	videoAccelType VideoProcessingAccelType
	bigVideoQueue  chan struct{}
	maxFileSize    int64
	timeout        time.Duration
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
		maxImages:      -1,
		maxVideos:      -1,
		excludedFiles:  map[string]struct{}{},
		image:          true,
		video:          true,
		maxFileSize:    -1,
		videoAccelType: Software,
		bigVideoQueue:  make(chan struct{}, BigVideosQueueSize),
	}

	for _, opt := range opts {
		opt(preview)
	}

	if preview.video {
		videoParams, err := videoInit(preview.videoAccelType)
		if err != nil {
			return nil, err
		}

		videoParams.timeout = preview.timeout
		preview.videoParams = videoParams
	}

	if preview.videoParams.accel != ffmpegSoftwareAccel {
		slog.Info(
			"Video accelerator",
			slog.String("type", string(preview.videoParams.accel)),
			slog.String("codecs", strings.Join(preview.videoParams.accelSupportedCodecs, ",")),
		)
	}

	return preview, nil
}

func videoInit(videoAccelType VideoProcessingAccelType) (VideoParams, error) {
	params := VideoParams{accel: ffmpegSoftwareAccel}

	if err := processorProbe(); err != nil {
		return params, err
	}

	if videoAccelType == Software {
		return params, nil
	}

	codecs, err := probeCuda(context.Background())
	if err == nil {
		params.accel = ffmpegCudaAccel
		params.accelSupportedCodecs = codecs

		return params, nil
	}

	if errors.Is(err, ErrNoSupportedCodecs) && videoAccelType == Hardware {
		return params, ErrFFmpegCudaDecodersNotFound
	}

	if errors.Is(err, ErrNoSupportedCodecs) {
		return params, nil
	}

	return params, fmt.Errorf("%w: %w", ErrFFmpegCudaProbe, err)
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
func (p Previewer) Pull(ctx context.Context, src Source) (Data, error) {
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
		if src.Size() > BigVideoSizeB {
			p.bigVideoQueue <- struct{}{}

			err := throttle(ctx, BigVideoThrottleMaxOccupiedPercent, BigVideoThrottleMaxWaiting)
			if errors.Is(err, context.Canceled) {
				return defaultPreview, context.Canceled
			}

			defer func() { <-p.bigVideoQueue }()
		}

		timer := time.AfterFunc(time.Minute, func() {
			slog.Warn("File media processing run for more than a minute", slog.String("file", src.Path()))
		})
		defer timer.Stop()

		preview, err := videoPreview(ctx, src.Path(), p.videoParams)
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
