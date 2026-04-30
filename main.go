package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	apiKey := os.Getenv("ALLO_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "mcp-withallo: ALLO_API_KEY environment variable is required")
		os.Exit(1)
	}

	s := server.NewMCPServer(
		"mcp-withallo",
		"0.2.0",
		server.WithToolCapabilities(true),
	)

	registerTools(s, NewAlloClient(apiKey))

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "mcp-withallo: server error: %v\n", err)
		os.Exit(1)
	}
}
