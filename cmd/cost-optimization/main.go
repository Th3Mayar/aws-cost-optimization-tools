package main

import (
	"os"

	"github.com/Th3Mayar/aws-cost-optimization-tools/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args))
}
