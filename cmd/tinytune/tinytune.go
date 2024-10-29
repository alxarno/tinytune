package main

import (
	"fmt"
	"log"
	"os"

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
	entries, err := internal.NewCrawlerOS(c.dir).Scan()
	if err != nil {
		panic(err)
	}
	fmt.Println(entries)
}
