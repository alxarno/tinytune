package httputil

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/alxarno/tinytune/pkg/index"
)

type MetaHTTPHandler func(dir *index.Meta, file *index.Meta, w http.ResponseWriter, r *http.Request)

type source interface {
	Pull(id index.ID) (*index.Meta, error)
}

type metaHandler struct {
	handler MetaHTTPHandler
	source  source
}

func MetaHandler(handler MetaHTTPHandler, source source) http.Handler {
	return &metaHandler{
		handler: handler,
		source:  source,
	}
}

func (h *metaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dir := &index.Meta{ID: index.ID("")}
	var file *index.Meta

	if len(r.PathValue("dirID")) != 0 {
		meta, err := h.source.Pull(index.ID(r.PathValue("dirID")))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404")

			return
		}

		dir = meta
	}

	if len(r.PathValue("fileID")) != 0 {
		replacedRTS := strings.ReplaceAll(r.PathValue("fileID"), ".m3u8", "")

		meta, err := h.source.Pull(index.ID(replacedRTS))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404")

			return
		}

		file = meta
	}

	h.handler(dir, file, w, r)
}
