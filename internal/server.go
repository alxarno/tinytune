package internal

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/alxarno/tinytune/pkg/httputil"
	"github.com/alxarno/tinytune/pkg/index"
	"github.com/alxarno/tinytune/web/assets"
	"github.com/alxarno/tinytune/web/templates"
	"github.com/justinas/alice"
)

type source interface {
	PullChildren(dirID index.ID) ([]*index.Meta, error)
	PullPreview(fileID index.ID) ([]byte, error)
	PullPaths(dirID index.ID) ([]*index.Meta, error)
	Pull(fileID index.ID) (*index.Meta, error)
	Search(query string, dirID index.ID) []*index.Meta
}

type Server struct {
	templates map[string]*template.Template
	source    source
	pwd       string
	port      int
	debugMode bool
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

func WithPWD(dir string) ServerOption {
	return func(s *Server) {
		s.pwd = dir
	}
}

func NewServer(ctx context.Context, opts ...ServerOption) *Server {
	server := &Server{}

	for _, opt := range opts {
		opt(server)
	}

	server.templates = loadTemplates(server.getTemplates())

	mux := http.NewServeMux()
	serverTimeoutSeconds := 30
	httpServer := &http.Server{
		BaseContext:       func(net.Listener) context.Context { return ctx },
		Addr:              fmt.Sprintf(":%d", server.port),
		Handler:           mux,
		ReadHeaderTimeout: time.Second * time.Duration(serverTimeoutSeconds),
	}
	chain := alice.New(httputil.LoggingHandler)
	register := func(route string, h httputil.MetaHTTPHandler) {
		mux.Handle(route, chain.Then(httputil.MetaHandler(h)))
	}

	register("GET /", server.indexHandler())
	register("GET /d/{dirID}/", server.indexHandler())

	register("GET /s", server.searchHandler())
	register("GET /s/{dirID}/", server.searchHandler())

	register("GET /preview/{fileID}/", server.previewHandler())

	register("GET /origin/{fileID}/", server.originHandler())

	register("GET /rts/{fileID}", server.hlsIndexHandler())

	register("GET /rts/{fileID}/{chunkID}", server.hlsChunkHandler())

	staticHandler := http.StripPrefix("/static", http.FileServer(http.FS(server.getAssets())))
	mux.Handle("GET /static/", chain.Then(staticHandler))

	go func() { PanicError(httpServer.ListenAndServe()) }()

	return server
}
