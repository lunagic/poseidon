package main

import (
	"context"
	"os"

	"github.com/aaronellington/poseidon/internal/cli"
)

func main() {
	ctx := context.Background()
	if err := cli.Cmd().ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
