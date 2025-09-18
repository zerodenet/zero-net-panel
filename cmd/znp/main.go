package main

import (
	"context"
	"log"

	"github.com/zero-net-panel/zero-net-panel/cmd/znp/cli"
)

func main() {
	if err := cli.Execute(context.Background()); err != nil {
		log.Fatalf("znp command failed: %v", err)
	}
}
