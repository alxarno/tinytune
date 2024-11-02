package internal

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/alxarno/tinytune/pkg/httputil"
	"github.com/justinas/alice"
)

type server struct {
}

func exactlyPathHandler(path string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404")
			return
		}
		h.ServeHTTP(w, r)
	})
}

func getRedirectHandler(to string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, to, http.StatusFound)
	})
}

func dirServeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Dir %s!", r.PathValue("dirID"))
	})
}

func previewHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Preview %s!", r.PathValue("fileID"))
	})
}

func originHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Origin %s!", r.PathValue("fileID"))
	})
}

func NewServer(ctx context.Context) *server {
	chain := alice.New(httputil.LoggingHandler)
	mux := http.NewServeMux()
	rootHandler := exactlyPathHandler("/", getRedirectHandler("/dir/root"))
	mux.Handle("GET /", chain.Then(rootHandler))
	mux.Handle("GET /dir/{dirID}/", chain.Then(dirServeHandler()))
	mux.Handle("GET /preview/{fileID}/", chain.Then(previewHandler()))
	mux.Handle("GET /origin/{fileID}/", chain.Then(originHandler()))

	srv := &http.Server{
		BaseContext: func(net.Listener) context.Context { return ctx },
		Addr:        ":8080",
		Handler:     mux,
	}
	go srv.ListenAndServe()
	return nil
}
