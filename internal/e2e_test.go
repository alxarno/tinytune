package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/alxarno/tinytune/pkg/index"
	"github.com/alxarno/tinytune/pkg/preview"
	"github.com/stretchr/testify/require"
)

func TestE2e(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	index := getIndex(ctx)

	streamingFiles := map[string]struct{}{
		"video/sample_960x400_ocean_with_audio.flv": {},
	}

	server := NewServer(
		ctx,
		WithSource(index),
		WithPWD("../test"),
		WithDebug(true),
		WithStreaming(streamingFiles),
		WithDry(),
	)
	serverHandler := server.registerHandlers(true)

	tests := []struct {
		Name string
		Path string
		Type string
	}{
		{
			Name: "index",
			Path: "/",
			Type: "text",
		},
		{
			Name: "dir",
			Path: "/d/6a4afe45d4/",
			Type: "text",
		},
		{
			Name: "search",
			Path: `/s?query=ocean_with_audio.flv`,
			Type: "text",
		},
		{
			Name: "preview-image",
			Path: `/preview/ff8609b730/`,
		},
		{
			Name: "preview-video",
			Path: `/preview/39508f59cd/`,
		},
		{
			Name: "origin",
			Path: `/origin/024eb2ebdb/`,
		},
		{
			Name: "hls-index",
			Path: `/rts/005f6b0265/`,
			Type: "text",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.Name, func(test *testing.T) {
			test.Parallel()

			require := require.New(test)
			fixtureFile, err := os.OpenFile(fmt.Sprintf("../e2e/%s.fixture", testCase.Name), os.O_RDWR|os.O_CREATE, 0755)
			require.NoError(err)

			defer fixtureFile.Close()

			r := httptest.NewRequest(http.MethodGet, testCase.Path, nil)
			w := httptest.NewRecorder()
			serverHandler.ServeHTTP(w, r)
			body, err := io.ReadAll(w.Body)
			require.NoError(err)
			fixture, err := io.ReadAll(fixtureFile)
			require.NoError(err)

			if testCase.Type == "text" {
				require.Equal(string(fixture), string(body))
			} else {
				require.Equal(fixture, body)
			}
		})
	}
}

func getIndex(ctx context.Context) *index.Index {
	previewer, err := preview.NewPreviewer( //nolint:contextcheck
		preview.WithTimeout(time.Second * 10),
	)
	PanicError(err)

	files, err := NewCrawlerOS("../test").Scan("../test/index.tinytune")
	PanicError(err)

	index, err := index.NewIndex(
		ctx,
		nil,
		index.WithFiles(files),
		index.WithPreview(previewer),
		index.WithWorkers(runtime.NumCPU()),
	)
	PanicError(err)

	return &index
}
