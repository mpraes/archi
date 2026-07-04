package main

import "github.com/mpraes/archi/internal/cli"

var version = "dev"

func main() {
	cli.Execute(version)
}
