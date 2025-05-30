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
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	if *run {
		if err := repl.Run(wd); err != nil {
			log.Fatalln(err)
		}
	}
	if *script != "" {
		if err := evaluate(wd, *script); err != nil {
			log.Fatalln(err)
		}
	}
}

func evaluate(wd, file string) error {
	loc := filepath.Join(wd, file)
	src, err := os.ReadFile(loc)
	if err != nil {
		return err
	}
	program, err := squeak.ParseString(string(src))
	if err != nil {
		return err
	}
	return squeak.NewInterpreter(filepath.Dir(loc), os.Stdout).Execute(program)
}
