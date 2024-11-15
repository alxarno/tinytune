package preview

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/alxarno/tinytune/pkg/timeutil"
	"github.com/davidbyttow/govips/v2/vips"
	"golang.org/x/image/bmp"
)

var (
	ErrPullSnapshot = errors.New("failed to pull snapshot")
	ErrImageDecode  = errors.New("failed to decode image")
	ErrImageEncode  = errors.New("failed to encode image")
	ErrImageScale   = errors.New("failed to scale image")
)

type flvPullSnapshot = func(string, time.Duration, io.Writer)

func getFlvSnapshoter(wg *sync.WaitGroup, errCh chan error, params VideoParams) flvPullSnapshot {
	return func(path string, timestamp time.Duration, w io.Writer) {
		defer wg.Done()

		ctx := context.Background()

		if params.timeout > 0 {
			var cancel func()
			ctx, cancel = context.WithTimeout(context.Background(), params.timeout)
			defer cancel()
		}

		ffmpegTimestamp := timeutil.String(timestamp)

		seekOptions := []string{"-accurate_seek", "-ss", ffmpegTimestamp}
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

	err = bmp.Encode(&buff, previewImage)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrImageEncode, err)
	}

	image, err := vips.NewImageFromReader(&buff)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrImageScale, err)
	}
	defer image.Close()

	return downScale(image)
}

func FLVPreviewImage(path string, duration time.Duration, params VideoParams) ([]byte, error) {
	images := []*bytes.Buffer{}
	parts := 5
	step := duration / time.Duration(parts)
	waitGroup := new(sync.WaitGroup)
	errs := make(chan error, parts)

	for v := range parts {
		waitGroup.Add(1)

		timestamp := step * time.Duration(v)
		if timestamp == 0 {
			timestamp = time.Second
		}

		buff := new(bytes.Buffer)
		images = append(images, buff)

		go getFlvSnapshoter(waitGroup, errs, params)(path, timestamp, buff)
	}

	waitGroup.Wait()
	close(errs)

	for err := range errs {
		return nil, fmt.Errorf("%w: %w", ErrPullSnapshot, err)
	}

	return combineImagesToPreview(images)
}
