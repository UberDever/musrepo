package main

import (
	"fmt"
	"log/slog"
	"musrepo/lib"
	"os"
	"sync/atomic"

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
	download_commands, err := musrepo.Download(*full_path)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}

	convert_commands, err := musrepo.Convert(*full_path, *out_path)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}

	if *dry_run {
		for _, c := range download_commands {
			fmt.Println(lib.FormatCommand(c))
		}
		for _, c := range convert_commands {
			fmt.Println(lib.FormatCommand(c))
		}
		os.Exit(0)
	}

	for _, cmd := range download_commands {
		lib.CreateOutDirs(cmd)
	}
	for _, cmd := range convert_commands {
		lib.CreateOutDirs(cmd)
	}

	type result struct {
		message  string
		is_error bool
	}
	results := make(chan result)

	var downloaded_count atomic.Int64
	jobs_count := int64(len(download_commands))
	download := func(load_cmd lib.Command) {
		defer func() {
			downloaded_count.Add(1)
			if downloaded_count.Load() >= jobs_count {
				close(results)
			}
		}()
		if err := lib.ExecCommand(load_cmd); err != nil {
			results <- result{
				message:  err.Error(),
				is_error: true,
			}
			return
		}
		results <- result{
			message:  fmt.Sprintf("Downloaded %s", load_cmd.Dump()),
			is_error: false,
		}

		for _, cmd := range convert_commands {
			if cmd.TrackId() == load_cmd.TrackId() {
				go func() {
					if err := lib.ExecCommand(load_cmd); err != nil {
						results <- result{
							message:  err.Error(),
							is_error: true,
						}
						return
					}
					results <- result{
						message:  fmt.Sprintf("Converted %s", cmd.Dump()),
						is_error: false,
					}
				}()
			}
		}
	}

	for _, cmd := range download_commands {
		go download(cmd)
	}

	for {
		res, more := <-results
		if !more {
			break
		}
		if res.is_error {
			slog.Error(res.message)
		} else {
			slog.Info(res.message)
		}
		fmt.Println()
	}

}
