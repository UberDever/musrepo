package main

import (
	"errors"
	"fmt"
	"log/slog"
	"musrepo/lib"
	"os"
	"sync"

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
		ReplaceAttr := func(group []string, a slog.Attr) slog.Attr {
			if a.Key == "time" || a.Key == "level" {
				return slog.Attr{}
			}
			return slog.Attr{Key: a.Key, Value: a.Value}
		}

		handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug, ReplaceAttr: ReplaceAttr})
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

	grouped_convert_cmd := make(map[int][]lib.Command)
	for _, cmd := range convert_commands {
		grouped_convert_cmd[cmd.TrackId()] = append(grouped_convert_cmd[cmd.TrackId()], cmd)
	}

	type result struct {
		message  string
		is_error bool
	}
	results := make(chan result)

	exec_cmd := func(cmd lib.Command) {
		var err error
		var std_out []byte
		if err, std_out = lib.ExecCommand(cmd); err != nil {
			results <- result{
				message:  err.Error(),
				is_error: true,
			}
			return
		}
		var out string
		if len(std_out) == 0 {
			out = "Successfully converted " + cmd.Dump()
		} else {
			out = string(std_out)
		}
		results <- result{
			message:  out,
			is_error: false,
		}
	}

	var wg sync.WaitGroup
	convert_cmd := func(cmd lib.Command) {
		defer wg.Done()

		if _, err := os.Stat(cmd.Out()); errors.Is(err, os.ErrNotExist) {
			exec_cmd(cmd)
		} else {
			results <- result{
				message:  fmt.Sprintf("Skipping %s -> already exist", cmd.Dump()),
				is_error: false,
			}
		}
	}

	download_cmd := func(cmd lib.Command) {
		// exec_cmd(cmd)

		to_convert := grouped_convert_cmd[cmd.TrackId()]
		for _, c := range to_convert {
			wg.Add(1)
			slog.Debug("Launch", "cmd", c.Cmd())
			go convert_cmd(c)
		}
		wg.Wait()

		fmt.Println("Wrapping up...")
		close(results)
		fmt.Println("Completed download")
	}

	for _, cmd := range download_commands {
		fmt.Printf("Launching %s\n", cmd.Dump())
		download_cmd(cmd)

		for {
			res, more := <-results
			if !more {
				break
			}
			if res.is_error {
				fmt.Fprintln(os.Stderr, res.message)
			} else {
				fmt.Println(res.message)
			}
		}
		break
	}

}
