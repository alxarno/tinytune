package httputil

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/alxarno/tinytune/pkg/bytesutil"
	"github.com/alxarno/tinytune/pkg/log"
	"github.com/lmittmann/tint"
)

type loggingHandler struct {
	handler http.Handler
	logger  *slog.Logger
}

func LoggingHandler(h http.Handler) http.Handler {
	const (
		ansiReset        = "\033[0m"
		ansiBrightRed    = "\033[91m"
		ansiBrightGreen  = "\033[92m"
		ansiBrightYellow = "\033[93m"
	)
	logger := log.GetLogger(log.WithOption(func(opt *tint.Options) {
		opt.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
			errorsCodes := []string{"GET 404", "GET 500"}
			if opt.NoColor {
				return a
			}
			if a.Key == slog.LevelKey {
				a = slog.Attr{}
			}
			if a.Key == slog.MessageKey && a.Value.String() == "GET 200" {
				a.Value = slog.StringValue(ansiBrightGreen + a.Value.String() + ansiReset)
			}
			if a.Key == slog.MessageKey && slices.Contains(errorsCodes, a.Value.String()) {
				a.Value = slog.StringValue(ansiBrightRed + a.Value.String() + ansiReset)
			}
			if a.Key == slog.MessageKey && a.Value.String() == "GET 302" {
				a.Value = slog.StringValue(ansiBrightYellow + a.Value.String() + ansiReset)
			}
			return a
		}
	}))
	return &loggingHandler{
		handler: h,
		logger:  logger,
	}
}

func (h *loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wrapped := wrapResponseWriter(w)
	h.handler.ServeHTTP(wrapped, r)
	printFunc := h.logger.Info
	if wrapped.status == http.StatusNotFound {
		printFunc = h.logger.Error
	} else if wrapped.status != http.StatusOK {
		printFunc = h.logger.Warn
	}

	printFunc(
		fmt.Sprintf("%s %d", r.Method, wrapped.Status()),
		slog.String("path", r.URL.Path),
		slog.String("response", bytesutil.PrettyByteSize(wrapped.size)),
		slog.String("ip", r.RemoteAddr),
		slog.String("user-agent", r.UserAgent()))
}
