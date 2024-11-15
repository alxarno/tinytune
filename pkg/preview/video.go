package preview

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

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
	device  string
	timeout time.Duration
}

func getVideoStream(streams []probeStream) *probeStream {
	for _, v := range streams {
		if v.CodecType == "video" {
			return &v
		}
	}

	return nil
}

func probeOutputFrames(a string) (string, float64, time.Duration, error) {
	data := probeData{}
	resolution := "0x0"
	frames := 3000.0

	if err := json.Unmarshal([]byte(a), &data); err != nil {
		return resolution, 0, 0, fmt.Errorf("%w: %w", ErrMetaInfoUnmarshal, err)
	}

	seconds, err := strconv.ParseFloat(data.Format.Duration, 64)
	if err != nil {
		return resolution, 0, 0, fmt.Errorf("%w: %w", ErrMetaInfoDurationParse, err)
	}

	videoStream := getVideoStream(data.Streams)
	if videoStream == nil {
		return resolution, 0, 0, ErrVideoStreamNotFound
	}

	resolution = fmt.Sprintf("%dx%d", videoStream.Width, videoStream.Height)

	// if no frames count in metadata, then just use some default for 1 min video, 24fps
	if videoStream.Frames != "" {
		frames, err = strconv.ParseFloat(data.Streams[0].Frames, 64)
		if err != nil {
			return resolution, 0, 0, fmt.Errorf("%w: %w", ErrMetaInfoFramesCountParse, err)
		}

		return resolution, frames, time.Duration(seconds) * time.Second, nil
	}

	if !strings.Contains(videoStream.AvgFrameRate, "/") {
		return resolution, frames, 0, nil
	}

	first, err := strconv.Atoi(strings.Split(videoStream.AvgFrameRate, "/")[0])
	if err != nil {
		return resolution, frames, 0, ErrParseFrameRate
	}

	second, err := strconv.Atoi(strings.Split(videoStream.AvgFrameRate, "/")[1])
	if err != nil {
		return resolution, frames, 0, ErrParseFrameRate
	}

	frames = seconds * float64(first/second)

	return resolution, frames, time.Duration(seconds) * time.Second, nil
}

func videoPreview(path string, params VideoParams) (data, error) {
	preview := data{resolution: "0x0"}

	metaJSON, err := videoProbe(path, params.timeout)
	if err != nil {
		return preview, err
	}

	resolution, frames, duration, err := probeOutputFrames(metaJSON)
	if err != nil {
		return preview, err
	}

	preview.resolution = resolution
	preview.duration = duration

	if path[len(path)-3:] == "flv" {
		if preview.data, err = FLVPreviewImage(path, duration, params); err != nil {
			return preview, err
		}

		return preview, nil
	}

	//nolint:gomnd,mnd
	previewFrames := []int64{
		int64(frames * 0.2),
		int64(frames * 0.4),
		int64(frames * 0.6),
		int64(frames * 0.8),
	}

	previewSelectString := "eq(n\\,0)"

	for _, v := range previewFrames {
		previewSelectString += fmt.Sprintf("+eq(n\\,%d)", v)
	}

	ctx := context.Background()

	if params.timeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(context.Background(), params.timeout)
		defer cancel()
	}

	commandArgs := []string{"-y", "-vsync", "0"}
	if params.device != "" {
		commandArgs = append(commandArgs, []string{"-hwaccel", params.device}...)
	}

	transform := fmt.Sprintf(`select=%s,scale=w='min(512\, iw*3/2):h=-1',tile=1x5`, previewSelectString)
	commandArgs = append(commandArgs, []string{"-i", path, "-frames", "1", "-vf", transform}...)
	commandArgs = append(commandArgs, []string{"-f", "image2", "pipe:1"}...)

	cmd := exec.CommandContext(ctx, "ffmpeg", commandArgs...)
	buf := bytes.NewBuffer(nil)
	stdErrBuf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	cmd.Stderr = stdErrBuf

	if err := cmd.Run(); err != nil {
		return preview, fmt.Errorf("[%s] %w", stdErrBuf.String(), err)
	}

	preview.data = buf.Bytes()

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

func pullVideoParams() (VideoParams, error) {
	result := VideoParams{}
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "ffmpeg", "-hwaccels")
	buf := bytes.NewBuffer(nil)
	stdErrBuf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	cmd.Stderr = stdErrBuf

	if err := cmd.Run(); err != nil {
		return result, fmt.Errorf("[%s] %w", stdErrBuf.String(), err)
	}

	accelerators := strings.Split(strings.Split(buf.String(), "Hardware acceleration methods:")[1], "\n")
	defaultAccelerators := []string{"cuda", "dxva2", "vaapi", "vdpau", "d3d11va"}

	for _, v := range defaultAccelerators {
		if slices.Contains(accelerators, v) {
			result.device = v

			break
		}
	}

	return result, nil
}
