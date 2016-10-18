package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func pathToName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

func Check(err error) {
	if err != nil {
		log.Fatal("An error occurred: ", err)
	}
}

func printUsage(cli *Cli) func() {
	return func() {
		fmt.Fprint(os.Stdout, "Usage: yolo [OPTIONS] COMMAND\n\nA fancy script runner")
		flag.CommandLine.SetOutput(os.Stdout)
		flag.PrintDefaults()

		help := "\nCommands:\n"

		for _, group := range cli.groups {
			help += fmt.Sprintf("	%-10.10s%s\n", group.name, group.description)
		}

		help += "\nRun 'yolo COMMAND --help' for more information on a command."
		fmt.Fprintf(os.Stdout, "%s\n", help)
	}
}

func main() {
	cli := NewCli()
	flag.Usage = printUsage(cli)
	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	start := time.Now()
	err := cli.Cmd(flag.Args()...)
	Check(err)

	delta := time.Now().Sub(start)
	log.Printf("Took %0.3fs\n", delta.Seconds())
}
