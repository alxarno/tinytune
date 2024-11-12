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

const (
	DEBUG_MODE      = "Debug"
	PRODUCTION_MODE = "Production"
)

var (
	Version        = "n/a"
	CommitHash     = "n/a"
	BuildTimestamp = "n/a"
	Mode           = DEBUG_MODE
)

const (
	INDEX_FILE_NAME         = "index.tinytune"
	PROCESSING_CLI_CATEGORY = "Processing:"
	FFMPEG_CLI_CATEGORY     = "FFmpeg:"
	SERVER_CLI_CATEGORY     = "Server:"
)

type config struct {
	dir              string
	videoProcessing  bool
	imageProcessing  bool
	acceleration     bool
	maxNewImageItems int64
	maxNewVideoItems int64
	port             int
}

func main() {
	cli.VersionPrinter = func(cCtx *cli.Context) {
		fmt.Printf(
			"Version=%s\nCommit-Hash=%s\nBuild-Time=%s\n",
			cCtx.App.Version,
			CommitHash,
			BuildTimestamp,
		)
	}
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "print-version",
		Aliases: []string{"v"},
		Usage:   "print only the version",
	}
	c := config{dir: os.Getenv("PWD")}
	app := &cli.App{
		Name:        "TinyTune",
		Usage:       "the tiny media server",
		Version:     Version,
		Copyright:   "(c) github.com/alxarno/tinytune",
		Suggest:     true,
		HideVersion: false,
		UsageText:   "tinytune [data folder path] [global options]",
		Authors: []*cli.Author{
			{
				Name:  "alxarno",
				Email: "alexarnowork@gmail.com",
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "video",
				Value:       true,
				Aliases:     []string{"av"},
				Usage:       "allows the server to process videos, creating previews and retrieving the necessary meta information",
				Destination: &c.videoProcessing,
				Category:    PROCESSING_CLI_CATEGORY,
			},
			&cli.BoolFlag{
				Name:        "image",
				Value:       true,
				Aliases:     []string{"ai"},
				Usage:       "allows the server to process images, creating previews and retrieving the necessary meta information",
				Destination: &c.imageProcessing,
				Category:    PROCESSING_CLI_CATEGORY,
			},
			&cli.Int64Flag{
				Name:        "max-new-image-items",
				Value:       -1,
				Aliases:     []string{"ni"},
				Usage:       "limits the number of new image files to be processed (use if initial processing of files takes a long time)",
				Destination: &c.maxNewImageItems,
				Category:    PROCESSING_CLI_CATEGORY,
			},
			&cli.Int64Flag{
				Name:        "max-new-video-items",
				Value:       -1,
				Aliases:     []string{"nv"},
				Usage:       "limits the number of new video files to be processed (use if initial processing of files takes a long time)",
				Destination: &c.maxNewVideoItems,
				Category:    PROCESSING_CLI_CATEGORY,
			},
			&cli.BoolFlag{
				Name:        "acceleration",
				Value:       true,
				Aliases:     []string{"a"},
				Usage:       "allows to utilize GPU computing power for ffmpeg",
				Destination: &c.acceleration,
				Category:    FFMPEG_CLI_CATEGORY,
			},
			&cli.IntFlag{
				Name:        "port",
				Usage:       "http server port",
				Value:       8080,
				Destination: &c.port,
				Aliases:     []string{"p"},
				Category:    SERVER_CLI_CATEGORY,
				Action: func(ctx *cli.Context, v int) error {
					if v >= 65536 {
						return fmt.Errorf("flag port value %v out of range[0-65535]", v)
					}
					return nil
				},
			},
		},
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
	previewer, err := internal.NewPreviewer(
		internal.WithImagePreview(c.imageProcessing),
		internal.WithVideoPreview(c.videoProcessing),
		internal.WithAcceleration(c.acceleration),
	)
	if err != nil {
		log.Fatal(err)
	}
	indexProgressBar := internal.Bar(len(files), "Processing ...")
	indexNewFiles := 0
	index, err := index.NewIndex(
		indexFileReader,
		index.WithID(idGenerator),
		index.WithFiles(files),
		index.WithPreview(previewer),
		index.WithWorkers(runtime.NumCPU()),
		index.WithContext(ctx),
		index.WithProgress(func() { indexProgressBar.Add(1) }),
		index.WithNewFiles(func() { indexNewFiles++ }),
		index.WithMaxNewImageItems(c.maxNewImageItems),
		index.WithMaxNewVideoItems(c.maxNewVideoItems))

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

	_ = internal.NewServer(
		ctx,
		internal.WithSource(index),
		internal.WithPort(c.port),
		internal.WithDebug(Mode == DEBUG_MODE),
	)
	slog.Info("Server started", slog.Int("port", c.port))
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
