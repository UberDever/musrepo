package main

import (
	"fmt"
	"log/slog"
	"musrepo/lib"
	"os"

	"github.com/akamensky/argparse"
)

func main() {
	parser := argparse.NewParser("Musrepo", lib.USAGE)
	verbose := parser.Flag("v", "verbose", &argparse.Options{
		Help: "Enables verbose output",
	})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Println(parser.Usage(err))
		os.Exit(-1)
	}

	if verbose != nil && *verbose == true {
		handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
		slog.SetDefault(slog.New(handler))

		slog.Debug("hello", "username", "Mike", "age", 18)
	}
}
