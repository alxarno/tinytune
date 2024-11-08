package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
)

type probeFormat struct {
	Duration string `json:"duration"`
}

type probeStream struct {
	Frames string `json:"nb_frames"`
}

type probeData struct {
	Format  probeFormat   `json:"format"`
	Streams []probeStream `json:"streams"`
}

type videoParams struct {
	device  string
	timeout time.Duration
}

func probeOutputFrames(a string) (float64, time.Duration, error) {
	pd := probeData{}
	err := json.Unmarshal([]byte(a), &pd)
	if err != nil {
		return 0, 0, err
	}
	// if no frames count in metadata, then just use some default for 1 min video, 24fps
	if len(pd.Streams) == 0 || pd.Streams[0].Frames == "" {
		return 3000, 0, nil
	}
	f, err := strconv.ParseFloat(pd.Streams[0].Frames, 64)
	if err != nil {
		return 0, 0, err
	}
	seconds, err := strconv.ParseFloat(pd.Format.Duration, 64)
	if err != nil {
		return 0, 0, err
	}
	return f, time.Duration(seconds) * time.Second, nil
}

func VideoPreview(path string, params videoParams) ([]byte, time.Duration, error) {
	metaJson, err := videoProbe(path, params.timeout)
	if err != nil {
		return nil, 0, err
	}
	frames, duration, err := probeOutputFrames(metaJson)
	if err != nil {
		return nil, 0, err
	}
	previewFrames := []int64{int64(frames * 0.2), int64(frames * 0.4), int64(frames * 0.6), int64(frames * 0.8)}
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
	commandArgs = append(commandArgs, []string{"-i", path, "-frames", "1", "-vf", fmt.Sprintf(`select=%s,scale=w='min(512\, iw*3/2):h=-1',tile=1x5`, previewSelectString)}...)
	commandArgs = append(commandArgs, []string{"-f", "image2", "pipe:1"}...)

	cmd := exec.CommandContext(ctx, "ffmpeg", commandArgs...)
	buf := bytes.NewBuffer(nil)
	stdErrBuf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	cmd.Stderr = stdErrBuf
	err = cmd.Run()
	if err != nil {
		return nil, 0, fmt.Errorf("[%s] %w", stdErrBuf.String(), err)
	}
	return buf.Bytes(), duration, nil
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
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("[%s] %w", stdErrBuf.String(), err)
	}
	return buf.String(), nil
}

func ProcessorProbe() error {
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
	cmd := exec.CommandContext(ctx, "bash", "-c", com+" -version | sed -n \"s/"+com+" version \\([-0-9.]*\\).*/\\1/p;\"")
	buf := bytes.NewBuffer(nil)
	stdErrBuf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	cmd.Stderr = stdErrBuf
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("[%s] %w", stdErrBuf.String(), err)
	}
	if buf.Len() == 0 {
		return fmt.Errorf("'%s': command not found", com)
	}
	clearVersion := strings.TrimSuffix(strings.TrimSuffix(buf.String(), "\n"), "-0")
	required, _ := version.NewVersion("4.4.2")
	existed, err := version.NewVersion(clearVersion)
	if err != nil {
		return fmt.Errorf("existed version(%s) of %s can't be parsed", buf.String(), com)
	}
	if existed.LessThan(required) {
		return fmt.Errorf("existed version(%s) of %s is less then recommend(%s)", existed.String(), com, required.String())
	}
	return nil
}

func PullVideoParams() (videoParams, error) {
	result := videoParams{}
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "ffmpeg", "-hwaccels")
	buf := bytes.NewBuffer(nil)
	stdErrBuf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	cmd.Stderr = stdErrBuf
	err := cmd.Run()
	if err != nil {
		return result, fmt.Errorf("[%s] %w", stdErrBuf.String(), err)
	}
	accelerators := strings.Split(strings.Split(buf.String(), "Hardware acceleration methods:")[1], "\n")
	if slices.Contains(accelerators, "cuda") {
		result.device = "cuda"
	} else if slices.Contains(accelerators, "dxva2") {
		result.device = "dxva2"
	} else if slices.Contains(accelerators, "vaapi") {
		result.device = "vaapi"
	} else if slices.Contains(accelerators, "vdpau") {
		result.device = "vdpau"
	} else if slices.Contains(accelerators, "d3d11va") {
		result.device = "d3d11va"
	}
	return result, nil
}
