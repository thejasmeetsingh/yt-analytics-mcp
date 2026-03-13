package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/services"
)

var (
	VIDEO_METRICS = []string{
		"views",
		"estimatedMinutesWatched",
		"averageViewDuration",
		"averageViewPercentage",
		"likes",
		"comments",
		"shares",
	}

	CHANNEL_METRICS = []string{
		"views",
		"estimatedMinutesWatched",
		"likes",
		"dislikes",
		"comments",
		"shares",
		"subscribersGained",
		"subscribersLost",
		"averageViewDuration",
		"averageViewPercentage",
		"annotationClickThroughRate",
		"annotationClickableImpressions",
		"videoThumbnailImpressions",
		"videoThumbnailImpressionsClickRate",
	}
)

func ListChannelsHandler(ctx context.Context, req *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	key := "channels_list"
	if cached, ok := services.Cache.Get(key); ok {
		return nil, MarkdownOutput{Content: cached}, nil
	}

	resp, err := services.YoutubeService.Channels.List([]string{"snippet", "statistics"}).Mine(true).Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	var md strings.Builder
	md.WriteString("# Your YouTube Channels\n\n")
	for i, ch := range resp.Items {
		md.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, ch.Snippet.Title))
		md.WriteString(fmt.Sprintf("- **Channel ID**: `%s`\n", ch.Id))
		md.WriteString(fmt.Sprintf("- **Subscribers**: %s\n", formatNumber(ch.Statistics.SubscriberCount)))
		md.WriteString(fmt.Sprintf("- **Total Videos**: %s\n", formatNumber(ch.Statistics.VideoCount)))
		md.WriteString(fmt.Sprintf("- **Total Views**: %s\n\n", formatNumber(ch.Statistics.ViewCount)))
	}

	result := md.String()
	services.Cache.Set(key, result)
	return nil, MarkdownOutput{Content: result}, nil
}

func GetChannelAnalyticsHandler(ctx context.Context, req *mcp.CallToolRequest, input ChannelAnalyticsInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	start := getDateOrDefault(input.StartDate, -30)
	end := getDateOrDefault(input.EndDate, 0)

	key := fmt.Sprintf("analytics_%s_%s_%s", input.ChannelID, start, end)
	if !input.ForceRefresh {
		if cached, ok := services.Cache.Get(key); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	metrics := strings.Join(CHANNEL_METRICS, ",")

	resp, err := services.AnalyticsService.Reports.Query().
		Ids("channel==" + input.ChannelID).StartDate(start).EndDate(end).
		Metrics(metrics).Dimensions("day").Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	data := formatAnalyticsData(resp.Rows)

	md := fmt.Sprintf("# Channel Analytics (%s to %s)\n\n"+
		"- **Views**: %s\n- **Watch Time**: %.1f hours\n- **Likes**: %s\n"+
		"- **Comments**: %s\n- **Shares**: %s\n- **Subs Gained**: %s\n"+
		"- **Subs Lost**: %s\n- **Net Subs**: %s\n- **Avg Duration**: %.0fs\n"+
		"- **Avg View %%**: %.1f%%\n- **Engagement Rate**: %.2f%%\n"+
		"- **Annotation Click Through Rate %%**: %.1f%%\n- **Annotation Clickable Impressions**: %.1f%%\n"+
		"- **Video Thumbnail Impressions %%**: %.1f%%\n- **Video Thumbnail Impressions Click Rate**: %.1f%%\n",
		start, end, formatNumber(uint64(data["views"])), float64(data["watchTime"])/60,
		formatNumber(uint64(data["likes"])), formatNumber(uint64(data["comments"])),
		formatNumber(uint64(data["shares"])), formatNumber(uint64(data["subsGained"])),
		formatNumber(uint64(data["subsLost"])), formatNumber(uint64(data["subsGained"]-data["subsLost"])),
		data["avgDuration"], data["avgPercent"], data["engagement"],
		data["clickThroughRate"], data["clickImpressions"], data["thumbImpressions"], data["thumbImpressionCtr"])

	services.Cache.Set(key, md)
	return nil, MarkdownOutput{Content: md}, nil
}

func GetVideoListHandler(ctx context.Context, req *mcp.CallToolRequest, input VideoListInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	maxResults := input.MaxResults
	if maxResults == 0 {
		maxResults = 50
	}

	key := fmt.Sprintf("videos_%s_%d", input.ChannelID, maxResults)
	if !input.ForceRefresh {
		if cached, ok := services.Cache.Get(key); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	chResp, err := services.YoutubeService.Channels.List([]string{"contentDetails"}).Id(input.ChannelID).Do()
	if err != nil || len(chResp.Items) == 0 {
		return nil, MarkdownOutput{}, fmt.Errorf("channel not found")
	}

	playlistID := chResp.Items[0].ContentDetails.RelatedPlaylists.Uploads
	resp, err := services.YoutubeService.PlaylistItems.List([]string{"snippet", "contentDetails"}).
		PlaylistId(playlistID).MaxResults(maxResults).Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	var md strings.Builder
	md.WriteString(fmt.Sprintf("# Videos (%d)\n\n", len(resp.Items)))
	for i, item := range resp.Items {
		md.WriteString(fmt.Sprintf("## %d. %s\n- **ID**: `%s`\n- **Description**: %s\n- **Published**: %s\n\n",
			i+1, item.Snippet.Title, item.ContentDetails.VideoId, item.Snippet.Description, item.ContentDetails.VideoPublishedAt))
	}

	result := md.String()
	services.Cache.Set(key, result)
	return nil, MarkdownOutput{Content: result}, nil
}

func GetVideoAnalyticsHandler(ctx context.Context, req *mcp.CallToolRequest, input VideoAnalyticsInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	start := getDateOrDefault(input.StartDate, -30)
	end := getDateOrDefault(input.EndDate, 0)

	key := fmt.Sprintf("video_analytics_%s_%s_%s", input.VideoIDs, start, end)
	if !input.ForceRefresh {
		if cached, ok := services.Cache.Get(key); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	ids := strings.Split(input.VideoIDs, ",")
	for i := range ids {
		ids[i] = strings.TrimSpace(ids[i])
	}

	videoResp, err := services.YoutubeService.Videos.List([]string{"snippet", "statistics"}).
		Id(strings.Join(ids, ",")).Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	var md strings.Builder
	md.WriteString(fmt.Sprintf("# Video Analytics (%s to %s)\n\n", start, end))
	for _, video := range videoResp.Items {
		md.WriteString(fmt.Sprintf("## %s\n\n- **ID**: `%s`\n- **Views**: %s\n- **Likes**: %s\n- **Comments**: %s\n\n",
			video.Snippet.Title, video.Id, formatNumber(video.Statistics.ViewCount),
			formatNumber(video.Statistics.LikeCount), formatNumber(video.Statistics.CommentCount)))
	}

	result := md.String()
	services.Cache.Set(key, result)
	return nil, MarkdownOutput{Content: result}, nil
}

func CompareChannelPeriodsHandler(ctx context.Context, req *mcp.CallToolRequest, input ComparePeriodsInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	channelID := input.ChannelID

	p1Start := getDateOrDefault(input.Period1Start, -30)
	p1End := getDateOrDefault(input.Period1End, 0)
	p2Start := getDateOrDefault(input.Period2Start, -30)
	p2End := getDateOrDefault(input.Period2End, 0)

	key := fmt.Sprintf("compare_periods_%s_%s_%s_%s_%s", channelID, p1Start, p1End, p2Start, p2End)
	if !input.ForceRefresh {
		if cached, ok := services.Cache.Get(key); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	metrics := strings.Join(CHANNEL_METRICS, ",")

	// Get period 1 data
	resp1, err := services.AnalyticsService.Reports.Query().
		Ids("channel==" + channelID).StartDate(p1Start).EndDate(p1End).
		Metrics(metrics).Dimensions("day").Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	// Get period 2 data
	resp2, err := services.AnalyticsService.Reports.Query().
		Ids("channel==" + channelID).StartDate(p2Start).EndDate(p2End).
		Metrics(metrics).Dimensions("day").Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	p1 := formatAnalyticsData(resp1.Rows)
	p2 := formatAnalyticsData(resp2.Rows)

	calcChange := func(old, new float64) (float64, string) {
		if old == 0 {
			return 0, "N/A"
		}
		change := ((new - old) / old) * 100
		arrow := "→"
		if change > 0 {
			arrow = "↗️"
		} else if change < 0 {
			arrow = "↘️"
		}
		return change, arrow
	}

	var md strings.Builder
	md.WriteString("# Channel Period Comparison\n\n")
	md.WriteString(fmt.Sprintf("**Period 1**: %s to %s\n", p1Start, p1End))
	md.WriteString(fmt.Sprintf("**Period 2**: %s to %s\n\n", p2Start, p2End))

	md.WriteString("## Performance Comparison\n\n")
	md.WriteString("| Metric | Period 1 | Period 2 | Change | % Change |\n")
	md.WriteString("|--------|----------|----------|--------|----------|\n")

	addRow := func(name string, val1, val2 float64, isTime, isPercent bool) {
		var v1Str, v2Str, changeStr string
		if isTime {
			v1Str = fmt.Sprintf("%.1f hrs", val1/60)
			v2Str = fmt.Sprintf("%.1f hrs", val2/60)
			changeStr = fmt.Sprintf("%+.1f hrs", (val2-val1)/60)
		} else if isPercent {
			v1Str = fmt.Sprintf("%.1f%%", val1)
			v2Str = fmt.Sprintf("%.1f%%", val2)
			changeStr = fmt.Sprintf("%+.1f%%", val2-val1)
		} else {
			v1Str = formatNumber(uint64(val1))
			v2Str = formatNumber(uint64(val2))
			changeStr = fmt.Sprintf("%+s", formatNumber(uint64(val2-val1)))
		}
		pctChange, arrow := calcChange(val1, val2)
		md.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %.1f%% %s |\n",
			name, v1Str, v2Str, changeStr, pctChange, arrow))
	}

	addRow("Views", p1["views"], p2["views"], false, false)
	addRow("Watch Time", p1["watchTime"], p2["watchTime"], true, false)
	addRow("Likes", p1["likes"], p2["likes"], false, false)
	addRow("Dislikes", p1["dislikes"], p2["dislikes"], false, false)
	addRow("Comments", p1["comments"], p2["comments"], false, false)
	addRow("Shares", p1["shares"], p2["shares"], false, false)
	addRow("Subs Gained", p1["subsGained"], p2["subsGained"], false, false)
	addRow("Subs Lost", p1["subsLost"], p2["subsLost"], false, false)
	netSubs1 := p1["subsGained"] - p1["subsLost"]
	netSubs2 := p2["subsGained"] - p2["subsLost"]
	addRow("Net Subscribers", netSubs1, netSubs2, false, false)
	addRow("Avg View Duration", p1["avgDuration"], p2["avgDuration"], false, false)
	addRow("Avg View %", p1["avgPercent"], p2["avgPercent"], false, true)
	addRow("Engagement Rate", p1["engagement"], p2["engagement"], false, true)
	addRow("Annotation Click Through Rate", p1["clickThroughRate"], p2["clickThroughRate"], false, false)
	addRow("Annotation Clickable Impressions", p1["clickImpressions"], p2["clickImpressions"], false, false)
	addRow("Video Thumbnail Impressions", p1["thumbImpressions"], p2["thumbImpressions"], false, false)
	addRow("Video Thumbnail Impressions Click Rate", p1["thumbImpressionCtr"], p2["thumbImpressionCtr"], false, false)

	md.WriteString("\n## Key Insights\n\n")
	insights := []string{}

	viewChange, _ := calcChange(p1["views"], p2["views"])
	if viewChange > 10 {
		insights = append(insights, fmt.Sprintf("✅ **Strong view growth** of %.1f%% in period 2", viewChange))
	} else if viewChange < -10 {
		insights = append(insights, fmt.Sprintf("⚠️ **Views declined** by %.1f%% - investigate content strategy", -viewChange))
	}

	engChange, _ := calcChange(p1["engagement"], p2["engagement"])
	if engChange > 5 {
		insights = append(insights, fmt.Sprintf("✅ **Engagement improved** by %.1f%% - audience is more interactive", engChange))
	} else if engChange < -5 {
		insights = append(insights, fmt.Sprintf("⚠️ **Engagement dropped** by %.1f%% - content may need adjustment", -engChange))
	}

	retentionChange, _ := calcChange(p1["avgPercent"], p2["avgPercent"])
	if retentionChange > 5 {
		insights = append(insights, "✅ **Better audience retention** - viewers watching more of each video")
	} else if retentionChange < -5 {
		insights = append(insights, "⚠️ **Lower retention** - viewers leaving videos earlier")
	}

	if len(insights) == 0 {
		insights = append(insights, "📊 Performance is relatively stable between periods")
	}

	for _, insight := range insights {
		md.WriteString(fmt.Sprintf("- %s\n", insight))
	}

	result := md.String()
	services.Cache.Set(key, result)
	return nil, MarkdownOutput{Content: result}, nil
}

func CompareVideosHandler(ctx context.Context, req *mcp.CallToolRequest, input CompareVideosInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	videoIDs := input.VideoIDs

	start := getDateOrDefault(input.StartDate, -30)
	end := getDateOrDefault(input.EndDate, 0)

	key := fmt.Sprintf("compare_videos_%s_%s_%s", videoIDs, start, end)
	if !input.ForceRefresh {
		if cached, ok := services.Cache.Get(key); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	ids := strings.Split(videoIDs, ",")
	for i := range ids {
		ids[i] = strings.TrimSpace(ids[i])
	}

	videoResp, err := services.YoutubeService.Videos.List([]string{"snippet", "statistics", "contentDetails"}).
		Id(strings.Join(ids, ",")).Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	type VideoData struct {
		Title       string
		ID          string
		Views       uint64
		Likes       uint64
		Comments    uint64
		WatchTime   float64
		AvgDuration float64
		AvgPercent  float64
		Shares      float64
		Engagement  float64
	}

	videos := []VideoData{}
	for _, video := range videoResp.Items {
		vd := VideoData{
			Title:    video.Snippet.Title,
			ID:       video.Id,
			Views:    video.Statistics.ViewCount,
			Likes:    video.Statistics.LikeCount,
			Comments: video.Statistics.CommentCount,
		}

		metrics := strings.Join(VIDEO_METRICS, ",")

		analyticsResp, err := services.AnalyticsService.Reports.Query().
			Ids("channel==MINE").StartDate(start).EndDate(end).
			Metrics(metrics).Filters("video==" + video.Id).Do()

		if err == nil && len(analyticsResp.Rows) > 0 {
			row := analyticsResp.Rows[0]
			vd.WatchTime = row[1].(float64)
			vd.AvgDuration = row[2].(float64)
			vd.AvgPercent = row[3].(float64)
			vd.Shares = row[6].(float64)
			if vd.Views > 0 {
				vd.Engagement = (float64(vd.Likes) + float64(vd.Comments) + vd.Shares) / float64(vd.Views) * 100
			}
		}
		videos = append(videos, vd)
	}

	var md strings.Builder
	md.WriteString("# Video Performance Comparison\n\n")
	md.WriteString(fmt.Sprintf("**Period**: %s to %s\n", start, end))
	md.WriteString(fmt.Sprintf("**Videos Analyzed**: %d\n\n", len(videos)))

	md.WriteString("## Performance Overview\n\n")
	md.WriteString("| Video | Views | Watch Time | Avg Duration | Avg View % | Engagement |\n")
	md.WriteString("|-------|-------|------------|--------------|------------|------------|\n")

	for _, v := range videos {
		title := v.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}
		md.WriteString(fmt.Sprintf("| %s | %s | %.1f hrs | %.0fs | %.1f%% | %.2f%% |\n",
			title, formatNumber(v.Views), v.WatchTime/60, v.AvgDuration, v.AvgPercent, v.Engagement))
	}

	md.WriteString("\n## Detailed Metrics\n\n")
	for i, v := range videos {
		md.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, v.Title))
		md.WriteString(fmt.Sprintf("- **Video ID**: `%s`\n", v.ID))
		md.WriteString(fmt.Sprintf("- **Views**: %s\n", formatNumber(v.Views)))
		md.WriteString(fmt.Sprintf("- **Likes**: %s\n", formatNumber(v.Likes)))
		md.WriteString(fmt.Sprintf("- **Comments**: %s\n", formatNumber(v.Comments)))
		md.WriteString(fmt.Sprintf("- **Watch Time**: %.1f hours\n", v.WatchTime/60))
		md.WriteString(fmt.Sprintf("- **Avg Duration**: %.0f seconds\n", v.AvgDuration))
		md.WriteString(fmt.Sprintf("- **Avg View %%**: %.1f%%\n", v.AvgPercent))
		md.WriteString(fmt.Sprintf("- **Engagement Rate**: %.2f%%\n\n", v.Engagement))
	}

	// Find best performers
	if len(videos) > 1 {
		md.WriteString("## Top Performers\n\n")

		maxViews := videos[0]
		maxEngagement := videos[0]
		maxRetention := videos[0]

		for _, v := range videos {
			if v.Views > maxViews.Views {
				maxViews = v
			}
			if v.Engagement > maxEngagement.Engagement {
				maxEngagement = v
			}
			if v.AvgPercent > maxRetention.AvgPercent {
				maxRetention = v
			}
		}

		md.WriteString(fmt.Sprintf("- **Most Views**: \"%s\" (%s views)\n", maxViews.Title, formatNumber(maxViews.Views)))
		md.WriteString(fmt.Sprintf("- **Highest Engagement**: \"%s\" (%.2f%%)\n", maxEngagement.Title, maxEngagement.Engagement))
		md.WriteString(fmt.Sprintf("- **Best Retention**: \"%s\" (%.1f%% avg view)\n", maxRetention.Title, maxRetention.AvgPercent))
	}

	result := md.String()
	services.Cache.Set(key, result)
	return nil, MarkdownOutput{Content: result}, nil
}

func ComparePublishingScheduleHandler(ctx context.Context, req *mcp.CallToolRequest, input CompareScheduleInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	channelID := input.ChannelID
	lookback := input.LookbackDays
	maxVids := input.MaxVideos

	key := fmt.Sprintf("compare_schedule_%s_%d_%d", channelID, lookback, maxVids)
	if !input.ForceRefresh {
		if cached, ok := services.Cache.Get(key); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	// Get channel's uploads
	chResp, err := services.YoutubeService.Channels.List([]string{"contentDetails"}).Id(channelID).Do()
	if err != nil || len(chResp.Items) == 0 {
		return nil, MarkdownOutput{}, fmt.Errorf("channel not found")
	}

	playlistID := chResp.Items[0].ContentDetails.RelatedPlaylists.Uploads
	resp, err := services.YoutubeService.PlaylistItems.List([]string{"snippet", "contentDetails"}).
		PlaylistId(playlistID).MaxResults(int64(maxVids)).Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	type DayStats struct {
		Count      int
		TotalViews uint64
		AvgViews   float64
		TotalEng   float64
		AvgEng     float64
	}

	dayStats := map[string]*DayStats{
		"Sunday": {}, "Monday": {}, "Tuesday": {}, "Wednesday": {},
		"Thursday": {}, "Friday": {}, "Saturday": {},
	}

	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -lookback).Format("2006-01-02")

	for _, item := range resp.Items {
		publishTime, err := time.Parse(time.RFC3339, item.ContentDetails.VideoPublishedAt)
		if err != nil {
			continue
		}

		dayName := publishTime.Weekday().String()
		videoID := item.ContentDetails.VideoId

		// Get video stats
		videoResp, err := services.YoutubeService.Videos.List([]string{"statistics"}).Id(videoID).Do()
		if err != nil || len(videoResp.Items) == 0 {
			continue
		}

		views := videoResp.Items[0].Statistics.ViewCount
		likes := videoResp.Items[0].Statistics.LikeCount
		comments := videoResp.Items[0].Statistics.CommentCount

		// Get analytics
		metrics := "shares"
		analyticsResp, _ := services.AnalyticsService.Reports.Query().
			Ids("channel==MINE").StartDate(startDate).EndDate(endDate).
			Metrics(metrics).Filters("video==" + videoID).Do()

		shares := float64(0)
		if analyticsResp != nil && len(analyticsResp.Rows) > 0 && len(analyticsResp.Rows[0]) > 0 {
			shares = analyticsResp.Rows[0][0].(float64)
		}

		engagement := float64(0)
		if views > 0 {
			engagement = (float64(likes) + float64(comments) + shares) / float64(views) * 100
		}

		stats := dayStats[dayName]
		stats.Count++
		stats.TotalViews += views
		stats.TotalEng += engagement
	}

	// Calculate averages
	for _, stats := range dayStats {
		if stats.Count > 0 {
			stats.AvgViews = float64(stats.TotalViews) / float64(stats.Count)
			stats.AvgEng = stats.TotalEng / float64(stats.Count)
		}
	}

	// Sort days by performance
	type DayPerf struct {
		Day   string
		Stats *DayStats
	}
	dayList := []DayPerf{
		{"Sunday", dayStats["Sunday"]}, {"Monday", dayStats["Monday"]},
		{"Tuesday", dayStats["Tuesday"]}, {"Wednesday", dayStats["Wednesday"]},
		{"Thursday", dayStats["Thursday"]}, {"Friday", dayStats["Friday"]},
		{"Saturday", dayStats["Saturday"]},
	}

	// Sort by average views
	for i := 0; i < len(dayList)-1; i++ {
		for j := i + 1; j < len(dayList); j++ {
			if dayList[j].Stats.AvgViews > dayList[i].Stats.AvgViews {
				dayList[i], dayList[j] = dayList[j], dayList[i]
			}
		}
	}

	var md strings.Builder
	md.WriteString("# Publishing Schedule Analysis\n\n")
	md.WriteString(fmt.Sprintf("**Analysis Period**: Last %d days\n", lookback))
	md.WriteString(fmt.Sprintf("**Videos Analyzed**: %d\n\n", len(resp.Items)))

	md.WriteString("## Performance by Day of Week\n\n")
	md.WriteString("| Day | Videos | Avg Views | Avg Engagement | Recommendation |\n")
	md.WriteString("|-----|--------|-----------|----------------|----------------|\n")

	for i, dp := range dayList {
		if dp.Stats.Count == 0 {
			continue
		}

		emoji := ""
		rec := ""
		if i == 0 {
			emoji = "⭐"
			rec = "Best"
		} else if i <= 2 {
			emoji = "✅"
			rec = "Good"
		} else if i <= 4 {
			emoji = "📊"
			rec = "OK"
		} else {
			emoji = "⚠️"
			rec = "Lower"
		}

		md.WriteString(fmt.Sprintf("| %s | %d | %s | %.2f%% | %s %s |\n",
			dp.Day, dp.Stats.Count, formatNumber(uint64(dp.Stats.AvgViews)),
			dp.Stats.AvgEng, emoji, rec))
	}

	md.WriteString("\n## Recommendations\n\n")
	if dayList[0].Stats.Count > 0 {
		md.WriteString(fmt.Sprintf("- **Best Day**: %s with average %s views per video\n",
			dayList[0].Day, formatNumber(uint64(dayList[0].Stats.AvgViews))))
	}
	if dayList[1].Stats.Count > 0 {
		md.WriteString(fmt.Sprintf("- **Second Best**: %s with average %s views per video\n",
			dayList[1].Day, formatNumber(uint64(dayList[1].Stats.AvgViews))))
	}
	if dayList[len(dayList)-1].Stats.Count > 0 {
		md.WriteString(fmt.Sprintf("- **Avoid**: %s shows lowest average performance\n",
			dayList[len(dayList)-1].Day))
	}

	result := md.String()
	services.Cache.Set(key, result)
	return nil, MarkdownOutput{Content: result}, nil
}

func CompareVideoFormatsHandler(ctx context.Context, req *mcp.CallToolRequest, input CompareFormatsInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	channelID := input.ChannelID
	keywords := input.FormatKeywords
	lookback := input.LookbackDays
	maxVids := input.MaxVideos

	key := fmt.Sprintf("compare_formats_%s_%s_%d_%d", channelID, keywords, lookback, maxVids)
	if !input.ForceRefresh {
		if cached, ok := services.Cache.Get(key); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	formatKeywords := strings.Split(keywords, ",")
	for i := range formatKeywords {
		formatKeywords[i] = strings.TrimSpace(strings.ToLower(formatKeywords[i]))
	}

	// Get channel's uploads
	chResp, err := services.YoutubeService.Channels.List([]string{"contentDetails"}).Id(channelID).Do()
	if err != nil || len(chResp.Items) == 0 {
		return nil, MarkdownOutput{}, fmt.Errorf("channel not found")
	}

	playlistID := chResp.Items[0].ContentDetails.RelatedPlaylists.Uploads
	resp, err := services.YoutubeService.PlaylistItems.List([]string{"snippet", "contentDetails"}).
		PlaylistId(playlistID).MaxResults(int64(maxVids)).Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	type FormatStats struct {
		Count      int
		TotalViews uint64
		AvgViews   float64
		TotalWatch float64
		AvgWatch   float64
		TotalDur   float64
		AvgDur     float64
		TotalPct   float64
		AvgPct     float64
		TotalEng   float64
		AvgEng     float64
	}

	formatStats := make(map[string]*FormatStats)
	for _, kw := range formatKeywords {
		formatStats[kw] = &FormatStats{}
	}
	formatStats["other"] = &FormatStats{}

	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -lookback).Format("2006-01-02")

	for _, item := range resp.Items {
		title := strings.ToLower(item.Snippet.Title)
		videoID := item.ContentDetails.VideoId

		// Categorize video
		format := "other"
		for _, kw := range formatKeywords {
			if strings.Contains(title, kw) {
				format = kw
				break
			}
		}

		// Get video stats
		videoResp, err := services.YoutubeService.Videos.List([]string{"statistics"}).Id(videoID).Do()
		if err != nil || len(videoResp.Items) == 0 {
			continue
		}

		views := videoResp.Items[0].Statistics.ViewCount
		likes := videoResp.Items[0].Statistics.LikeCount
		comments := videoResp.Items[0].Statistics.CommentCount

		// Get analytics
		metrics := "estimatedMinutesWatched,averageViewDuration,averageViewPercentage,shares"
		analyticsResp, _ := services.AnalyticsService.Reports.Query().
			Ids("channel==MINE").StartDate(startDate).EndDate(endDate).
			Metrics(metrics).Filters("video==" + videoID).Do()

		watchTime := float64(0)
		avgDur := float64(0)
		avgPct := float64(0)
		shares := float64(0)

		if analyticsResp != nil && len(analyticsResp.Rows) > 0 && len(analyticsResp.Rows[0]) >= 4 {
			row := analyticsResp.Rows[0]
			watchTime = row[0].(float64)
			avgDur = row[1].(float64)
			avgPct = row[2].(float64)
			shares = row[3].(float64)
		}

		engagement := float64(0)
		if views > 0 {
			engagement = (float64(likes) + float64(comments) + shares) / float64(views) * 100
		}

		stats := formatStats[format]
		stats.Count++
		stats.TotalViews += views
		stats.TotalWatch += watchTime
		stats.TotalDur += avgDur
		stats.TotalPct += avgPct
		stats.TotalEng += engagement
	}

	// Calculate averages
	for _, stats := range formatStats {
		if stats.Count > 0 {
			stats.AvgViews = float64(stats.TotalViews) / float64(stats.Count)
			stats.AvgWatch = stats.TotalWatch / float64(stats.Count)
			stats.AvgDur = stats.TotalDur / float64(stats.Count)
			stats.AvgPct = stats.TotalPct / float64(stats.Count)
			stats.AvgEng = stats.TotalEng / float64(stats.Count)
		}
	}

	// Sort formats by average views
	type FormatPerf struct {
		Format string
		Stats  *FormatStats
	}
	formatList := []FormatPerf{}
	for format, stats := range formatStats {
		if stats.Count > 0 {
			formatList = append(formatList, FormatPerf{format, stats})
		}
	}

	for i := 0; i < len(formatList)-1; i++ {
		for j := i + 1; j < len(formatList); j++ {
			if formatList[j].Stats.AvgViews > formatList[i].Stats.AvgViews {
				formatList[i], formatList[j] = formatList[j], formatList[i]
			}
		}
	}

	var md strings.Builder
	md.WriteString("# Video Format Comparison\n\n")
	md.WriteString(fmt.Sprintf("**Analysis Period**: Last %d days\n", lookback))
	md.WriteString(fmt.Sprintf("**Videos Analyzed**: %d\n", len(resp.Items)))
	md.WriteString(fmt.Sprintf("**Format Keywords**: %s\n\n", keywords))

	md.WriteString("## Performance by Format\n\n")
	md.WriteString("| Format | Count | Avg Views | Avg Watch Time | Avg Duration | Avg View % | Engagement |\n")
	md.WriteString("|--------|-------|-----------|----------------|--------------|------------|------------|\n")

	for _, fp := range formatList {
		md.WriteString(fmt.Sprintf("| %s | %d | %s | %.1f hrs | %.0fs | %.1f%% | %.2f%% |\n",
			strings.ToTitle(fp.Format), fp.Stats.Count, formatNumber(uint64(fp.Stats.AvgViews)),
			fp.Stats.AvgWatch/60, fp.Stats.AvgDur, fp.Stats.AvgPct, fp.Stats.AvgEng))
	}

	md.WriteString("\n## Insights & Recommendations\n\n")
	if len(formatList) > 0 {
		best := formatList[0]
		md.WriteString(fmt.Sprintf("- **Best Performing Format**: \"%s\" with %s avg views\n",
			strings.ToTitle(best.Format), formatNumber(uint64(best.Stats.AvgViews))))

		for _, fp := range formatList {
			if fp.Stats.AvgPct > 70 {
				md.WriteString(fmt.Sprintf("- **High Retention**: \"%s\" format keeps %.1f%% of viewers\n",
					strings.ToTitle(fp.Format), fp.Stats.AvgPct))
			}
			if fp.Stats.AvgEng > 8 {
				md.WriteString(fmt.Sprintf("- **High Engagement**: \"%s\" format has %.2f%% engagement rate\n",
					strings.ToTitle(fp.Format), fp.Stats.AvgEng))
			}
		}

		md.WriteString(fmt.Sprintf("\n**Strategy Tip**: Focus on creating more \"%s\" content as it shows the strongest performance.\n",
			strings.ToTitle(best.Format)))
	}

	result := md.String()
	services.Cache.Set(key, result)
	return nil, MarkdownOutput{Content: result}, nil
}

func GetVideoCommentsHandler(ctx context.Context, req *mcp.CallToolRequest, input VideoCommentsInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	videoID := input.VideoID
	maxResults := input.MaxResults
	if maxResults == 0 {
		maxResults = 50
	}
	if maxResults > 100 {
		maxResults = 100
	}

	key := fmt.Sprintf("comments_%s_%d", videoID, maxResults)
	if !input.ForceRefresh {
		if cached, ok := services.Cache.Get(key); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	// Get video details first
	videoResp, err := services.YoutubeService.Videos.List([]string{"snippet"}).Id(videoID).Do()
	if err != nil || len(videoResp.Items) == 0 {
		return nil, MarkdownOutput{}, fmt.Errorf("video not found")
	}

	videoTitle := videoResp.Items[0].Snippet.Title

	// Get comments for the video
	commentsResp, err := services.YoutubeService.CommentThreads.List([]string{"snippet"}).
		VideoId(videoID).MaxResults(maxResults).
		TextFormat("plainText").
		Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	var md strings.Builder
	md.WriteString(fmt.Sprintf("# Comments for Video: %s\n\n", videoTitle))
	md.WriteString(fmt.Sprintf("**Video ID**: `%s`\n", videoID))
	md.WriteString(fmt.Sprintf("**Total Comments Retrieved**: %d\n\n", len(commentsResp.Items)))

	if len(commentsResp.Items) == 0 {
		md.WriteString("No comments found for this video.")
		result := md.String()
		services.Cache.Set(key, result)
		return nil, MarkdownOutput{Content: result}, nil
	}

	md.WriteString("## Comments\n\n")

	for i, thread := range commentsResp.Items {
		topComment := thread.Snippet.TopLevelComment.Snippet
		md.WriteString(fmt.Sprintf("### %d. %s\n", i+1, topComment.AuthorDisplayName))
		md.WriteString(fmt.Sprintf("- **Likes**: %d\n", topComment.LikeCount))
		md.WriteString(fmt.Sprintf("- **Published At**: %s\n", topComment.PublishedAt))
		md.WriteString(fmt.Sprintf("- **Comment**: %s\n", topComment.TextDisplay))

		// Include reply count if there are replies
		if thread.Snippet.TotalReplyCount > 0 {
			md.WriteString(fmt.Sprintf("- **Replies**: %d\n", thread.Snippet.TotalReplyCount))
		}
		md.WriteString("\n")
	}

	result := md.String()
	services.Cache.Set(key, result)
	return nil, MarkdownOutput{Content: result}, nil
}
