package main

import (
	"os"

	"github.com/saichandankadarla/appconfigguard/pkg/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
