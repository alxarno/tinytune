package internal

import (
	"context"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/alxarno/tinytune/pkg/httputil"
	"github.com/alxarno/tinytune/pkg/index"
	"github.com/justinas/alice"
	"golang.org/x/exp/maps"
)

type source interface {
	PullChildren(string) ([]*index.IndexMeta, error)
	PullPreview(string) ([]byte, error)
	PullPaths(string) ([]*index.IndexMeta, error)
	Pull(string) (*index.IndexMeta, error)
	Search(string, string) []*index.IndexMeta
}

type server struct {
	templates map[string]*template.Template
	source    source
	port      int
}

func getSorts() map[string]metaSortFunc {
	m := map[string]metaSortFunc{
		"A-Z":            metaSortAlphabet,
		"Z-A":            metaSortAlphabetReverse,
		"Last Modified":  metaSortLastModified,
		"First Modified": metaSortFirstModified,
		"Type":           metaSortType,
		"Size":           metaSortSize,
	}
	return m
}

func (s *server) loadtemplates() {
	funcs := template.FuncMap{
		"ext": func(name string) string {
			extension := path.Ext(name)
			if extension != "" {
				return extension[1:]
			}
			return ""
		},
		"width": func(res string) string {
			return strings.Split(res, "x")[0]
		},
		"height": func(res string) string {
			return strings.Split(res, "x")[1]
		},
		"eqMinusOne": func(x int, y int) bool {
			return x == y-1
		},
		"dur": func(d time.Duration) string {
			result := ""
			if int(d.Hours()) != 0 {
				result += fmt.Sprintf("%02d:", int(d.Hours()))
			}
			result += fmt.Sprintf("%02d:", int(d.Minutes()))
			result += fmt.Sprintf("%02d", int(d.Seconds()))
			return result
		},
	}
	s.templates = make(map[string]*template.Template)
	s.templates["index.html"] = template.Must(template.New("index.html").Funcs(funcs).ParseGlob("./web/templates/*.html"))
}

type PageData struct {
	Items      []*index.IndexMeta
	Path       []*index.IndexMeta
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

func (s *server) indexHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := error(nil)
		data := PageData{
			Items: []*index.IndexMeta{},
			Path:  []*index.IndexMeta{},
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
		s.templates["index.html"].ExecuteTemplate(w, "index", data)
	})
}

func (s *server) searchHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Expose-Headers", "Hx-Push-Url")
		w.Header().Set("HX-Push-Url", r.RequestURI)
		err := error(nil)
		data := PageData{
			Items: []*index.IndexMeta{},
			Path:  []*index.IndexMeta{},
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
		data.Path = append(data.Path, &index.IndexMeta{
			Name: "Search",
		})
		data = applyCookies(r, data)
		w.WriteHeader(http.StatusOK)
		s.templates["index.html"].ExecuteTemplate(w, "index", data)
	})
}

func (s *server) previewHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := s.source.PullPreview(r.PathValue("fileID"))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404")
			return
		}
		// Cache for week
		w.Header().Add("Cache-Control", "max-age=604800")
		w.Header().Add("Content-Type", http.DetectContentType(data))
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})
}

func (s *server) originHandler() http.Handler {
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

type ServerOption func(*server)

func WithSource(source source) ServerOption {
	return func(s *server) {
		s.source = source
	}
}

func WithPort(port int) ServerOption {
	return func(s *server) {
		s.port = port
	}
}

func NewServer(ctx context.Context, opts ...ServerOption) *server {
	server := server{}
	server.loadtemplates()
	for _, opt := range opts {
		opt(&server)
	}
	mux := http.NewServeMux()
	httpServer := &http.Server{
		BaseContext: func(net.Listener) context.Context { return ctx },
		Addr:        fmt.Sprintf(":%d", server.port),
		Handler:     mux,
	}
	chain := alice.New(httputil.LoggingHandler)
	staticHandler := http.StripPrefix("/static", http.FileServer(http.Dir("./web/assets/")))
	mux.Handle("GET /", chain.Then(server.indexHandler()))
	mux.Handle("GET /d/{dirID}/", chain.Then(server.indexHandler()))
	mux.Handle("GET /s", chain.Then(server.searchHandler()))
	mux.Handle("GET /s/{dirID}/", chain.Then(server.searchHandler()))
	mux.Handle("GET /preview/{fileID}/", chain.Then(server.previewHandler()))
	mux.Handle("GET /origin/{fileID}/", chain.Then(server.originHandler()))
	mux.Handle("GET /static/", chain.Then(staticHandler))

	go httpServer.ListenAndServe()
	return nil
}
