package tools

func GetTools() []Tool {
	return []Tool{
		{
			Name:        "list_channels",
			Description: "List all YouTube channels accessible with the API key. Use this first to get channel IDs for your accounts.",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{},
				Required:   []string{},
			},
		},
		{
			Name:        "get_channel_details",
			Description: "Get detailed information about a specific YouTube channel including subscriber count, total views, and video count.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"channel_id": {
						Type:        "string",
						Description: "The YouTube channel ID (get this from list_channels tool)",
					},
					"force_refresh": {
						Type:        "boolean",
						Description: "Force refresh data ignoring cache (default: false)",
					},
				},
				Required: []string{"channel_id"},
			},
		},
		{
			Name:        "get_channel_analytics",
			Description: "Get comprehensive analytics data for a channel within a date range including views, watch time, likes, engagement metrics, and traffic sources.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"channel_id": {
						Type:        "string",
						Description: "The YouTube channel ID",
					},
					"start_date": {
						Type:        "string",
						Description: "Start date in YYYY-MM-DD format (default: 30 days ago)",
					},
					"end_date": {
						Type:        "string",
						Description: "End date in YYYY-MM-DD format (default: today)",
					},
					"force_refresh": {
						Type:        "boolean",
						Description: "Force refresh data ignoring cache (default: false)",
					},
				},
				Required: []string{"channel_id"},
			},
		},
		{
			Name:        "get_video_list",
			Description: "Get a list of all videos from a channel with basic information including video IDs, titles, and publish dates.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"channel_id": {
						Type:        "string",
						Description: "The YouTube channel ID",
					},
					"max_results": {
						Type:        "string",
						Description: "Maximum number of videos to return (default: 50, max: 50)",
						Default:     "50",
					},
					"force_refresh": {
						Type:        "boolean",
						Description: "Force refresh data ignoring cache (default: false)",
					},
				},
				Required: []string{"channel_id"},
			},
		},
		{
			Name:        "get_video_analytics",
			Description: "Get detailed analytics for specific videos including views, watch time, likes, comments, average view duration, and audience retention metrics.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"video_ids": {
						Type:        "string",
						Description: "Comma-separated list of video IDs (e.g., 'abc123,def456')",
					},
					"start_date": {
						Type:        "string",
						Description: "Start date in YYYY-MM-DD format (default: 30 days ago)",
					},
					"end_date": {
						Type:        "string",
						Description: "End date in YYYY-MM-DD format (default: today)",
					},
					"force_refresh": {
						Type:        "boolean",
						Description: "Force refresh data ignoring cache (default: false)",
					},
				},
				Required: []string{"video_ids"},
			},
		},
		{
			Name:        "compare_channel_periods",
			Description: "Compare channel performance between two time periods to identify growth trends and changes in key metrics.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"channel_id": {
						Type:        "string",
						Description: "The YouTube channel ID",
					},
					"period1_start": {
						Type:        "string",
						Description: "Start date of first period in YYYY-MM-DD format",
					},
					"period1_end": {
						Type:        "string",
						Description: "End date of first period in YYYY-MM-DD format",
					},
					"period2_start": {
						Type:        "string",
						Description: "Start date of second period in YYYY-MM-DD format",
					},
					"period2_end": {
						Type:        "string",
						Description: "End date of second period in YYYY-MM-DD format",
					},
					"force_refresh": {
						Type:        "boolean",
						Description: "Force refresh data ignoring cache (default: false)",
					},
				},
				Required: []string{"channel_id", "period1_start", "period1_end", "period2_start", "period2_end"},
			},
		},
		{
			Name:        "compare_videos",
			Description: "Compare performance metrics across multiple videos to identify top performers and patterns.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"video_ids": {
						Type:        "string",
						Description: "Comma-separated list of video IDs to compare (e.g., 'abc123,def456,ghi789')",
					},
					"start_date": {
						Type:        "string",
						Description: "Start date for analytics in YYYY-MM-DD format (default: 30 days ago)",
					},
					"end_date": {
						Type:        "string",
						Description: "End date for analytics in YYYY-MM-DD format (default: today)",
					},
					"force_refresh": {
						Type:        "boolean",
						Description: "Force refresh data ignoring cache (default: false)",
					},
				},
				Required: []string{"video_ids"},
			},
		},
		{
			Name:        "compare_publishing_schedule",
			Description: "Analyze which days of the week perform best for video uploads based on historical data.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"channel_id": {
						Type:        "string",
						Description: "The YouTube channel ID",
					},
					"lookback_days": {
						Type:        "string",
						Description: "Number of days to look back for analysis (default: 90)",
						Default:     "90",
					},
					"max_videos": {
						Type:        "string",
						Description: "Maximum number of videos to analyze (default: 50)",
						Default:     "50",
					},
					"force_refresh": {
						Type:        "boolean",
						Description: "Force refresh data ignoring cache (default: false)",
					},
				},
				Required: []string{"channel_id"},
			},
		},
		{
			Name:        "compare_video_formats",
			Description: "Compare performance across different video formats/types by analyzing title patterns to identify which content types resonate best with your audience.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"channel_id": {
						Type:        "string",
						Description: "The YouTube channel ID",
					},
					"format_keywords": {
						Type:        "string",
						Description: "Comma-separated keywords to categorize videos (e.g., 'tutorial,review,tips,shorts'). Videos will be categorized if their title contains these keywords.",
					},
					"lookback_days": {
						Type:        "string",
						Description: "Number of days to look back for analysis (default: 90)",
						Default:     "90",
					},
					"max_videos": {
						Type:        "string",
						Description: "Maximum number of videos to analyze (default: 50)",
						Default:     "50",
					},
					"force_refresh": {
						Type:        "boolean",
						Description: "Force refresh data ignoring cache (default: false)",
					},
				},
				Required: []string{"channel_id", "format_keywords"},
			},
		},
	}
}
