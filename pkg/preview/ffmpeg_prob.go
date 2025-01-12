package preview

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
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
	ErrNoSupportedCodecs        = errors.New("there are no supported codecs")
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
	CodecName    string `json:"codec_nae"`      //nolint:tagliatelle
}

type probeData struct {
	Format  probeFormat   `json:"format"`
	Streams []probeStream `json:"streams"`
}

func processorProbe() error {
	ctx := context.Background()

	if err := probeFFmpeg(ctx, "ffmpeg"); err != nil {
		return err
	}

	if err := probeFFmpeg(ctx, "ffprobe"); err != nil {
		return err
	}

	return nil
}

func formatCodecs(buff *bytes.Buffer) []string {
	codecs := []string{}
	scanner := bufio.NewScanner(buff)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		withoutPrefix := strings.TrimPrefix(scanner.Text(), "(codec ")
		codec := strings.TrimSuffix(withoutPrefix, ")")
		codecs = append(codecs, codec)
	}

	return codecs
}

func probeCuda(ctx context.Context) ([]string, error) {
	//nolint:lll
	cmd := exec.CommandContext(ctx, "bash", "-c", "ffprobe -hide_banner -decoders | grep Nvidia | grep -oP \"[(]codec\\s\\w+[)]\"")
	outBuff := bytes.NewBuffer(nil)
	errBuff := bytes.NewBuffer(nil)
	cmd.Stdout = outBuff
	cmd.Stderr = errBuff

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("[%s] %w", errBuff.String(), err)
	}

	if outBuff.Len() == 0 {
		return nil, fmt.Errorf("'%s' %w", "ffprobe -hide_banner -decoders", ErrCommandNotFound)
	}

	supportedCudaCodecs := formatCodecs(outBuff)
	if len(supportedCudaCodecs) == 0 {
		return nil, ErrNoSupportedCodecs
	}

	return supportedCudaCodecs, nil
}

func probeFFmpeg(ctx context.Context, com string) error {
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

func getVideoStream(streams []probeStream) *probeStream {
	for _, v := range streams {
		if v.CodecType == "video" {
			return &v
		}
	}

	return nil
}

type probeOutput struct {
	width    int
	height   int
	duration time.Duration
	codec    string
}

func probeOutputFrames(a string) (probeOutput, error) {
	data := probeData{}
	output := probeOutput{}

	if err := json.Unmarshal([]byte(a), &data); err != nil {
		return output, fmt.Errorf("%w: %w", ErrMetaInfoUnmarshal, err)
	}

	seconds, err := strconv.ParseFloat(data.Format.Duration, 64)
	if err != nil {
		return output, fmt.Errorf("%w: %w", ErrMetaInfoDurationParse, err)
	}

	videoStream := getVideoStream(data.Streams)
	if videoStream == nil {
		return output, ErrVideoStreamNotFound
	}

	output.width = videoStream.Width
	output.height = videoStream.Height
	output.duration = time.Duration(seconds) * time.Second
	output.codec = videoStream.CodecName

	return output, nil
}

func videoProbe(ctx context.Context, path string, timeOut time.Duration) (string, error) {
	if timeOut > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, timeOut)
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
