package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/alxarno/tinytune/internal"
	"github.com/alxarno/tinytune/pkg/bytesutil"
	"github.com/alxarno/tinytune/pkg/index"
	"github.com/alxarno/tinytune/pkg/logging"
	"github.com/alxarno/tinytune/pkg/preview"
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
)

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
	rawConfig := internal.DefaultRawConfig()
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
				Value:       rawConfig.Video,
				Aliases:     []string{"av"},
				Usage:       "allows the server to process videos",
				Destination: &rawConfig.Video,
				Category:    ProcessingCLICategory,
			},
			&cli.BoolFlag{
				Name:        "image",
				Value:       rawConfig.Images,
				Aliases:     []string{"ai"},
				Usage:       "allows the server to process images",
				Destination: &rawConfig.Images,
				Category:    ProcessingCLICategory,
			},
			&cli.Int64Flag{
				Name:        "max-new-image-items",
				Value:       rawConfig.MaxImages,
				Aliases:     []string{"ni"},
				Usage:       "limits the number of new image files to be processed",
				Destination: &rawConfig.MaxImages,
				Category:    ProcessingCLICategory,
			},
			&cli.Int64Flag{
				Name:        "max-new-video-items",
				Value:       rawConfig.MaxVideos,
				Aliases:     []string{"nv"},
				Usage:       "limits the number of new video files to be processed",
				Destination: &rawConfig.MaxVideos,
				Category:    ProcessingCLICategory,
			},
			&cli.IntFlag{
				Name:        "parallel",
				Value:       rawConfig.Parallel,
				Aliases:     []string{"pl"},
				Usage:       "simultaneous file processing (!large values increase RAM consumption!)",
				Destination: &rawConfig.Parallel,
				Category:    ProcessingCLICategory,
			},
			&cli.StringFlag{
				Name:        "includes",
				Value:       rawConfig.Includes,
				Aliases:     []string{"i"},
				Usage:       "excludes from selected by --excludes files by regexp",
				Destination: &rawConfig.Includes,
				Category:    ProcessingCLICategory,
			},
			&cli.StringFlag{
				Name:        "excludes",
				Value:       rawConfig.Excludes,
				Aliases:     []string{"e"},
				Usage:       "excludes from media processing by regexp",
				Destination: &rawConfig.Excludes,
				Category:    ProcessingCLICategory,
			},
			&cli.StringFlag{
				Name:        "max-file-size",
				Usage:       "",
				Value:       rawConfig.MaxFileSize,
				Destination: &rawConfig.MaxFileSize,
				Aliases:     []string{"mfs"},
				Category:    ProcessingCLICategory,
			},
			&cli.BoolFlag{
				Name:        "acceleration",
				Value:       rawConfig.Acceleration,
				Aliases:     []string{"a"},
				Usage:       "allows to utilize GPU computing power for ffmpeg",
				Destination: &rawConfig.Acceleration,
				Category:    FFmpegCLICategory,
			},
			&cli.StringFlag{
				Name:        "streaming",
				Usage:       "",
				Value:       rawConfig.Streaming,
				Destination: &rawConfig.Streaming,
				Aliases:     []string{"s"},
				Category:    ServerCLICategory,
			},
			&cli.IntFlag{
				Name:        "port",
				Usage:       "http server port",
				Value:       rawConfig.Port,
				Destination: &rawConfig.Port,
				Aliases:     []string{"p"},
				Category:    ServerCLICategory,
			},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.Args().Len() != 0 {
				rawConfig.Dir = ctx.Args().Get(ctx.Args().Len() - 1)
			}
			config := internal.NewConfig(rawConfig)
			start(config)

			return nil
		},
	}

	internal.PanicError(app.Run(os.Args))
}

func start(config internal.Config) {
	slog.SetDefault(logging.Get())

	ctx := gracefulShutdownCtx()

	slog.Info("TinyTune", slog.String("version", Version))
	config.Print()

	indexFilePath := filepath.Join(config.Dir, IndexFileName)

	files, err := internal.NewCrawlerOS(config.Dir).Scan(indexFilePath)
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

	excludedFromPreview := internal.GetExcludedFiles(
		files,
		config.Process.Includes,
		config.Process.Excludes,
	)

	if len(excludedFromPreview) != 0 {
		slog.Info(fmt.Sprintf("Got %v excluded files from media processing", len(excludedFromPreview)))
	}

	previewer, err := preview.NewPreviewer(
		preview.WithImage(config.Process.Image.Process),
		preview.WithVideo(config.Process.Video.Process),
		preview.WithAcceleration(config.Process.Acceleration),
		preview.WithExcludedFiles(excludedFromPreview),
		preview.WithMaxImages(config.Process.Image.MaxItems),
		preview.WithMaxVideos(config.Process.Video.MaxItems),
		preview.WithMaxFileSize(config.Process.MaxFileSize),
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
		index.WithFiles(files),
		index.WithPreview(previewer.Pull),
		index.WithWorkers(config.Process.Parallel),
		index.WithProgress(progressBarAdd),
	)
	internal.PanicError(err)

	if indexNewFiles != 0 {
		slog.Info("New files found", slog.Int("count", indexNewFiles))
	}

	totalFiles, previewFilesCount, previewsSize := index.FilesWithPreviewStat()

	slog.Info("Indexing done")
	slog.Info(
		"Stat",
		slog.Int("total files", totalFiles),
		slog.Int("files with preview", previewFilesCount),
		slog.String("total preview data size", bytesutil.PrettyByteSize(previewsSize)),
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

	streamingFiles := internal.GetIncludedFiles(files, config.Streaming)
	if len(streamingFiles) != 0 {
		slog.Info(fmt.Sprintf("Got %v files for streaming", len(streamingFiles)))
	}

	_ = internal.NewServer(
		ctx,
		internal.WithSource(&index),
		internal.WithPort(config.Port),
		internal.WithPWD(config.Dir),
		internal.WithDebug(Mode == DebugMode),
		internal.WithStreaming(streamingFiles),
	)

	slog.Info("Server started", slog.Int("port", config.Port), slog.String("mode", Mode))
	<-ctx.Done()
	slog.Info("Successful shutdown")
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
