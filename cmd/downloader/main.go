package main

import (
	"fmt"
	"musrepo/lib"
	"os"
	"strings"

	"log/slog"

	"github.com/akamensky/argparse"
)

func main() {
	parser := argparse.NewParser("Musrepo", lib.USAGE)
	verbose := parser.Flag("v", "verbose", &argparse.Options{
		Help: "Enable verbose output",
	})
	dry_run := parser.Flag("r", "dry-run", &argparse.Options{
		Help: "Only compose and print the commands that would be executed",
	})
	music_path := parser.String("p", "music-path", &argparse.Options{
		Help:    "Path to music.yaml",
		Default: "music.yaml",
	})
	out_path := parser.String("o", "out-path", &argparse.Options{
		Help:     "Path to output directory",
		Required: true,
	})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Println(parser.Usage(err))
		os.Exit(-1)
	}

	if *verbose {
		handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
		slog.SetDefault(slog.New(handler))
	}

	music, err := lib.LoadMusic(*music_path)
	if err != nil {
		slog.Default().Error(err.Error())
	}

	commands, err := lib.Download(*music, *out_path)

	if *dry_run {
		for _, c := range commands {
			fmt.Println("\"" + strings.Join(c, "\", \"") + "\"")
		}
		os.Exit(0)
	}
}
