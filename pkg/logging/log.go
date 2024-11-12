package logging

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

type Logger struct{}

type Option func(*tint.Options)

func WithOption(applier func(opt *tint.Options)) Option {
	return func(o *tint.Options) {
		applier(o)
	}
}

func Get(opts ...Option) *slog.Logger {
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
