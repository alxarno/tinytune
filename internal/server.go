package internal

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"net"
	"net/http"
	//nolint:gosec
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/alxarno/tinytune/pkg/index"
	"github.com/alxarno/tinytune/web/assets"
	"github.com/alxarno/tinytune/web/templates"
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
	streaming map[string]struct{}
	source    source
	pwd       string
	port      int
	debugMode bool
	dryMode   bool
}

func (s Server) getTemplates() fs.FS {
	if s.dryMode {
		return os.DirFS("../web/templates/")
	}

	if s.debugMode {
		return os.DirFS("./web/templates/")
	}

	return templates.Templates
}

func (s Server) getAssets() fs.FS {
	if s.dryMode {
		return os.DirFS("../web/assets/")
	}

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

func WithStreaming(files map[string]struct{}) ServerOption {
	return func(s *Server) {
		s.streaming = files
	}
}

func WithDry() ServerOption {
	return func(s *Server) {
		s.dryMode = true
	}
}

func NewServer(ctx context.Context, opts ...ServerOption) *Server {
	server := &Server{}

	for _, opt := range opts {
		opt(server)
	}

	server.templates = loadTemplates(server.getTemplates(), server.streaming)

	serverTimeoutSeconds := 30
	httpServer := &http.Server{
		BaseContext:       func(net.Listener) context.Context { return ctx },
		Addr:              fmt.Sprintf(":%d", server.port),
		Handler:           server.registerHandlers(false),
		ReadHeaderTimeout: time.Second * time.Duration(serverTimeoutSeconds),
	}

	if !server.dryMode {
		go func() { PanicError(httpServer.ListenAndServe()) }()
	}

	return server
}
