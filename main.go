package main

import (
	"fmt"
	"os"

	"cbz2epub/cmd/cbz2epub"
)

func main() {
	if err := cbz2epub.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
