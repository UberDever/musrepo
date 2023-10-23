package main

import (
	"errors"
	"fmt"
	"log/slog"
	"musrepo/lib"
	"os"

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
	skip_missing := parser.Flag("s", "skip-missing", &argparse.Options{
		Help: "Skip missing tracks",
	})
	music_path := parser.String("p", "music-path", &argparse.Options{
		Help:    "Path to music.yaml",
		Default: "music.yaml",
	})
	full_path := parser.String("f", "full-path", &argparse.Options{
		Help:     "Path to full track directory",
		Required: true,
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
		slog.Error(err.Error())
		os.Exit(-1)
	}

	musrepo := lib.NewMusRepo(music)
	commands, err := musrepo.Convert(*full_path, *out_path)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}

	if *dry_run {
		for _, c := range commands {
			fmt.Println(lib.FormatCommand(c))
		}
		os.Exit(0)
	}

	if _, err := os.Stat(*out_path); errors.Is(err, os.ErrNotExist) {
		slog.Info(fmt.Sprintf("Creating directory '%s'", *out_path))
		os.MkdirAll(*out_path, os.ModePerm)
	}
	for _, c := range commands {
		if _, err := os.Stat(c.In); errors.Is(err, os.ErrNotExist) {
			if *skip_missing {
				slog.Info(fmt.Sprintf("Skipping missing '%s' for '%s'", c.In, c.Out))
				continue
			} else {
				slog.Error(err.Error())
				os.Exit(-1)
			}
		}
	}
}
