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

	lib.SetupLogHandlers(*verbose)

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

	max_results := func() int { // this is fucking ridiculous, but here we are
		stupid := func(x, y int) int {
			if x > y {
				return x
			}
			return y
		}
		max_results := -1
		for _, commands := range grouped_convert_cmd {
			max_results = stupid(max_results, len(commands))
		}
		return max_results
	}

	type result struct {
		message  string
		is_error bool
	}
	results_size := max_results()
	results := make(chan result, results_size)
	slog.Debug("Max channel size", results_size)

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

	convert_cmd := func(cmd lib.Command, wg *sync.WaitGroup) {
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
		if _, err := os.Stat(cmd.Out()); errors.Is(err, os.ErrNotExist) {
			exec_cmd(cmd)
		} else {
			slog.Info(fmt.Sprintf("Skipping %s -> already exist", cmd.Dump()))
		}

		to_convert := grouped_convert_cmd[cmd.TrackId()]
		var wg sync.WaitGroup
		for _, c := range to_convert {
			slog.Debug("Launching", "cmd", c.Dump())
			wg.Add(1)
			go convert_cmd(c, &wg)
		}
		wg.Wait()
	}

	go func() {
		for res := range results {
			if res.is_error {
				slog.Error(res.message)
			} else {
				slog.Info(res.message)
			}
		}
	}()

	for _, cmd := range download_commands {
		slog.Info("Launching", "cmd", cmd.Dump())
		download_cmd(cmd)
	}

	slog.Info("Wrapping up...")
	close(results)
	slog.Info("Work is done")

}
