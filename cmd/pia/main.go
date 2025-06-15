package main

import (
	"flag"
	"fmt"
	"github.com/crookdc/pia/cmd/pia/internal/repl"
	"github.com/crookdc/pia/cmd/pia/internal/tui"
	"log"
	"os"
)

var (
	mode = flag.String("mode", "", "")
)

func main() {
	flag.Parse()
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	var run func(string) error
	switch *mode {
	case "repl":
		run = repl.Run
	case "", "tui":
		run = tui.Run
	default:
		panic(fmt.Errorf("unrecognized mode: %s", *mode))
	}
	if err := run(wd); err != nil {
		log.Fatalln(err)
	}
}
