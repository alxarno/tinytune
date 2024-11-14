package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/alxarno/tinytune/internal"
	"github.com/alxarno/tinytune/pkg/bytesutil"
	"github.com/alxarno/tinytune/pkg/index"
	"github.com/alxarno/tinytune/pkg/logging"
	"github.com/urfave/cli/v2"
)

const (
	DebugMode      = "Debug"
	ProductionMode = "Production"
)

//nolint:gochecknoglobals //variables used in build time for args passing
var (
	Version        = "n/a"
	CommitHash     = "n/a"
	BuildTimestamp = "n/a"
	Mode           = DebugMode
)

const (
	IndexFileName         = "index.tinytune"
	ProcessingCLICategory = "Processing:"
	FFmpegCLICategory     = "FFmpeg:"
	ServerCLICategory     = "Server:"
	MaxPortNumber         = 65536
	DefaultPort           = 8080
)

type Config struct {
	dir              string
	videoProcessing  bool
	imageProcessing  bool
	acceleration     bool
	maxNewImageItems int64
	maxNewVideoItems int64
	parallel         int
	port             int
	includes         string
	excludes         string
}

func main() {
	cli.VersionPrinter = func(cCtx *cli.Context) {
		slog.Info(
			fmt.Sprintf(
				"Version=%s\nCommit-Hash=%s\nBuild-Time=%s\n",
				cCtx.App.Version,
				CommitHash,
				BuildTimestamp,
			),
		)
	}
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "print-version",
		Aliases: []string{"v"},
		Usage:   "print only the version",
	}
	config := Config{dir: os.Getenv("PWD")}
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
				Usage:       "allows the server to process videos",
				Destination: &config.videoProcessing,
				Category:    ProcessingCLICategory,
			},
			&cli.BoolFlag{
				Name:        "image",
				Value:       true,
				Aliases:     []string{"ai"},
				Usage:       "allows the server to process images",
				Destination: &config.imageProcessing,
				Category:    ProcessingCLICategory,
			},
			&cli.Int64Flag{
				Name:        "max-new-image-items",
				Value:       -1,
				Aliases:     []string{"ni"},
				Usage:       "limits the number of new image files to be processed",
				Destination: &config.maxNewImageItems,
				Category:    ProcessingCLICategory,
			},
			&cli.Int64Flag{
				Name:        "max-new-video-items",
				Value:       -1,
				Aliases:     []string{"nv"},
				Usage:       "limits the number of new video files to be processed",
				Destination: &config.maxNewVideoItems,
				Category:    ProcessingCLICategory,
			},
			&cli.IntFlag{
				Name:        "parallel",
				Value:       runtime.NumCPU(),
				Aliases:     []string{"pl"},
				Usage:       "simultaneous file processing (!large values increase RAM consumption!)",
				Destination: &config.parallel,
				Category:    ProcessingCLICategory,
			},
			&cli.StringFlag{
				Name:        "includes",
				Value:       "",
				Aliases:     []string{"i"},
				Usage:       "excludes from selected by --excludes files by regexp",
				Destination: &config.includes,
				Category:    ProcessingCLICategory,
			},
			&cli.StringFlag{
				Name:        "excludes",
				Value:       "",
				Aliases:     []string{"e"},
				Usage:       "excludes from media processing by regexp",
				Destination: &config.excludes,
				Category:    ProcessingCLICategory,
			},
			&cli.BoolFlag{
				Name:        "acceleration",
				Value:       true,
				Aliases:     []string{"a"},
				Usage:       "allows to utilize GPU computing power for ffmpeg",
				Destination: &config.acceleration,
				Category:    FFmpegCLICategory,
			},
			&cli.IntFlag{
				Name:        "port",
				Usage:       "http server port",
				Value:       DefaultPort,
				Destination: &config.port,
				Aliases:     []string{"p"},
				Category:    ServerCLICategory,
				Action: func(_ *cli.Context, v int) error {
					if v >= MaxPortNumber {
						return fmt.Errorf("flag port value out of range[0-65535]: %v", v) //nolint:err113
					}

					return nil
				},
			},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.Args().Len() != 0 {
				config.dir = ctx.Args().Get(ctx.Args().Len() - 1)
			}
			start(config)

			return nil
		},
	}

	internal.PanicError(app.Run(os.Args))
}

func start(config Config) {
	slog.SetDefault(logging.Get())

	ctx := gracefulShutdownCtx()

	slog.Info(
		"TinyTune",
		slog.String("dir", config.dir),
		slog.String("version", Version),
		slog.Bool("image-processing", config.imageProcessing),
		slog.Bool("video-processing", config.videoProcessing),
		slog.Bool("acceleration", config.acceleration),
		slog.Int64("max-new-images", config.maxNewImageItems),
		slog.Int64("max-new-videos", config.maxNewVideoItems),
	)

	indexFilePath := filepath.Join(config.dir, IndexFileName)

	files, err := internal.NewCrawlerOS(config.dir).Scan(indexFilePath)
	internal.PanicError(err)

	indexFileRights := 0755

	indexFile, err := os.OpenFile(indexFilePath, os.O_RDWR|os.O_CREATE, fs.FileMode(indexFileRights))
	internal.PanicError(err)
	defer indexFile.Close()

	indexFileReader := io.Reader(indexFile)
	fileInfo, err := indexFile.Stat()
	internal.PanicError(err)

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
		internal.WithImagePreview(config.imageProcessing),
		internal.WithVideoPreview(config.videoProcessing),
		internal.WithAcceleration(config.acceleration),
	)
	internal.PanicError(err)

	indexProgressBar := internal.Bar(len(files), "Processing ...")
	progressBarAdd := func() {
		internal.PanicError(indexProgressBar.Add(1))
	}

	indexNewFiles := 0
	index, err := index.NewIndex(
		ctx,
		indexFileReader,
		index.WithID(idGenerator),
		index.WithFiles(files),
		index.WithPreview(previewer),
		index.WithWorkers(config.parallel),
		index.WithProgress(progressBarAdd),
		index.WithNewFiles(func() { indexNewFiles++ }),
		index.WithMaxNewImageItems(config.maxNewImageItems),
		index.WithMaxNewVideoItems(config.maxNewVideoItems))
	internal.PanicError(err)

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
		err = indexFile.Truncate(0)
		internal.PanicError(err)
		_, err = indexFile.Seek(0, 0)
		internal.PanicError(err)
		count, err := index.Encode(indexFile)
		internal.PanicError(err)
		slog.Info("Index file saved", slog.String("size", bytesutil.PrettyByteSize(count)))
	}

	_ = internal.NewServer(
		ctx,
		internal.WithSource(&index),
		internal.WithPort(config.port),
		internal.WithPWD(config.dir),
		internal.WithDebug(Mode == DebugMode),
	)

	slog.Info("Server started", slog.Int("port", config.port), slog.String("mode", Mode))
	<-ctx.Done()
	slog.Info("Successful shutdown")
}

func idGenerator(p index.FileMeta) (string, error) {
	idSource := []byte(fmt.Sprintf("%s%s", p.RelativePath(), p.ModTime()))
	fileID := sha256.Sum256(idSource)

	return hex.EncodeToString(fileID[:5]), nil
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
