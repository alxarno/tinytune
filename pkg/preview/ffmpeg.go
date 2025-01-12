package preview

import (
	"fmt"
	"slices"
	"time"

	"github.com/alxarno/tinytune/pkg/timeutil"
)

type ffmpegHWAccelType string

const (
	ffmpegSoftwareAccel ffmpegHWAccelType = "software"
	ffmpegCudaAccel     ffmpegHWAccelType = "cuda"
)

func cudaOptions(path string, start time.Duration, index int) []string {
	mapValue := fmt.Sprintf("%d:v:0", index)
	cudaOptions := []string{"-hwaccel", "cuda", "-hwaccel_output_format", "cuda"}
	inputOptions := []string{"-ss", timeutil.String(start), "-i", path}
	filterOptions := []string{"-vf", "scale_cuda=256:-1", "-frames:v", "1"}
	encoderOptions := []string{"-c:v", "h264_nvenc"}
	outputOptions := []string{"-an", "-f", "rawvideo", "-map", mapValue, "pipe:1"}

	return slices.Concat(cudaOptions, inputOptions, filterOptions, encoderOptions, outputOptions)
}

func swOptions(path string, start time.Duration, index int) []string {
	mapValue := fmt.Sprintf("%d:v:0", index)
	inputOptions := []string{"-ss", timeutil.String(start), "-i", path}
	filterOptions := []string{"-vf", "scale=256:-2", "-frames:v", "1"}
	encoderOptions := []string{"-c:v", "libx264", "-preset", "ultrafast", "-tune", "zerolatency", "-crf", "0"}
	outputOptions := []string{"-an", "-f", "rawvideo", "-map", mapValue, "pipe:1"}

	return slices.Concat(inputOptions, filterOptions, encoderOptions, outputOptions)
}

func options(accel ffmpegHWAccelType) func(path string, start time.Duration, index int) []string {
	switch accel {
	case ffmpegCudaAccel:
		return cudaOptions
	case ffmpegSoftwareAccel:
		return swOptions
	default:
		return swOptions
	}
}

func tileOptions() []string {
	quietOptions := []string{"-hide_banner", "-loglevel", "error"}
	inputOptions := []string{"-i", "pipe:0"}
	filterOptions := []string{"-c:v", "libwebp", "-vf", "tile=1x5", "-frames:v", "1"}
	outputOptions := []string{"-f", "image2", "-an", "pipe:1"}

	return slices.Concat(quietOptions, inputOptions, filterOptions, outputOptions)
}
