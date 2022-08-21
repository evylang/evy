//go:build !tinygo

package main

import (
	"flag"
	"fmt"
	"strings"
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

	fmt.Println(evaluate("some program"))
}

func evaluate(program string) string {
	return strings.ToUpper(truncate(program, 20))
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}
