package main

import (
	"os"

	"github.com/dills122/kyn/internal/cli"
)

var execute = cli.Execute
var exit = os.Exit

func main() {
	exit(execute())
}
