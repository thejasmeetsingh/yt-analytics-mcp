package main

import (
	"context"
	"log"
	"os"
	"time"

	goMCP "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/cache"
	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/mcp"
	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/services"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"google.golang.org/api/youtubeanalytics/v2"
)

func main() {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		log.Fatal("YOUTUBE_API_KEY environment variable is required")
	}

	ctx := context.Background()

	// Initialize YouTube services
	var err error
	services.YoutubeService, err = youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create YouTube service: %v", err)
	}

	services.AnalyticsService, err = youtubeanalytics.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create Analytics service: %v", err)
	}

	// Initialize cache and rate limiter
	services.Cache = cache.NewCache(5 * time.Minute)

	server, err := mcp.NewMCPServer()

	// Run server over stdin/stdout
	if err := server.Run(ctx, &goMCP.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
