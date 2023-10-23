package main

import (
	"errors"
	"fmt"
	"log/slog"
	"musrepo/lib"
	"os"
	"os/exec"
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
	download_commands, err := musrepo.Download(*out_path)
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

	if _, err := os.Stat(*out_path); errors.Is(err, os.ErrNotExist) {
		slog.Info(fmt.Sprintf("Creating directory '%s'", *out_path))
		os.MkdirAll(*out_path, os.ModePerm)
	}

	errors := make(chan error)
	downloaded_ids := make(chan int)
	converted_ids := make(chan int)

	exec_command := func(c lib.Command) error {
		command := c.Cmd()
		cmd := exec.Command(command[0], command[1:]...)
		err := cmd.Start()
		if err != nil {
			return err
		}
		err = cmd.Wait()
		if err != nil {
			return err
		}
		return nil
	}

	var downloaded_count atomic.Int64
	download := func(c lib.Command) {
		defer func() {
			downloaded_count.Add(1)
			if downloaded_count.Load() == int64(len(download_commands)) {
				close(downloaded_ids)
			}
		}()
		if err := exec_command(c); err != nil {
			errors <- err
			downloaded_count.Add(1)
			return
		}
		downloaded_ids <- c.TrackId()
	}

	var converted_count atomic.Int64
	convert := func(c lib.Command) {
		defer func() {
			converted_count.Add(1)
			if converted_count.Load() == int64(len(convert_commands)) {
				close(converted_ids)
			}
		}()
		if err := exec_command(c); err != nil {
			errors <- err
			return
		}
		converted_ids <- c.TrackId()
	}

	for _, cmd := range download_commands {
		go download(cmd)
	}

	for {
		err, more := <-errors
		if !more {
			break
		}
		slog.Error(err.Error())
	}

	for {
		loaded, more := <-downloaded_ids
		if !more {
			break
		}
		slog.Info(fmt.Sprintf("Downloaded track %d", loaded))
		for _, cmd := range convert_commands {
			if cmd.TrackId() == loaded {
				go convert(cmd)
			}
		}
	}

	for {
		converted, more := <-converted_ids
		if !more {
			break
		}
		slog.Info(fmt.Sprintf("Converted track from %d", converted))
	}

}
