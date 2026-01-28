package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Konboi/cb/internal/cli"
	"github.com/Konboi/cb/internal/service"
)

func main() {
	ctx := context.Background()
	svc, err := service.New(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing service: %v\n", err)
		os.Exit(1)
	}
	c := &cli.CLI{Service: svc, Out: os.Stdout, Err: os.Stderr}
	code := c.Execute(ctx, os.Args[1:])
	if code != 0 {
		// Ensure a newline if the last error didn't end with one
		fmt.Fprintln(os.Stderr)
	}
	os.Exit(code)
}
