package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/alxarno/tinytune/internal"
	"github.com/alxarno/tinytune/pkg/bytesutil"
	"github.com/alxarno/tinytune/pkg/index"
	"github.com/urfave/cli/v2"
)

const INDEX_FILE_NAME = "index.tinytune"

type config struct {
	dir string
}

func main() {
	c := config{dir: os.Getenv("PWD")}
	app := &cli.App{
		Name:  "tinytune",
		Usage: "tiny media server",
		Action: func(ctx *cli.Context) error {
			if ctx.Args().Len() != 0 {
				c.dir = ctx.Args().First()
			}
			start(c)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func start(c config) {
	ctx := gracefulShutdownCtx()
	indexFilePath := filepath.Join(c.dir, INDEX_FILE_NAME)

	files, err := internal.NewCrawlerOS(c.dir).Scan(indexFilePath)
	if err != nil {
		log.Fatal(err)
	}
	indexFile, err := os.OpenFile(indexFilePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer indexFile.Close()
	indexFileReader := io.Reader(indexFile)

	fileInfo, err := indexFile.Stat()
	if err != nil {
		log.Fatal(err)
	}

	if fileInfo.Size() != 0 {
		slog.Info(
			"Found index file",
			slog.String("size", bytesutil.PrettyByteSize(fileInfo.Size())),
			slog.String("path", indexFilePath),
		)
	} else {
		slog.Info("Created new index file", slog.String("path", indexFilePath))
		indexFileReader = nil
	}
	slog.Info("Indexing started")
	indexProgressBar := internal.Bar(len(files), "Processing ...")
	indexNewFiles := 0
	index, err := index.NewIndex(
		indexFileReader,
		index.WithID(idGenerator),
		index.WithFiles(files),
		index.WithPreview(internal.GeneratePreview),
		index.WithWorkers(runtime.NumCPU()),
		// index.WithWorkers(6),
		index.WithContext(ctx),
		index.WithProgress(func() { indexProgressBar.Add(1) }),
		index.WithNewFiles(func() { indexNewFiles++ }))

	if err != nil {
		log.Fatal(err.Error())
	}

	if indexNewFiles != 0 {
		slog.Info("New files found", slog.Int("files", indexNewFiles))
	}

	previewFilesCount, previewsSize := index.FilesWithPreviewStat()
	slog.Info("Indexing done")
	slog.Info(
		"Preview stat",
		slog.Int("files", previewFilesCount),
		slog.String("size", bytesutil.PrettyByteSize(previewsSize)),
	)

	if index.OutDated() {
		indexFile.Truncate(0)
		indexFile.Seek(0, 0)
		count, err := index.Encode(indexFile)
		if err != nil {
			panic(err)
		}
		slog.Info("Index file saved", slog.String("size", bytesutil.PrettyByteSize(count)))
	}

	_ = internal.NewServer(ctx)
	slog.Info("Server started")
	<-ctx.Done()
	slog.Info("Successful shutdown")
}

func idGenerator(p index.FileMeta) (string, error) {
	idSource := []byte(fmt.Sprintf("%s%s", p.RelativePath(), p.ModTime()))
	if id, err := internal.SHA256Hash(bytes.NewReader(idSource)); err != nil {
		return id, err
	} else {
		return id[:10], nil
	}
}

func gracefulShutdownCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-done
		slog.Warn("A shutdown request has been received!")
		cancel()
	}()
	return ctx
}
