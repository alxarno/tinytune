package internal

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/alxarno/tinytune/pkg/bytesutil"
)

const defaultPort = 8080
const maxPortNumber = 65536

type RawConfig struct {
	Dir          string
	Parallel     int
	Video        bool
	Images       bool
	Acceleration bool
	MaxImages    int64
	MaxVideos    int64
	Includes     string
	Excludes     string
	MaxFileSize  string
	Port         int
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
	Parallel     int
	Video        MediaTypeConfig
	Image        MediaTypeConfig
	Acceleration bool
	Includes     []*regexp.Regexp
	Excludes     []*regexp.Regexp
	MaxFileSize  int64
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
		slog.Bool("acceleration", c.Acceleration),
	}

	if includes != "" {
		params = append(params, slog.String("includes", includes))
	}

	if excludes != "" {
		params = append(params, slog.String("excludes", excludes))
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
	Dir     string
	Port    int
	Process ProcessConfig
}

func (c Config) Print() {
	slog.Info("Config:", slog.String("dir", c.Dir), slog.Int("port", c.Port))
	c.Process.Print()
}

func DefaultRawConfig() RawConfig {
	return RawConfig{
		Dir:          os.Getenv("PWD"),
		Parallel:     runtime.NumCPU(),
		Port:         defaultPort,
		Video:        true,
		Images:       true,
		Acceleration: true,
		MaxImages:    -1,
		MaxVideos:    -1,
	}
}

func NewConfig(raw RawConfig) Config {
	if raw.Port > maxPortNumber {
		panic(fmt.Errorf("flag port value out of range[0-65535]: %v", raw.Port)) //nolint:err113
	}

	return Config{
		Dir:  raw.Dir,
		Port: raw.Port,
		Process: ProcessConfig{
			Parallel:     raw.Parallel,
			Video:        MediaTypeConfig{raw.Video, raw.MaxVideos},
			Image:        MediaTypeConfig{raw.Images, raw.MaxImages},
			Acceleration: raw.Acceleration,
			Includes:     getRegularExpressions(raw.Includes),
			Excludes:     getRegularExpressions(raw.Excludes),
			MaxFileSize:  getMaxFileSize(raw.MaxFileSize),
		},
	}
}

func getMaxFileSize(value string) int64 {
	maxFileSize := int64(-1)
	if value != "" {
		maxFileSize = 1024
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
