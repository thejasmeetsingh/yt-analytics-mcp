package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/auth"
	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/cache"
	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/services"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"google.golang.org/api/youtubeanalytics/v2"
)

func Run(clientSecretPath *string, generateToken *bool) error {
	// Step 1: Validate credentials path
	if err := validateCredentialsPath(clientSecretPath); err != nil {
		return err
	}

	// Step 2: Load OAuth2 configuration
	config, err := auth.GetConfig(*clientSecretPath)
	if err != nil {
		return fmt.Errorf("failed to load OAuth2 config: %w", err)
	}

	// Step 3: Get application root directory
	rootPath, err := auth.GetAppRootDir()
	if err != nil {
		return fmt.Errorf("failed to determine app root directory: %w", err)
	}

	// Step 4: Create authenticated HTTP client
	client, err := auth.GetClient(rootPath, config, *generateToken)
	if err != nil {
		return fmt.Errorf("failed to create authenticated client: %w", err)
	}

	// Step 5: Initialize YouTube API services
	if err := initializeServices(client); err != nil {
		return fmt.Errorf("failed to initialize YouTube services: %w", err)
	}

	return nil
}

// validateCredentialsPath ensures the credentials flag is provided and points to a valid file
func validateCredentialsPath(clientSecretPath *string) error {
	if *clientSecretPath == "" {
		return fmt.Errorf("missing required flag: -credentials must specify path to client_secret.json")
	}

	// Check if file exists
	if _, err := os.Stat(*clientSecretPath); os.IsNotExist(err) {
		return fmt.Errorf("credentials file not found at path: %s", *clientSecretPath)
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
