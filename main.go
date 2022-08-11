package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

var (
	version string
)

func main() {
	versionFlag := flag.Bool("version", false, "Print version information")
	flag.Parse()
	if *versionFlag {
		fmt.Println("Version", version)
		return
	}

	err := compile(os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func compile(out io.Writer) error {
	fmt.Fprintln(out, "🌱")
	return nil
}
