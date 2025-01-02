package preview

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/alxarno/tinytune/pkg/timeutil"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/hashicorp/go-version"
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
	ErrImagesJoin   = errors.New("failed to join images")
	ErrImageEncode  = errors.New("failed to encode image")
	ErrImageScale   = errors.New("failed to scale image")
	ErrBufferCopy   = errors.New("failed to copy buffer")
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

func getSnapshot(path string, timestamp time.Duration, timeout time.Duration, w io.Writer) error {
	ctx := context.Background()

	if timeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	}

	ffmpegTimestamp := timeutil.String(timestamp)

	seekOptions := []string{"-loglevel", "quiet", "-accurate_seek", "-ss", ffmpegTimestamp}
	inputOptions := []string{"-i", path}
	outputOptions := []string{"-vf", "scale=256:-1", "-frames:v", "1", "-c:v", "mjpeg", "-f", "image2", "pipe:1"}

	options := seekOptions
	options = append(options, inputOptions...)
	options = append(options, outputOptions...)
	cmd := exec.CommandContext(ctx, "ffmpeg", options...)
	stdErrBuf := bytes.NewBuffer(nil)
	cmd.Stdout = w
	cmd.Stderr = stdErrBuf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("[%s] %w", stdErrBuf.String(), err)
	}

	return nil
}

func combineImagesToPreview(buffers []*bytes.Buffer) ([]byte, error) {
	preview, err := vips.NewImageFromBuffer(buffers[0].Bytes())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrImageDecode, err)
	}
	defer preview.Close()

	images := make([]*vips.ImageRef, 0, len(buffers)-1)

	defer func() {
		for _, v := range images {
			v.Close()
		}
	}()

	for _, v := range buffers[1:] {
		image, err := vips.NewImageFromBuffer(v.Bytes())
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrImageDecode, err)
		}

		images = append(images, image)
	}

	if err := preview.ArrayJoin(images, 1); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrImagesJoin, err)
	}

	if err := downScale(preview, imageCollage); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrImageScale, err)
	}

	return exportWebP(preview)
}

func getTimeCodes(duration time.Duration) ([]time.Duration, []bool) {
	parts := 5
	minimumStart := 2
	timestamps := []time.Duration{}
	repeat := make([]bool, parts)
	step := duration / time.Duration(parts)

	if step < time.Second {
		step = time.Second
	}

	for part := range parts {
		timestamp := step * time.Duration(part)
		if timestamp == 0 && duration > time.Second*7 {
			timestamp = time.Second * time.Duration(minimumStart)
		}

		// if video is small (< parts*time.Second), then just copy last images to remaining buffers
		if timestamp > duration && part > 1 || timestamp == duration {
			repeat[part] = true

			continue
		}

		timestamps = append(timestamps, timestamp)
	}

	return timestamps, repeat
}

func produceVideoPreview(path string, duration time.Duration, params VideoParams) ([]byte, error) {
	timeCodes, repeats := getTimeCodes(duration)
	images := make([]*bytes.Buffer, len(repeats))

	for index, timestamp := range timeCodes {
		images[index] = &bytes.Buffer{}

		if err := getSnapshot(path, timestamp, params.timeout, images[index]); err != nil {
			return nil, err
		}
	}

	for index, repeat := range repeats {
		if !repeat {
			continue
		}

		images[index] = images[index-1]
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

	logOptions := []string{"-hide_banner", "-loglevel", "quiet"}
	jobOptions := []string{"-show_format", "-show_streams", "-of", "json", path}
	options := []string{}
	options = append(options, logOptions...)
	options = append(options, jobOptions...)

	cmd := exec.CommandContext(ctx, "ffprobe", options...)
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
