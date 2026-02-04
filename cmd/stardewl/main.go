package main

import (
	"fmt"
	"os"
	
	"github.com/submlit21/stardewl-ink/cmd/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}