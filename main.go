package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	claude := flag.Bool("claude", false, "render the status line from built-in sample JSON instead of stdin")
	flag.Parse()

	var r io.Reader
	if *claude {
		r = strings.NewReader(sampleInput)
	} else {
		r = os.Stdin
	}

	var in StatusInput
	if err := json.NewDecoder(r).Decode(&in); err != nil {
		fmt.Println("status-line: failed to read input")
		return
	}

	fmt.Println(render(in))
}
