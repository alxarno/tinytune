package internal

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"slices"
	"time"

	"github.com/alxarno/tinytune/pkg/httputil"
	"github.com/alxarno/tinytune/pkg/index"
	"github.com/alxarno/tinytune/web/assets"
	"github.com/alxarno/tinytune/web/templates"
	"github.com/justinas/alice"
	"golang.org/x/exp/maps"
)

type source interface {
	PullChildren(ID string) ([]*index.Meta, error)
	PullPreview(ID string) ([]byte, error)
	PullPaths(ID string) ([]*index.Meta, error)
	Pull(ID string) (*index.Meta, error)
	Search(query string, dir string) []*index.Meta
}

type Server struct {
	templates map[string]*template.Template
	source    source
	port      int
	debugMode bool
}

func logWriteErr(_ int, err error) {
	if err != nil {
		slog.Error(fmt.Sprintf("Write failed: %v", err))
	}
}

func getSorts() map[string]metaSortFunc {
	return map[string]metaSortFunc{
		"A-Z":            metaSortAlphabet,
		"Z-A":            metaSortAlphabetReverse,
		"Last Modified":  metaSortLastModified,
		"First Modified": metaSortFirstModified,
		"Type":           metaSortType,
		"Size":           metaSortSize,
	}
}

func (s *Server) getTemplates() fs.FS {
	if s.debugMode {
		return os.DirFS("./web/templates/")
	}

	return templates.Templates
}

func (s *Server) getAssets() fs.FS {
	if s.debugMode {
		return os.DirFS("./web/assets/")
	}

	return assets.Assets
}

func (s *Server) loadTemplates() {
	funcs := template.FuncMap{
		"ext":    extension,
		"width":  width,
		"height": height,
		"eqMinusOne": func(x int, y int) bool {
			return x == y-1
		},
		"dur": durationPrint,
	}
	s.templates = make(map[string]*template.Template)
	s.templates["index.html"] = template.Must(template.New("index.html").Funcs(funcs).ParseFS(s.getTemplates(), "*.html"))
}

type PageData struct {
	Items      []*index.Meta
	Path       []*index.Meta
	Zoom       string
	Sorts      []string
	ActiveSort string
	Search     string
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

func (s *Server) indexHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := error(nil) //nolint:wastedassign
		data := PageData{
			Items: []*index.Meta{},
			Path:  []*index.Meta{},
			Sorts: []string{},
		}

		if data.Items, err = s.source.PullChildren(r.PathValue("dirID")); err != nil {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		if data.Path, err = s.source.PullPaths(r.PathValue("dirID")); err != nil {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		data = applyCookies(r, data)

		w.WriteHeader(http.StatusOK)

		if err := s.templates["index.html"].ExecuteTemplate(w, "index", data); err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}
	})
}

func (s *Server) searchHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Expose-Headers", "Hx-Push-Url")
		w.Header().Set("HX-Push-Url", r.RequestURI)

		err := error(nil) //nolint:wastedassign
		data := PageData{
			Items: []*index.Meta{},
			Path:  []*index.Meta{},
			Sorts: []string{},
		}
		params, _ := url.ParseQuery(r.URL.RawQuery)
		data.Search = params.Get("query")

		if data.Search == "" {
			http.Redirect(w, r, "/", http.StatusNotFound)
		}

		if data.Search, err = url.QueryUnescape(data.Search); err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		dirID := r.PathValue("dirID")

		if dirID != "" {
			if data.Path, err = s.source.PullPaths(r.PathValue("dirID")); err != nil {
				w.WriteHeader(http.StatusNotFound)

				return
			}
		}

		data.Items = s.source.Search(data.Search, dirID)
		data.Path = append(data.Path, &index.Meta{
			Name: "Search",
		})
		data = applyCookies(r, data)

		w.WriteHeader(http.StatusOK)

		if err := s.templates["index.html"].ExecuteTemplate(w, "index", data); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func (s *Server) previewHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := s.source.PullPreview(r.PathValue("fileID"))
		if err != nil || len(data) == 0 {
			w.WriteHeader(http.StatusNotFound)

			return
		}
		// Cache for week
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Cache-Control", "max-age=604800")
		w.Header().Add("Content-Type", http.DetectContentType(data))
		logWriteErr(w.Write(data))
	})
}

func (s *Server) originHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		meta, err := s.source.Pull(r.PathValue("fileID"))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404")

			return
		}
		// Cache for hour
		w.Header().Add("Cache-Control", "max-age=3600")
		http.ServeFile(w, r, meta.Path)
	})
}

type ServerOption func(*Server)

func WithSource(source source) ServerOption {
	return func(s *Server) {
		s.source = source
	}
}

func WithPort(port int) ServerOption {
	return func(s *Server) {
		s.port = port
	}
}

func WithDebug(debug bool) ServerOption {
	return func(s *Server) {
		s.debugMode = debug
	}
}

func NewServer(ctx context.Context, opts ...ServerOption) *Server {
	server := &Server{}

	for _, opt := range opts {
		opt(server)
	}

	server.loadTemplates()

	mux := http.NewServeMux()
	serverTimeoutSeconds := 30
	httpServer := &http.Server{
		BaseContext:       func(net.Listener) context.Context { return ctx },
		Addr:              fmt.Sprintf(":%d", server.port),
		Handler:           mux,
		ReadHeaderTimeout: time.Second * time.Duration(serverTimeoutSeconds),
	}
	chain := alice.New(httputil.LoggingHandler)
	staticHandler := http.StripPrefix("/static", http.FileServer(http.FS(server.getAssets())))
	mux.Handle("GET /", chain.Then(server.indexHandler()))
	mux.Handle("GET /d/{dirID}/", chain.Then(server.indexHandler()))
	mux.Handle("GET /s", chain.Then(server.searchHandler()))
	mux.Handle("GET /s/{dirID}/", chain.Then(server.searchHandler()))
	mux.Handle("GET /preview/{fileID}/", chain.Then(server.previewHandler()))
	mux.Handle("GET /origin/{fileID}/", chain.Then(server.originHandler()))
	mux.Handle("GET /static/", chain.Then(staticHandler))

	go func() { PanicError(httpServer.ListenAndServe()) }()

	return server
}
