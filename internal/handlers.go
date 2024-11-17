package internal

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"time"

	"github.com/alxarno/tinytune/pkg/httputil"
	"github.com/alxarno/tinytune/pkg/index"
	"github.com/justinas/alice"
	"golang.org/x/exp/maps"
)

type PageData struct {
	Items      []*index.Meta
	Path       []*index.Meta
	Zoom       string
	Sorts      []string
	ActiveSort string
	Search     string
}

func (s Server) newPageData() PageData {
	return PageData{
		Items: []*index.Meta{},
		Path:  []*index.Meta{},
		Sorts: []string{},
	}
}

func logWriteErr(_ int, err error) {
	if err != nil {
		slog.Error(fmt.Sprintf("Write failed: %v", err))
	}
}

func applyCookies(r *http.Request, data PageData) PageData {
	sorts := maps.Keys(getSorts())
	slices.Sort(sorts)
	data.Sorts = sorts

	if cookie, err := r.Cookie("zoom"); err == nil {
		data.Zoom = cookie.Value
	} else {
		data.Zoom = "medium"
	}

	if cookie, err := r.Cookie("sort"); err == nil {
		if decodedValue, err := url.QueryUnescape(cookie.Value); err != nil {
			data.ActiveSort = "Type"
		} else {
			data.ActiveSort = decodedValue
		}
	} else {
		data.ActiveSort = "Type"
	}

	if s, ok := getSorts()[data.ActiveSort]; ok {
		data.Items = s(data.Items)
	}

	return data
}

func (s Server) handleBasicTemplate(data PageData, w http.ResponseWriter, r *http.Request) {
	data = applyCookies(r, data)

	w.WriteHeader(http.StatusOK)

	if err := s.templates["index.html"].ExecuteTemplate(w, "index", data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s Server) indexHandler() httputil.MetaHTTPHandler {
	return func(dir, _ *index.Meta, w http.ResponseWriter, r *http.Request) {
		err := error(nil) //nolint:wastedassign
		data := s.newPageData()

		if data.Items, err = s.source.PullChildren(dir.ID); err != nil {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		if data.Path, err = s.source.PullPaths(dir.ID); err != nil {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		s.handleBasicTemplate(data, w, r)
	}
}

func (s Server) searchHandler() httputil.MetaHTTPHandler {
	return func(dir, _ *index.Meta, w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Expose-Headers", "Hx-Push-Url")
		w.Header().Set("HX-Push-Url", r.RequestURI)

		err := error(nil) //nolint:wastedassign
		data := s.newPageData()

		params, _ := url.ParseQuery(r.URL.RawQuery)
		data.Search = params.Get("query")

		if data.Search == "" {
			http.Redirect(w, r, "/", http.StatusNotFound)
		}

		if data.Search, err = url.QueryUnescape(data.Search); err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if data.Path, err = s.source.PullPaths(dir.ID); err != nil {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		data.Path = append(data.Path, &index.Meta{Name: "Search"})
		data.Items = s.source.Search(data.Search, dir.ID)
		s.handleBasicTemplate(data, w, r)
	}
}

func (s Server) previewHandler() httputil.MetaHTTPHandler {
	return func(_, file *index.Meta, w http.ResponseWriter, _ *http.Request) {
		data, err := s.source.PullPreview(file.ID)
		if err != nil || len(data) == 0 {
			w.WriteHeader(http.StatusNotFound)

			return
		}
		// Cache for week
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Cache-Control", "max-age=604800")
		w.Header().Add("Content-Type", http.DetectContentType(data))
		logWriteErr(w.Write(data))
	}
}

func (s Server) originHandler() httputil.MetaHTTPHandler {
	return func(_, file *index.Meta, w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "max-age=3600")
		http.ServeFile(w, r, filepath.Join(s.pwd, string(file.RelativePath)))
	}
}

func (s Server) hlsIndexHandler() httputil.MetaHTTPHandler {
	return func(_, file *index.Meta, w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		if err := pullHLSIndex(file, w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (s Server) hlsChunkHandler() httputil.MetaHTTPHandler {
	return func(_, file *index.Meta, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		timeout := 5 * time.Second //nolint:gomnd,mnd
		chunkID := r.PathValue("chunkID")
		ctx := r.Context()

		if err := pullHLSChunk(ctx, file, chunkID, timeout, w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (s Server) registerHandlers(silent bool) http.Handler {
	mux := http.NewServeMux()

	chain := alice.New()
	if !silent {
		chain = chain.Append(httputil.LoggingHandler)
	}

	register := func(route string, h httputil.MetaHTTPHandler) {
		mux.Handle(route, chain.Then(httputil.MetaHandler(h, s.source)))
	}

	register("GET /", s.indexHandler())
	register("GET /d/{dirID}/", s.indexHandler())

	register("GET /s", s.searchHandler())
	register("GET /s/{dirID}/", s.searchHandler())

	register("GET /preview/{fileID}/", s.previewHandler())

	register("GET /origin/{fileID}/", s.originHandler())

	register("GET /rts/{fileID}/", s.hlsIndexHandler())

	register("GET /rts/{fileID}/{chunkID}/", s.hlsChunkHandler())

	staticHandler := http.StripPrefix("/static", http.FileServer(http.FS(s.getAssets())))
	mux.Handle("GET /static/", chain.Then(staticHandler))

	return mux
}
