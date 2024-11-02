package log

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func init() {
	slog.SetDefault(GetLogger())
}

type Logger struct{}

type LogOption func(*tint.Options)

func WithOption(applier func(opt *tint.Options)) LogOption {
	return func(o *tint.Options) {
		applier(o)
	}
}

func GetLogger(opts ...LogOption) *slog.Logger {
	w := os.Stdout
	noColor := !isatty.IsTerminal(w.Fd())
	options := &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.DateTime,
		NoColor:    noColor,
	}
	for _, option := range opts {
		option(options)
	}
	return slog.New(
		tint.NewHandler(os.Stdout, options),
	)
}
