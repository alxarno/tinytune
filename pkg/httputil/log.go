package httputil

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/alxarno/tinytune/pkg/bytesutil"
	"github.com/alxarno/tinytune/pkg/logging"
	"github.com/lmittmann/tint"
)

type loggingHandler struct {
	handler http.Handler
	logger  *slog.Logger
}

func LoggingHandler(handler http.Handler) http.Handler {
	const (
		ansiReset        = "\033[0m"
		ansiBrightRed    = "\033[91m"
		ansiBrightGreen  = "\033[92m"
		ansiBrightYellow = "\033[93m"
		ansiBrightBlue   = "\033[94m"
	)
	valueToColor := map[string]string{
		"GET 404": ansiBrightRed,
		"GET 500": ansiBrightRed,
		"GET 200": ansiBrightGreen,
		"GET 206": ansiBrightBlue,
		"GET 302": ansiBrightYellow,
		"GET 303": ansiBrightYellow,
	}
	logger := logging.Get(logging.WithOption(func(opt *tint.Options) {
		opt.ReplaceAttr = func(_ []string, attribute slog.Attr) slog.Attr {
			if opt.NoColor {
				return attribute
			}

			if attribute.Key == slog.LevelKey {
				attribute = slog.Attr{}
			}

			code, ok := valueToColor[attribute.Value.String()]
			if ok {
				attribute.Value = slog.StringValue(code + attribute.Value.String() + ansiReset)
			}

			return attribute
		}
	}))

	return &loggingHandler{
		handler: handler,
		logger:  logger,
	}
}

func subString(input string, start int, length int) string {
	asRunes := []rune(input)
	postfix := ""

	if start >= len(asRunes) {
		return ""
	}

	if start+length > len(asRunes) {
		length = len(asRunes) - start
	} else {
		postfix = "..."
	}

	return string(asRunes[start:start+length]) + postfix
}

func (h *loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wrapped := wrapResponseWriter(w)
	h.handler.ServeHTTP(wrapped, r)
	printFunc := h.logger.Info
	maxUserAgentLength := 30

	if wrapped.status == http.StatusNotFound {
		printFunc = h.logger.Error
	} else if wrapped.status != http.StatusOK {
		printFunc = h.logger.Warn
	}

	printFunc(
		fmt.Sprintf("%s %d", r.Method, wrapped.Status()),
		slog.String("path", r.URL.Path),
		slog.String("type", strings.Replace(w.Header().Get("Content-Type"), " charset=utf-8", "", 1)),
		slog.String("response", bytesutil.PrettyByteSize(wrapped.size)),
		slog.String("ip", r.RemoteAddr),
		slog.String("user-agent", subString(r.UserAgent(), 0, maxUserAgentLength)))
}
