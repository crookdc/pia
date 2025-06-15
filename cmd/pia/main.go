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
	var runner func(string) error
	switch *mode {
	case "repl":
		runner = repl.Run
	case "", "tui":
		runner = tui.Run
	default:
		panic(fmt.Errorf("unrecognized mode: %s", *mode))
	}
	if err := runner(wd); err != nil {
		log.Fatalln(err)
	}
}
