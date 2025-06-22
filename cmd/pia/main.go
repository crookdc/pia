package main

import (
	"bufio"
	"flag"
	"github.com/crookdc/pia/cmd/pia/internal/tui"
	"log"
	"os"
	"strings"
)

func main() {
	flag.Parse()
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	props := make(map[string]string)
	if len(os.Args) > 1 {
		src, err := os.ReadFile(os.Args[1])
		if err != nil {
			log.Fatalln(err)
		}
		scn := bufio.NewScanner(strings.NewReader(string(src)))
		for scn.Scan() {
			line := scn.Text()
			segments := strings.SplitN(line, "=", 2)
			if len(segments) != 2 {
				continue
			}
			props[segments[0]] = strings.TrimSpace(
				strings.Trim(segments[1], "\""),
			)
		}
	}
	if err := tui.Run(wd, props); err != nil {
		log.Fatalln(err)
	}
}
