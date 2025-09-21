package main

import (
	"os"

	"github.com/chan27-2/appconfigguard/pkg/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
