package internal

import (
	"context"
	"fmt"
	"html/template"
	"net"
	"net/http"

	"github.com/alxarno/tinytune/pkg/httputil"
	"github.com/alxarno/tinytune/pkg/index"
	"github.com/justinas/alice"
)

type source interface {
	PullChildren(id string) []index.IndexMeta
}

type server struct {
	Templates map[string]*template.Template
	Source    source
}

func (s *server) LoadTemplates() {
	s.Templates = make(map[string]*template.Template)
	s.Templates["index.html"] = template.Must(template.ParseGlob("./web/templates/*.html"))
}

// func exactlyPathHandler(path string, h http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if r.URL.Path != path {
// 			w.WriteHeader(http.StatusNotFound)
// 			fmt.Fprint(w, "404")
// 			return
// 		}
// 		h.ServeHTTP(w, r)
// 	})
// }

// func getRedirectHandler(to string) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		http.Redirect(w, r, to, http.StatusFound)
// 	})
// }

func (s *server) indexHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404")
			return
		}
		w.WriteHeader(http.StatusOK)
		s.Templates["index.html"].ExecuteTemplate(w, "index", s.Source.PullChildren(""))
	})
}

func (s *server) dirServeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Dir %s!", r.PathValue("dirID"))
	})
}

func (s *server) previewHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Preview %s!", r.PathValue("fileID"))
	})
}

func (s *server) originHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Origin %s!", r.PathValue("fileID"))
	})
}

// func (s *server) staticHandler() http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 		fmt.Fprintf(w, "Static %s!", r.PathValue("fileID"))
// 	})
// }

type ServerOption func(*server)

func WithSource(source source) ServerOption {
	return func(s *server) {
		s.Source = source
	}
}

func NewServer(ctx context.Context, opts ...ServerOption) *server {
	server := server{}
	server.LoadTemplates()
	for _, opt := range opts {
		opt(&server)
	}

	mux := http.NewServeMux()
	httpServer := &http.Server{
		BaseContext: func(net.Listener) context.Context { return ctx },
		Addr:        ":8080",
		Handler:     mux,
	}
	chain := alice.New(httputil.LoggingHandler)
	// rootHandler := exactlyPathHandler("/", getRedirectHandler("/dir/root"))
	mux.Handle("GET /", chain.Then(server.indexHandler()))
	mux.Handle("GET /dir/{dirID}/", chain.Then(server.dirServeHandler()))
	mux.Handle("GET /preview/{fileID}/", chain.Then(server.previewHandler()))
	mux.Handle("GET /origin/{fileID}/", chain.Then(server.originHandler()))
	mux.Handle("GET /static/", chain.Then(http.StripPrefix("/static", http.FileServer(http.Dir("./web/assets/")))))

	go httpServer.ListenAndServe()
	return nil
}
