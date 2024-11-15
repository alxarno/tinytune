package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"time"

	"github.com/alxarno/tinytune/pkg/index"
	"github.com/alxarno/tinytune/pkg/timeutil"
)

var (
	ErrParseChunkIndex = errors.New("failed parse chunk index")
	ErrFFMpeg          = errors.New("ffmpeg command failed")
	ErrCopy            = errors.New("failed to copy data")
)

const hlsChunkDurationSeconds = 10

func pullHLSIndex(meta *index.Meta, w io.Writer) error {
	fullSamples := 0
	lastSampleDuration := int(meta.Duration.Seconds())

	for lastSampleDuration >= 10 {
		lastSampleDuration -= hlsChunkDurationSeconds
		fullSamples++
	}

	data := bytes.Buffer{}

	data.WriteString("#EXTM3U\n")
	data.WriteString("#EXT-X-PLAYLIST-TYPE:VOD\n")
	data.WriteString("#EXT-X-TARGETDURATION:10\n")
	data.WriteString("#EXT-X-VERSION:4\n")
	data.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n")

	for c := range fullSamples {
		data.WriteString("#EXTINF:10.0,\n")
		data.WriteString(fmt.Sprintf("%v/%d.ts\n", meta.ID, c))
	}

	data.WriteString(fmt.Sprintf("#EXTINF:%d.0,\n", lastSampleDuration))
	data.WriteString(fmt.Sprintf("%v/%d.ts\n", meta.ID, fullSamples))
	data.WriteString("#EXT-X-ENDLIST")

	_, err := io.Copy(w, &data)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCopy, err)
	}

	return nil
}

func pullHLSChunk(ctx context.Context, meta *index.Meta, name string, timeout time.Duration, w io.Writer) error {
	index, err := strconv.Atoi(name[:len(name)-3])
	if err != nil {
		return fmt.Errorf("%w: %w", ErrParseChunkIndex, err)
	}

	start := time.Duration(index*hlsChunkDurationSeconds) * time.Second

	end := start + hlsChunkDurationSeconds*time.Second
	if end > meta.Duration {
		end = meta.Duration
	}

	if timeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	seekOptions := []string{"-accurate_seek", "-ss", timeutil.String(start), "-to", timeutil.String(end)}
	inputOptions := []string{"-i", string(meta.AbsolutePath)}
	outputOptions := []string{
		"-preset", "ultrafast",
		"-crf", "30",
		"-map_metadata", "0",
		"-movflags", "frag_keyframe+empty_moov+default_base_moof+faststart",
		"-copyts", "-copytb", "0",
		"-c:v", "libx264",
		"-c:a", "aac",
		"-b:a", "160k",
		"-bsf:v", "h264_mp4toannexb",
		"-f", "mpegts",
		"-crf", "32",
		"pipe:1",
	}

	options := seekOptions
	options = append(options, inputOptions...)
	options = append(options, outputOptions...)

	cmd := exec.CommandContext(ctx, "ffmpeg", options...)
	stdErrBuf := bytes.NewBuffer(nil)
	cmd.Stdout = w
	cmd.Stderr = stdErrBuf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w [%s]: %w", ErrFFMpeg, stdErrBuf.String(), err)
	}

	return nil
}
