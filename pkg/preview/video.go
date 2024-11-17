package preview

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alxarno/tinytune/pkg/timeutil"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/hashicorp/go-version"
	"golang.org/x/image/tiff"
)

var (
	ErrCommandNotFound          = errors.New("not found")
	ErrIncorrectFFmpegVersion   = errors.New("can't be parsed")
	ErrOutdatedFFmpegVersion    = errors.New("is outdated")
	ErrMetaInfoUnmarshal        = errors.New("failed to decode file's meta information")
	ErrMetaInfoFramesCountParse = errors.New("failed to parse frames count from meta information")
	ErrMetaInfoDurationParse    = errors.New("failed to parse duration from meta information")
	ErrVideoStreamNotFound      = errors.New("video stream not found")
	ErrParseFrameRate           = errors.New("failed to parse frame rate")

	ErrPullSnapshot = errors.New("failed to pull snapshot")
	ErrImageDecode  = errors.New("failed to decode image")
	ErrImageEncode  = errors.New("failed to encode image")
	ErrImageScale   = errors.New("failed to scale image")
)

type probeFormat struct {
	Duration string `json:"duration"`
}

type probeStream struct {
	Frames       string `json:"nb_frames"` //nolint:tagliatelle
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	AvgFrameRate string `json:"avg_frame_rate"` //nolint:tagliatelle
	CodecType    string `json:"codec_type"`     //nolint:tagliatelle
}

type probeData struct {
	Format  probeFormat   `json:"format"`
	Streams []probeStream `json:"streams"`
}

type VideoParams struct {
	timeout time.Duration
}

type videoPullSnapshot = func(string, time.Duration, io.Writer)

func getVideoSnapshoter(wg *sync.WaitGroup, errCh chan error, params VideoParams) videoPullSnapshot {
	return func(path string, timestamp time.Duration, w io.Writer) {
		defer wg.Done()

		ctx := context.Background()

		if params.timeout > 0 {
			var cancel func()
			ctx, cancel = context.WithTimeout(context.Background(), params.timeout)
			defer cancel()
		}

		ffmpegTimestamp := timeutil.String(timestamp)

		seekOptions := []string{"-loglevel", "quiet", "-accurate_seek", "-ss", ffmpegTimestamp}
		inputOptions := []string{"-i", path}
		outputOptions := []string{"-frames:v", "1", "-c:v", "bmp", "-f", "image2", "pipe:1"}

		options := seekOptions
		options = append(options, inputOptions...)
		options = append(options, outputOptions...)

		cmd := exec.CommandContext(ctx, "ffmpeg", options...)
		stdErrBuf := bytes.NewBuffer(nil)
		cmd.Stdout = w
		cmd.Stderr = stdErrBuf

		if err := cmd.Run(); err != nil {
			errCh <- fmt.Errorf("[%s] %w", stdErrBuf.String(), err)
		}
	}
}

func combineImagesToPreview(buffers []*bytes.Buffer) ([]byte, error) {
	firstImage, _, err := image.Decode(buffers[0])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrImageDecode, err)
	}

	height := firstImage.Bounds().Dy()
	previewHeight := height * len(buffers)
	previewImage := image.NewRGBA(image.Rectangle{
		Min: firstImage.Bounds().Min,
		Max: image.Point{
			X: firstImage.Bounds().Max.X,
			Y: previewHeight,
		},
	})
	draw.Draw(previewImage, previewImage.Bounds(), firstImage, image.Point{0, 0}, draw.Src)

	for i, v := range buffers[1:] {
		img, _, err := image.Decode(v)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrImageDecode, err)
		}

		s := image.Point{0, (i + 1) * height}
		r := image.Rectangle{s, s.Add(img.Bounds().Size())}

		draw.Draw(previewImage, r, img, image.Point{0, 0}, draw.Src)
	}

	buff := bytes.Buffer{}
	err = tiff.Encode(&buff, previewImage, &tiff.Options{
		Compression: tiff.Uncompressed,
		Predictor:   false,
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrImageEncode, err)
	}

	image, err := vips.NewImageFromReader(&buff)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrImageScale, err)
	}
	defer image.Close()

	return downScale(image, imageCollage)
}

func produceVideoPreview(path string, duration time.Duration, params VideoParams) ([]byte, error) {
	minimumStart := 3
	images := []*bytes.Buffer{}
	parts := 5
	step := duration / time.Duration(parts)
	waitGroup := new(sync.WaitGroup)
	errs := make(chan error, parts)

	for v := range parts {
		waitGroup.Add(1)

		timestamp := step * time.Duration(v)
		if timestamp == 0 {
			timestamp = time.Second * time.Duration(minimumStart)
		}

		buff := new(bytes.Buffer)
		images = append(images, buff)

		go getVideoSnapshoter(waitGroup, errs, params)(path, timestamp, buff)
	}

	waitGroup.Wait()
	close(errs)

	for err := range errs {
		return nil, fmt.Errorf("%w: %w", ErrPullSnapshot, err)
	}

	return combineImagesToPreview(images)
}

func getVideoStream(streams []probeStream) *probeStream {
	for _, v := range streams {
		if v.CodecType == "video" {
			return &v
		}
	}

	return nil
}

func probeOutputFrames(a string) (int, int, time.Duration, error) {
	data := probeData{}

	if err := json.Unmarshal([]byte(a), &data); err != nil {
		return 0, 0, 0, fmt.Errorf("%w: %w", ErrMetaInfoUnmarshal, err)
	}

	seconds, err := strconv.ParseFloat(data.Format.Duration, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("%w: %w", ErrMetaInfoDurationParse, err)
	}

	videoStream := getVideoStream(data.Streams)
	if videoStream == nil {
		return 0, 0, 0, ErrVideoStreamNotFound
	}

	return videoStream.Width, videoStream.Height, time.Duration(seconds) * time.Second, nil
}

func videoPreview(path string, params VideoParams) (data, error) {
	preview := data{}

	metaJSON, err := videoProbe(path, params.timeout)
	if err != nil {
		return preview, err
	}

	width, height, duration, err := probeOutputFrames(metaJSON)
	if err != nil {
		return preview, err
	}

	preview.width = width
	preview.height = height
	preview.duration = duration

	if preview.data, err = produceVideoPreview(path, duration, params); err != nil {
		return preview, err
	}

	return preview, nil
}

func videoProbe(path string, timeOut time.Duration) (string, error) {
	ctx := context.Background()

	if timeOut > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(context.Background(), timeOut)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, "ffprobe", "-show_format", "-show_streams", "-of", "json", path)
	buf := bytes.NewBuffer(nil)
	stdErrBuf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	cmd.Stderr = stdErrBuf

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("[%s] %w", stdErrBuf.String(), err)
	}

	return buf.String(), nil
}

func processorProbe() error {
	if err := probeFFmpeg("ffmpeg"); err != nil {
		return err
	}

	if err := probeFFmpeg("ffprobe"); err != nil {
		return err
	}

	return nil
}

func probeFFmpeg(com string) error {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "bash", "-c", com+" -version | sed -n \"s/"+com+" version \\([-0-9.]*\\).*/\\1/p;\"") //nolint:gosec,lll
	buf := bytes.NewBuffer(nil)
	stdErrBuf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	cmd.Stderr = stdErrBuf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("[%s] %w", stdErrBuf.String(), err)
	}

	if buf.Len() == 0 {
		return fmt.Errorf("'%s' %w", com, ErrCommandNotFound)
	}

	clearVersion := strings.TrimSuffix(strings.TrimSuffix(buf.String(), "\n"), "-0")
	required, _ := version.NewVersion("4.4.2")

	existed, err := version.NewVersion(clearVersion)
	if err != nil {
		return fmt.Errorf("%w version(%s) of %s: %w", ErrIncorrectFFmpegVersion, buf.String(), com, err)
	}

	if existed.LessThan(required) {
		return fmt.Errorf("%w version(%s) of %s: %w", ErrOutdatedFFmpegVersion, existed.String(), com, err)
	}

	return nil
}
