package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/mcp"
)

func main() {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		log.Fatal("YOUTUBE_API_KEY required")
		return
	}
	server, err := mcp.NewMCPServer(apiKey)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("YouTube Analytics MCP Server started")
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)
	for {
		var req mcp.MCPRequest
		if err := decoder.Decode(&req); err != nil {
			log.Fatal(err)
			break
		}
		resp := server.HandleRequest(req)
		if err := encoder.Encode(resp); err != nil {
			log.Fatal(err)
			break
		}
	}
}
