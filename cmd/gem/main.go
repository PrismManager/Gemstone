package main

import (
	"os"

	"github.com/PrismManager/gemstone/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
