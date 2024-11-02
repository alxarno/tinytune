package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"time"
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

func probeOutputFrames(a string) (float64, error) {
	pd := probeData{}
	err := json.Unmarshal([]byte(a), &pd)
	if err != nil {
		return 0, err
	}
	// if no frames count in metadata, then just use some default for 1 min video, 24fps
	if len(pd.Streams) == 0 || pd.Streams[0].Frames == "" {
		return 3000, nil
	}
	f, err := strconv.ParseFloat(pd.Streams[0].Frames, 64)
	if err != nil {
		return 0, err
	}
	return f, nil
}

func VideoPreview(path string, timeOut time.Duration) ([]byte, error) {
	metaJson, err := videoProbe(path, timeOut)
	if err != nil {
		return nil, err
	}
	frames, err := probeOutputFrames(metaJson)
	if err != nil {
		return nil, err
	}
	previewFrames := []int64{int64(frames * 0.2), int64(frames * 0.4), int64(frames * 0.6), int64(frames * 0.8)}
	previewSelectString := "eq(n\\,0)"
	for _, v := range previewFrames {
		previewSelectString += fmt.Sprintf("+eq(n\\,%d)", v)
	}
	ctx := context.Background()
	if timeOut > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(context.Background(), timeOut)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-vsync", "0", "-hwaccel", "cuda", "-i", path, "-frames", "1", "-vf", fmt.Sprintf(`select=%s,scale=w='min(512\, iw*3/2):h=-1',tile=1x5`, previewSelectString), "-f", "webp", "pipe:1")
	buf := bytes.NewBuffer(nil)
	stdErrBuf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	cmd.Stderr = stdErrBuf
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("[%s] %w", stdErrBuf.String(), err)
	}
	return buf.Bytes(), nil
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
