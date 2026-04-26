package main

import (
	"os"

	"kyn/internal/cli"
)

var execute = cli.Execute
var exit = os.Exit

func main() {
	exit(execute())
}
