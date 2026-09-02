package main

import (
	"flag"
	"os"
	"strings"

	"github.com/tarcisiozf/goom"
	"github.com/tarcisiozf/goom/perf"
)

var (
	enableProfile = flag.Bool("profile", false, "enable CPU profiling")
	loadDump      = flag.Bool("load", false, "load dump")
)

func init() {
	flag.Parse()
}

func main() {
	args := os.Args
	if len(args) == 0 {
		printHelp()
		os.Exit(1)
	}

	wadFile := args[len(args)-1]
	if !strings.HasSuffix(wadFile, ".wad") {
		println("Error: Last argument must be a .wad file")
		printHelp()
		os.Exit(1)
	}

	if *enableProfile {
		stop := perf.StartProfileCPU()
		defer stop()
	}

	game, err := goom.New(wadFile)
	if err != nil {
		println("Error initializing GOOM:", err.Error())
		os.Exit(1)
	}

	if err := game.Run(); err != nil {
		println("Error running GOOM:", err.Error())
		os.Exit(1)
	}
}

func printHelp() {
	println("Usage: goom [options] <wad-file>")
	println("Options:")
	println("  -v, --verbose   Enable verbose logging")
}
