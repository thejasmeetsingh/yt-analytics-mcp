package services

import (
	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/cache"
	"google.golang.org/api/youtube/v3"
	"google.golang.org/api/youtubeanalytics/v2"
)

var (
	YoutubeService   *youtube.Service
	AnalyticsService *youtubeanalytics.Service
	Cache            *cache.Cache
)
