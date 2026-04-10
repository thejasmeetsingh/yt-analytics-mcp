package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/auth"
	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/cache"
	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/services"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"google.golang.org/api/youtubeanalytics/v2"
)

func Run(generateToken *bool) error {
	// Step 1: Load OAuth2 configuration from environment variables
	config, err := auth.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load OAuth2 config: %w", err)
	}

	// Step 2: Get application root directory
	rootPath, err := auth.GetAppRootDir()
	if err != nil {
		return fmt.Errorf("failed to determine app root directory: %w", err)
	}

	// Step 3: Create authenticated HTTP client
	client, err := auth.GetClient(rootPath, config, *generateToken)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Step 4: Initialize YouTube API services
	if err := initializeServices(client); err != nil {
		return fmt.Errorf("failed to initialize YouTube services: %w", err)
	}

	return nil
}

func initializeServices(client *http.Client) error {
	ctx := context.Background()

	// Initialize YouTube Data API service
	youtubeService, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("YouTube Data API initialization failed: %w", err)
	}
	services.YoutubeService = youtubeService

	// Initialize YouTube Analytics API service
	analyticsService, err := youtubeanalytics.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("YouTube Analytics API initialization failed: %w", err)
	}
	services.AnalyticsService = analyticsService

	// Initialize In-memory cache service
	services.Cache = cache.NewCache(5 * time.Minute)

	return nil
}
