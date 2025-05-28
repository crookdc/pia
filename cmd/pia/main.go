package main

import (
	"flag"
	"github.com/crookdc/pia/cmd/pia/internal/repl"
	"github.com/crookdc/pia/squeak"
	"log"
	"os"
	"path/filepath"
)

var (
	run    = flag.Bool("run", false, "starts the Squeak REPL in an interactive mode")
	script = flag.String("script", "", "runs the specified Squeak script")
)

func main() {
	flag.Parse()
	if *run {
		if err := repl.Run(); err != nil {
			log.Fatalln(err)
		}
	}
	if *script != "" {
		if err := evaluate(*script); err != nil {
			log.Fatalln(err)
		}
	}
}

func evaluate(file string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	src, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return err
	}
	program, err := squeak.ParseString(string(src))
	if err != nil {
		return err
	}
	return squeak.NewInterpreter(os.Stdout).Execute(program)
}
