package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/cache"
	ratelimiter "github.com/thejasmeetsingh/yt-analytics-mcp/pkg/rate-limiter"
	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/tools"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"google.golang.org/api/youtubeanalytics/v2"
)

func NewMCPServer(apiKey string) (*MCPServer, error) {
	ctx := context.Background()
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %v", err)
	}
	analyticsService, err := youtubeanalytics.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Analytics service: %v", err)
	}
	return &MCPServer{
		youtubeService:   youtubeService,
		analyticsService: analyticsService,
		cache:            cache.NewCache(5 * time.Minute),
		rateLimiter:      ratelimiter.NewRateLimiter(10, time.Second),
	}, nil
}

func (s *MCPServer) HandleRequest(req MCPRequest) MCPResponse {
	resp := MCPResponse{Jsonrpc: "2.0", ID: req.ID}
	switch req.Method {
	case "initialize":
		resp.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{"tools": map[string]bool{}},
			"serverInfo":      map[string]string{"name": "youtube-analytics-mcp", "version": "1.0.0"},
		}
	case "tools/list":
		resp.Result = map[string]interface{}{"tools": tools.GetTools()}
	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &MCPError{Code: -32602, Message: "Invalid params"}
			return resp
		}
		if !s.rateLimiter.Allow() {
			resp.Error = &MCPError{Code: -32000, Message: "Rate limit exceeded"}
			return resp
		}
		result, err := s.executeTool(params.Name, params.Arguments)
		if err != nil {
			resp.Error = &MCPError{Code: -32000, Message: err.Error()}
			return resp
		}
		resp.Result = map[string]interface{}{
			"content": []map[string]string{{"type": "text", "text": result}},
		}
	default:
		resp.Error = &MCPError{Code: -32601, Message: "Method not found"}
	}
	return resp
}

func (s *MCPServer) executeTool(name string, args map[string]interface{}) (string, error) {
	ctx := context.Background()
	forceRefresh, _ := args["force_refresh"].(bool)

	switch name {
	case "list_channels":
		return s.listChannels(ctx)
	case "get_channel_details":
		channelID, _ := args["channel_id"].(string)
		if channelID == "" {
			return "", fmt.Errorf("channel_id required")
		}
		return s.getChannelDetails(ctx, channelID, forceRefresh)
	case "get_channel_analytics":
		channelID, _ := args["channel_id"].(string)
		if channelID == "" {
			return "", fmt.Errorf("channel_id required")
		}
		startDate := getDateOrDefault(args["start_date"], -30)
		endDate := getDateOrDefault(args["end_date"], 0)
		return s.getChannelAnalytics(ctx, channelID, startDate, endDate, forceRefresh)
	case "get_video_list":
		channelID, _ := args["channel_id"].(string)
		if channelID == "" {
			return "", fmt.Errorf("channel_id required")
		}
		maxResults := int64(50)
		if val, ok := args["max_results"].(string); ok {
			fmt.Sscanf(val, "%d", &maxResults)
		}
		return s.getVideoList(ctx, channelID, maxResults, forceRefresh)
	case "get_video_analytics":
		videoIDs, _ := args["video_ids"].(string)
		if videoIDs == "" {
			return "", fmt.Errorf("video_ids required")
		}
		startDate := getDateOrDefault(args["start_date"], -30)
		endDate := getDateOrDefault(args["end_date"], 0)
		return s.getVideoAnalytics(ctx, videoIDs, startDate, endDate, forceRefresh)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}
