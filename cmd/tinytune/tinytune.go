package main

import (
	"bytes"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/alxarno/tinytune/internal"
	"github.com/alxarno/tinytune/pkg/bytesutil"
	"github.com/alxarno/tinytune/pkg/index"
	"github.com/lmittmann/tint"
	"github.com/urfave/cli/v2"
)

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

func init() {
	setLogger()
}

func start(c config) {
	indexFilePath := filepath.Join(c.dir, "index.tinytune")
	files, err := internal.NewCrawlerOS(c.dir).Scan(indexFilePath)
	if err != nil {
		panic(err)
	}
	indexFile, err := os.OpenFile(indexFilePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	fileInfo, err := indexFile.Stat()
	if err != nil {
		panic(err)
	}
	if fileInfo.Size() != 0 {
		slog.Info(
			"Found index file!",
			slog.String("size", bytesutil.PrettyByteSize(fileInfo.Size())),
		)
	}
	slog.Info("Indexing started!")
	indexProgressBar := internal.Bar(len(files), "Processing ...")
	indexNewFiles := 0
	index := index.NewIndex(
		indexFile,
		index.WithID(idGenerator),
		index.WithFiles(files),
		index.WithPreview(internal.GeneratePreview),
		index.WithProgress(func() { indexProgressBar.Add(1) }),
		index.WithNewFiles(func() { indexNewFiles++ }))

	if indexNewFiles != 0 {
		slog.Info("New files found!", slog.Int("files", indexNewFiles))
	}

	previewFilesCount, previewsSize := index.FilesWithPreviewStat()
	slog.Info(
		"Indexing done! Preview stat:",
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
		slog.Info("Index file saved!", slog.String("size", bytesutil.PrettyByteSize(count)))
	}
}

func setLogger() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.DateTime,
		}),
	))
}

func idGenerator(p index.FileMeta) (string, error) {
	idSource := []byte(fmt.Sprintf("%s%s", p.RelativePath(), p.ModTime()))
	if id, err := internal.SHA256Hash(bytes.NewReader(idSource)); err != nil {
		return id, err
	} else {
		return id[:10], nil
	}
}
