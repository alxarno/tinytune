package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/alxarno/tinytune/internal"
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
		log.Printf("Found index file! %d", fileInfo.Size())
	}
	index := internal.NewIndex(
		indexFile,
		internal.WithFiles(files),
		internal.WithPreview(internal.GeneratePreview),
		internal.WithID(func(p internal.FileMeta) (string, error) {
			idSource := []byte(fmt.Sprintf("%s%s", p.RelativePath(), p.ModTime()))
			id, err := internal.SHA256Hash(bytes.NewReader(idSource))
			if err != nil {
				return id, err
			}
			id = id[:10]
			return id, nil
		}))
	if index.OutDated() {
		indexFile.Truncate(0)
		indexFile.Seek(0, 0)
		count, err := index.Encode(indexFile)
		if err != nil {
			panic(err)
		}
		log.Printf("Wrote %d", count)
	}
}
