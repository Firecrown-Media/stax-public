package main

import (
	"os"

	"github.com/firecrown-media/stax/cmd"
	_ "github.com/firecrown-media/stax/pkg/providers/wpengine"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
