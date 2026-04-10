package main

import (
	"context"
	"flag"
	"log"

	goMCP "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/cmd"
	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/mcp"
)

var (
	generateToken = flag.Bool("token", false, "if set, Generates google auth token")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	if err := cmd.Run(generateToken); err != nil {
		log.Fatalf("Application failed: %v", err)
		return
	}

	if !(*generateToken) {
		// Initalize MCP server
		server, err := mcp.NewMCPServer()
		if err != nil {
			log.Fatal(err)
			return
		}

		// Run server over stdin/stdout
		if err := server.Run(ctx, &goMCP.StdioTransport{}); err != nil {
			log.Fatal(err)
		}
	}
}
