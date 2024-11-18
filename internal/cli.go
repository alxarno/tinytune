package internal

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/alxarno/tinytune/pkg/bytesutil"
)

const defaultPort = 8080
const maxPortNumber = 65536

type RawConfig struct {
	Dir           string
	Parallel      int
	Video         bool
	Images        bool
	MaxImages     int64
	MaxVideos     int64
	Includes      string
	Excludes      string
	MaxFileSize   string
	Streaming     string
	MediaTimeout  string
	IndexFileSave bool
	Port          int
}

type MediaTypeConfig struct {
	Process  bool
	MaxItems int64
}

func (c MediaTypeConfig) Print(name string) {
	params := []any{
		slog.Bool("processing", c.Process),
	}

	if c.MaxItems != -1 {
		params = append(params, slog.Int64("max-items", c.MaxItems))
	}

	slog.Info(
		name,
		params...,
	)
}

type ProcessConfig struct {
	Parallel    int
	Video       MediaTypeConfig
	Timeout     time.Duration
	Image       MediaTypeConfig
	Includes    []*regexp.Regexp
	Excludes    []*regexp.Regexp
	MaxFileSize int64
}

func (c ProcessConfig) Print() {
	includes := ""
	excludes := ""

	for _, v := range c.Includes {
		includes += v.String()
	}

	for _, v := range c.Excludes {
		excludes += v.String()
	}

	params := []any{
		slog.Int("parallel", c.Parallel),
	}

	if includes != "" {
		params = append(params, slog.String("includes", includes))
	}

	if excludes != "" {
		params = append(params, slog.String("excludes", excludes))
	}

	if c.Timeout != 0 {
		params = append(params, slog.String("timeout", c.Timeout.String()))
	}

	if c.MaxFileSize != -1 {
		slog.String("max-file-size", bytesutil.PrettyByteSize(c.MaxFileSize))
	}

	slog.Info(
		"Processing:",
		params...,
	)

	c.Image.Print("Image:")
	c.Video.Print("Video:")
}

type Config struct {
	Dir           string
	Port          int
	Streaming     []*regexp.Regexp
	IndexFileSave bool
	Process       ProcessConfig
}

func (c Config) Print() {
	streamingOriginalPatterns := make([]string, len(c.Streaming))
	for i, v := range c.Streaming {
		streamingOriginalPatterns[i] = v.String()
	}

	slog.Info(
		"Config:",
		slog.String("dir", c.Dir),
		slog.Int("port", c.Port),
		slog.String("streaming", strings.Join(streamingOriginalPatterns, ",")),
		slog.Bool("index-file-saving", c.IndexFileSave),
	)
	c.Process.Print()
}

func DefaultRawConfig() RawConfig {
	return RawConfig{
		Dir:           os.Getenv("PWD"),
		Parallel:      runtime.NumCPU(),
		Port:          defaultPort,
		Video:         true,
		Images:        true,
		IndexFileSave: true,
		MaxImages:     -1,
		MaxVideos:     -1,
		MaxFileSize:   "-1B",
		Streaming:     "\\.(flv|f4v|avi)$",
		MediaTimeout:  "2m",
	}
}

func NewConfig(raw RawConfig) Config {
	if raw.Port > maxPortNumber {
		panic(fmt.Errorf("flag port value out of range[0-65535]: %v", raw.Port)) //nolint:err113
	}

	return Config{
		Dir:           raw.Dir,
		Port:          raw.Port,
		Streaming:     getRegularExpressions(raw.Streaming),
		IndexFileSave: raw.IndexFileSave,
		Process: ProcessConfig{
			Timeout:     getDuration(raw.MediaTimeout),
			Parallel:    raw.Parallel,
			Video:       MediaTypeConfig{raw.Video, raw.MaxVideos},
			Image:       MediaTypeConfig{raw.Images, raw.MaxImages},
			Includes:    getRegularExpressions(raw.Includes),
			Excludes:    getRegularExpressions(raw.Excludes),
			MaxFileSize: getMaxFileSize(raw.MaxFileSize),
		},
	}
}

func getDuration(durationEncoded string) time.Duration {
	duration, err := time.ParseDuration(durationEncoded)
	if err != nil {
		panic(err)
	}

	return duration
}

func getMaxFileSize(value string) int64 {
	maxFileSize := int64(-1)
	if value != "" {
		maxFileSize = bytesutil.ParseByteSize(value)
	}

	return maxFileSize
}

func getRegularExpressions(list string) []*regexp.Regexp {
	patterns := strings.Split(list, ",")
	compiled := make([]*regexp.Regexp, len(patterns))

	for i, pattern := range patterns {
		reg, err := regexp.Compile(pattern)
		if err != nil {
			panic(err)
		}

		compiled[i] = reg
	}

	return compiled
}
