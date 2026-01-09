package mcp

import (
	"encoding/json"

	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/cache"
	ratelimiter "github.com/thejasmeetsingh/yt-analytics-mcp/pkg/rate-limiter"
	"google.golang.org/api/youtube/v3"
	"google.golang.org/api/youtubeanalytics/v2"
)

type MCPServer struct {
	youtubeService   *youtube.Service
	analyticsService *youtubeanalytics.Service
	cache            *cache.Cache
	rateLimiter      *ratelimiter.RateLimiter
}

type MCPRequest struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type MCPResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
