package preview

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"slices"
	"time"
)

var (
	ErrPullSnapshot = errors.New("failed to pull snapshot")
)

const (
	VideoPreviewCollageItems = 5
)

type VideoProcessingAccelType string

const (
	Auto     VideoProcessingAccelType = "auto"
	Software VideoProcessingAccelType = "software"
	Hardware VideoProcessingAccelType = "hardware"
)

type VideoParams struct {
	timeout              time.Duration
	accel                ffmpegHWAccelType
	accelSupportedCodecs []string
}

func getScreenshots(ctx context.Context, path string, timestamps []time.Duration, params VideoParams) ([]byte, error) {
	screenshotsOptions := []string{"-hide_banner", "-loglevel", "error"}

	for i, v := range timestamps {
		screenshotsOptions = append(screenshotsOptions, options(params.accel)(path, v, i)...)
	}

	for range VideoPreviewCollageItems - len(timestamps) {
		lastOptionIndex := len(timestamps) - 1
		currentOptions := options(params.accel)(path, timestamps[lastOptionIndex], lastOptionIndex)
		screenshotsOptions = append(screenshotsOptions, currentOptions...)
	}

	pipeReader, pipeWriter := io.Pipe()
	stdOutBuf := bytes.NewBuffer(nil)

	producerCmd := exec.CommandContext(ctx, "ffmpeg", screenshotsOptions...)
	producerErrBuff := bytes.NewBuffer(nil)
	producerCmd.Stderr = producerErrBuff
	producerCmd.Stdout = pipeWriter

	mixerCmd := exec.CommandContext(ctx, "ffmpeg", tileOptions()...) //nolint:gosec
	mixerErrBuff := bytes.NewBuffer(nil)
	mixerCmd.Stdin = pipeReader
	mixerCmd.Stdout = stdOutBuf
	mixerCmd.Stderr = mixerErrBuff

	if err := producerCmd.Start(); err != nil {
		return nil, fmt.Errorf("producer start error %w", err)
	}

	if err := mixerCmd.Start(); err != nil {
		return nil, fmt.Errorf("mixer start error %w", err)
	}

	producerErr := producerCmd.Wait()

	pipeWriter.Close()

	mixerErr := mixerCmd.Wait()

	if producerErr != nil && !errors.Is(producerErr, context.Canceled) {
		return nil, fmt.Errorf("producer error [%s] %w", producerErrBuff.String(), producerErr)
	}

	if mixerErr != nil && !errors.Is(mixerErr, context.Canceled) {
		return nil, fmt.Errorf("mixer error [%s] %w", mixerErrBuff.String(), mixerErr)
	}

	return stdOutBuf.Bytes(), nil
}

func getTimeCodes(duration time.Duration) []time.Duration {
	parts := 5
	minimumStart := 2
	timestamps := []time.Duration{}
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
			continue
		}

		timestamps = append(timestamps, timestamp)
	}

	return timestamps
}

func produceVideoPreview(ctx context.Context, path string, duration time.Duration, params VideoParams) ([]byte, error) {
	if params.timeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, params.timeout)
		defer cancel()
	}

	screenshotTimeCodes := getTimeCodes(duration)

	data, err := getScreenshots(ctx, path, screenshotTimeCodes, params)
	if err != nil {
		return nil, fmt.Errorf("%w:%w", ErrPullSnapshot, err)
	}

	return data, nil
}

func videoPreview(ctx context.Context, path string, params VideoParams) (data, error) {
	preview := data{}

	metaJSON, err := videoProbe(ctx, path, params.timeout)
	if err != nil {
		return preview, err
	}

	output, err := probeOutputFrames(metaJSON)
	if err != nil {
		return preview, err
	}

	preview.width = output.width
	preview.height = output.height
	preview.duration = output.duration

	// switch unsupported codecs to software processing
	if len(params.accelSupportedCodecs) != 0 && !slices.Contains(params.accelSupportedCodecs, output.codec) {
		params.accel = ffmpegSoftwareAccel
	}

	if preview.data, err = produceVideoPreview(ctx, path, output.duration, params); err != nil {
		return preview, err
	}

	return preview, nil
}
