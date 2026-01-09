package tools

type EmptyInput struct{}

type ChannelIDInput struct {
	ChannelID    string `json:"channel_id" jsonschema:"The YouTube channel ID"`
	ForceRefresh bool   `json:"force_refresh,omitempty" jsonschema:"Force refresh data ignoring cache (default: false)"`
}

type ChannelAnalyticsInput struct {
	ChannelID    string `json:"channel_id" jsonschema:"The YouTube channel ID"`
	StartDate    string `json:"start_date,omitempty" jsonschema:"Start date in YYYY-MM-DD format (default: 30 days ago)"`
	EndDate      string `json:"end_date,omitempty" jsonschema:"End date in YYYY-MM-DD format (default: today)"`
	ForceRefresh bool   `json:"force_refresh,omitempty" jsonschema:"Force refresh data ignoring cache"`
}

type VideoListInput struct {
	ChannelID    string `json:"channel_id" jsonschema:"The YouTube channel ID"`
	MaxResults   int64  `json:"max_results,omitempty" jsonschema:"Maximum number of videos to return (default: 50 max: 50)"`
	ForceRefresh bool   `json:"force_refresh,omitempty" jsonschema:"Force refresh data ignoring cache"`
}

type VideoAnalyticsInput struct {
	VideoIDs     string `json:"video_ids" jsonschema:"Comma-separated list of video IDs (e.g. abc123 def456)"`
	StartDate    string `json:"start_date,omitempty" jsonschema:"Start date in YYYY-MM-DD format (default: 30 days ago)"`
	EndDate      string `json:"end_date,omitempty" jsonschema:"End date in YYYY-MM-DD format (default: today)"`
	ForceRefresh bool   `json:"force_refresh,omitempty" jsonschema:"Force refresh data ignoring cache"`
}

type ComparePeriodsInput struct {
	ChannelID    string `json:"channel_id" jsonschema:"The YouTube channel ID"`
	Period1Start string `json:"period1_start" jsonschema:"Start date of first period in YYYY-MM-DD format"`
	Period1End   string `json:"period1_end" jsonschema:"End date of first period in YYYY-MM-DD format"`
	Period2Start string `json:"period2_start" jsonschema:"Start date of second period in YYYY-MM-DD format"`
	Period2End   string `json:"period2_end" jsonschema:"End date of second period in YYYY-MM-DD format"`
	ForceRefresh bool   `json:"force_refresh,omitempty" jsonschema:"Force refresh data ignoring cache"`
}

type CompareVideosInput struct {
	VideoIDs     string `json:"video_ids" jsonschema:"Comma-separated list of video IDs to compare"`
	StartDate    string `json:"start_date,omitempty" jsonschema:"Start date in YYYY-MM-DD format (default: 30 days ago)"`
	EndDate      string `json:"end_date,omitempty" jsonschema:"End date in YYYY-MM-DD format (default: today)"`
	ForceRefresh bool   `json:"force_refresh,omitempty" jsonschema:"Force refresh data ignoring cache"`
}

type CompareScheduleInput struct {
	ChannelID    string `json:"channel_id" jsonschema:"The YouTube channel ID"`
	LookbackDays int    `json:"lookback_days,omitempty" jsonschema:"Number of days to look back for analysis (default: 90)"`
	MaxVideos    int    `json:"max_videos,omitempty" jsonschema:"Maximum number of videos to analyze (default: 50)"`
	ForceRefresh bool   `json:"force_refresh,omitempty" jsonschema:"Force refresh data ignoring cache"`
}

type CompareFormatsInput struct {
	ChannelID      string `json:"channel_id" jsonschema:"The YouTube channel ID"`
	FormatKeywords string `json:"format_keywords" jsonschema:"Comma-separated keywords to categorize videos (e.g. tutorial review tips shorts)"`
	LookbackDays   int    `json:"lookback_days,omitempty" jsonschema:"Number of days to look back for analysis (default: 90)"`
	MaxVideos      int    `json:"max_videos,omitempty" jsonschema:"Maximum number of videos to analyze (default: 50)"`
	ForceRefresh   bool   `json:"force_refresh,omitempty" jsonschema:"Force refresh data ignoring cache"`
}

type MarkdownOutput struct {
	Content string `json:"content" jsonschema:"Markdown formatted output"`
}
