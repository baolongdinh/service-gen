package main

import (
	"fmt"
	"os"

	"github.com/interlabs/service-gen/generator"
)

func main() {
	if err := generator.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
