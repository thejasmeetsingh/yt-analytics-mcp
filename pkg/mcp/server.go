package mcp

import (
	"log"
	"time"

	goMCP "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/tools"
	"golang.org/x/time/rate"
)

func NewMCPServer() (*goMCP.Server, error) {
	// Create MCP server
	server := goMCP.NewServer(&goMCP.Implementation{
		Name:    "youtube-analytics",
		Version: "1.0.0",
	}, nil)

	// Add basic tools
	goMCP.AddTool(server, &goMCP.Tool{
		Name:        "list_channels",
		Description: "List all YouTube channels accessible with the API key",
	}, tools.ListChannelsHandler)

	goMCP.AddTool(server, &goMCP.Tool{
		Name:        "get_channel_details",
		Description: "Get detailed information about a specific YouTube channel",
	}, tools.GetChannelDetailsHandler)

	goMCP.AddTool(server, &goMCP.Tool{
		Name:        "get_channel_analytics",
		Description: "Get comprehensive analytics data for a channel within a date range",
	}, tools.GetChannelAnalyticsHandler)

	goMCP.AddTool(server, &goMCP.Tool{
		Name:        "get_video_list",
		Description: "Get a list of all videos from a channel with basic information",
	}, tools.GetVideoListHandler)

	goMCP.AddTool(server, &goMCP.Tool{
		Name:        "get_video_analytics",
		Description: "Get detailed analytics for specific videos",
	}, tools.GetVideoAnalyticsHandler)

	// Add comparison tools
	goMCP.AddTool(server, &goMCP.Tool{
		Name:        "compare_channel_periods",
		Description: "Compare channel performance between two time periods",
	}, tools.CompareChannelPeriodsHandler)

	goMCP.AddTool(server, &goMCP.Tool{
		Name:        "compare_videos",
		Description: "Compare performance metrics across multiple videos",
	}, tools.CompareVideosHandler)

	goMCP.AddTool(server, &goMCP.Tool{
		Name:        "compare_publishing_schedule",
		Description: "Analyze which days of the week perform best for video uploads",
	}, tools.ComparePublishingScheduleHandler)

	goMCP.AddTool(server, &goMCP.Tool{
		Name:        "compare_video_formats",
		Description: "Compare performance across different video formats/types by analyzing title patterns",
	}, tools.CompareVideoFormatsHandler)

	// Add Rate Limiter Middleware
	server.AddReceivingMiddleware(RateLimiter(rate.NewLimiter(rate.Every(time.Second/5), 10)))

	log.Println("YouTube Analytics MCP Server started")

	return server, nil
}
