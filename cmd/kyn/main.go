package main

import (
	"os"

	"kyn/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
