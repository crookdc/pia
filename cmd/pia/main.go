package main

import (
	"bytes"
	"errors"
	"flag"
	"github.com/crookdc/pia/cmd/pia/internal/repl"
	"github.com/crookdc/pia/squeak"
	"io"
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
	lx, err := squeak.NewLexer(bytes.NewReader(src))
	if err != nil {
		return err
	}
	plx, err := squeak.NewPeekingLexer(lx)
	if err != nil {
		return err
	}
	in := squeak.NewInterpreter(os.Stdout)
	ps := squeak.NewParser(plx)
	for {
		nxt, err := ps.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := in.Execute(nxt); err != nil {
			return err
		}
	}
}
