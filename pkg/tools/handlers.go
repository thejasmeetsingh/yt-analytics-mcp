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
	videoMetrics = []string{
		"estimatedMinutesWatched",            // Total minutes watched during the period.
		"averageViewDuration",                // Average seconds a viewer watched per session.
		"averageViewPercentage",              // Percentage of the video watched on average.
		"shares",                             // Number of times the video was shared.
		"annotationClickThroughRate",         // CTR on interactive video annotations.
		"annotationClickableImpressions",     // Total impressions on clickable annotations.
		"videoThumbnailImpressions",          // Number of times the thumbnail was shown.
		"videoThumbnailImpressionsClickRate", // Click-through rate on the thumbnail.
	}

	channelMetrics = []string{
		"views",                   // Total number of video views.
		"estimatedMinutesWatched", // Total watch time across all videos.
		"likes",                   // Total likes received.
		"dislikes",                // Total dislikes received.
		"comments",                // Total comments posted.
		"shares",                  // Total shares across videos.
		"subscribersGained",       // New subscribers added in the period.
		"subscribersLost",         // Subscribers who unsubscribed in the period.
		"averageViewDuration",     // Average seconds watched per view.
		"averageViewPercentage",   // Average percentage of videos watched.
	}
)

// ListChannelsHandler returns a Markdown summary of all YouTube channels owned by the user.
func ListChannelsHandler(ctx context.Context, req *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	const cacheKey = "channels_list"

	// Return cached result if available — avoids an unnecessary API round-trip.
	if cached, ok := services.Cache.Get(cacheKey); ok {
		return nil, MarkdownOutput{Content: cached}, nil
	}

	// Fetch channel metadata and statistics for the authenticated user's channels.
	resp, err := services.YoutubeService.Channels.List([]string{"snippet", "statistics"}).Mine(true).Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	var b strings.Builder
	b.WriteString("# Your YouTube Channels\n\n")

	for i, ch := range resp.Items {
		// Write a numbered section for each channel with key stats.
		b.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, ch.Snippet.Title))
		b.WriteString(fmt.Sprintf("- **Channel ID**: `%s`\n", ch.Id))
		b.WriteString(fmt.Sprintf("- **Subscribers**: %s\n", formatNumber(ch.Statistics.SubscriberCount)))
		b.WriteString(fmt.Sprintf("- **Total Videos**: %s\n", formatNumber(ch.Statistics.VideoCount)))
		b.WriteString(fmt.Sprintf("- **Total Views**: %s\n\n", formatNumber(ch.Statistics.ViewCount)))
	}

	out := b.String()
	services.Cache.Set(cacheKey, out)
	return nil, MarkdownOutput{Content: out}, nil
}

// ChannelAnalyticsHandler returns aggregated analytics for a given channel over
// a specified date range (defaults to the last 30 days if not provided).
func ChannelAnalyticsHandler(ctx context.Context, req *mcp.CallToolRequest, input ChannelAnalyticsInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	// Resolve dates, falling back to sensible defaults when not supplied.
	startDate := getDateOrDefault(input.StartDate, -30) // Default: 30 days ago.
	endDate := getDateOrDefault(input.EndDate, 0)       // Default: today.

	cacheKey := fmt.Sprintf("analytics_%s_%s_%s", input.ChannelID, startDate, endDate)

	// Honour the cache unless the caller explicitly requests a fresh fetch.
	if !input.ForceRefresh {
		if cached, ok := services.Cache.Get(cacheKey); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	metrics := strings.Join(channelMetrics, ",")

	// Query the Analytics API for daily-dimension data across all channel metrics.
	resp, err := services.AnalyticsService.Reports.Query().
		Ids("channel==" + input.ChannelID).
		StartDate(startDate).
		EndDate(endDate).
		Metrics(metrics).
		Dimensions("day").
		Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	// Aggregate raw rows into a flat map of named metric totals/averages.
	data := formatAnalyticsData(resp.Rows)

	// Build the Markdown summary using pre-computed aggregate values.
	out := fmt.Sprintf(
		"# Channel Analytics (%s to %s)\n\n"+
			"- **Views**: %s\n- **Watch Time**: %.1f hours\n- **Likes**: %s\n"+
			"- **Comments**: %s\n- **Shares**: %s\n- **Subs Gained**: %s\n"+
			"- **Subs Lost**: %s\n- **Net Subs**: %s\n- **Avg Duration**: %.0fs\n"+
			"- **Avg View %%**: %.1f%%\n- **Engagement Rate**: %.2f%%\n",
		startDate, endDate,
		formatNumber(uint64(data["views"])),
		data["watchTime"]/60, // Convert minutes to hours.
		formatNumber(uint64(data["likes"])),
		formatNumber(uint64(data["comments"])),
		formatNumber(uint64(data["shares"])),
		formatNumber(uint64(data["subsGained"])),
		formatNumber(uint64(data["subsLost"])),
		formatNumber(uint64(data["subsGained"]-data["subsLost"])), // Net subscriber change.
		data["avgDuration"],
		data["avgPercent"],
		data["engagement"],
	)

	services.Cache.Set(cacheKey, out)
	return nil, MarkdownOutput{Content: out}, nil
}

// VideoListHandler lists the most recent videos uploaded to the specified channel.
// maxResults defaults to 50 when not provided. The list is fetched via the channel's
// uploads playlist for efficiency.
func VideoListHandler(ctx context.Context, req *mcp.CallToolRequest, input VideoListInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	maxResults := input.MaxResults
	if maxResults == 0 {
		maxResults = 50 // Sensible default — keeps API costs low while still useful.
	}

	cacheKey := fmt.Sprintf("videos_%s_%d", input.ChannelID, maxResults)
	if !input.ForceRefresh {
		if cached, ok := services.Cache.Get(cacheKey); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	// Resolve the uploads playlist ID for the channel.
	chResp, err := services.YoutubeService.Channels.List([]string{"contentDetails"}).Id(input.ChannelID).Do()
	if err != nil || len(chResp.Items) == 0 {
		return nil, MarkdownOutput{}, fmt.Errorf("channel not found")
	}

	// The uploads playlist contains all publicly published videos for the channel.
	uploadsPlaylistID := chResp.Items[0].ContentDetails.RelatedPlaylists.Uploads

	resp, err := services.YoutubeService.PlaylistItems.List([]string{"snippet", "contentDetails"}).
		PlaylistId(uploadsPlaylistID).
		MaxResults(maxResults).
		Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Videos (%d)\n\n", len(resp.Items)))

	for i, item := range resp.Items {
		b.WriteString(fmt.Sprintf(
			"## %d. %s\n- **ID**: `%s`\n- **Description**: %s\n- **Published**: %s\n\n",
			i+1,
			item.Snippet.Title,
			item.ContentDetails.VideoId,
			item.Snippet.Description,
			item.ContentDetails.VideoPublishedAt,
		))
	}

	out := b.String()
	services.Cache.Set(cacheKey, out)
	return nil, MarkdownOutput{Content: out}, nil
}

// VideoAnalyticsHandler returns public statistics (views, likes, comments) for one
// or more videos identified by a comma-separated list of video IDs.
func VideoAnalyticsHandler(ctx context.Context, req *mcp.CallToolRequest, input VideoAnalyticsInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	cacheKey := fmt.Sprintf("video_analytics_%s", input.VideoIDs)
	if !input.ForceRefresh {
		if cached, ok := services.Cache.Get(cacheKey); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	// Parse and trim the comma-separated video ID list supplied by the caller.
	ids := strings.Split(input.VideoIDs, ",")
	for i := range ids {
		ids[i] = strings.TrimSpace(ids[i])
	}

	// Batch-fetch snippet and statistics for all video IDs in a single API call.
	videoResp, err := services.YoutubeService.Videos.List([]string{"snippet", "statistics", "contentDetails"}).
		Id(strings.Join(ids, ",")).
		Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	var b strings.Builder
	b.WriteString("# Video Analytics\n\n")

	for _, v := range videoResp.Items {
		b.WriteString(fmt.Sprintf(
			"## %s\n\n- **ID**: `%s`\n- **Views**: %s\n- **Likes**: %s\n- **Comments**: %s\n- **Duration**: %s\n\n",
			v.Snippet.Title,
			v.Id,
			formatNumber(v.Statistics.ViewCount),
			formatNumber(v.Statistics.LikeCount),
			formatNumber(v.Statistics.CommentCount),
			v.ContentDetails.Duration,
		))
	}

	out := b.String()
	services.Cache.Set(cacheKey, out)
	return nil, MarkdownOutput{Content: out}, nil
}

// CompareChannelPeriodsHandler compares channel-level analytics between two arbitrary
// date ranges and surfaces percentage changes with directional arrows. It also generates
// brief narrative insights based on threshold-driven rules for views, engagement, and
// audience retention.
func CompareChannelPeriodsHandler(ctx context.Context, req *mcp.CallToolRequest, input ComparePeriodsInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	channelID := input.ChannelID

	// Resolve both period boundaries, defaulting each end to today and start to 30 days ago.
	p1Start := getDateOrDefault(input.Period1Start, -30)
	p1End := getDateOrDefault(input.Period1End, 0)
	p2Start := getDateOrDefault(input.Period2Start, -30)
	p2End := getDateOrDefault(input.Period2End, 0)

	cacheKey := fmt.Sprintf("compare_periods_%s_%s_%s_%s_%s", channelID, p1Start, p1End, p2Start, p2End)
	if !input.ForceRefresh {
		if cached, ok := services.Cache.Get(cacheKey); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	metrics := strings.Join(channelMetrics, ",")

	// Fetch analytics for Period 1.
	resp1, err := services.AnalyticsService.Reports.Query().
		Ids("channel==" + channelID).
		StartDate(p1Start).
		EndDate(p1End).
		Metrics(metrics).
		Dimensions("day").
		Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	// Fetch analytics for Period 2.
	resp2, err := services.AnalyticsService.Reports.Query().
		Ids("channel==" + channelID).
		StartDate(p2Start).
		EndDate(p2End).
		Metrics(metrics).
		Dimensions("day").
		Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	p1Data := formatAnalyticsData(resp1.Rows)
	p2Data := formatAnalyticsData(resp2.Rows)

	// calcPercentChange computes the percentage change from oldVal to newVal and
	// returns a directional arrow emoji. Returns (0, "N/A") when oldVal is zero to
	// avoid a divide-by-zero error.
	calcPercentChange := func(oldVal, newVal float64) (float64, string) {
		if oldVal == 0 {
			return 0, "N/A"
		}
		pct := ((newVal - oldVal) / oldVal) * 100
		arrow := "→" // Neutral — negligible change.
		if pct > 0 {
			arrow = "↗️" // Positive trend.
		} else if pct < 0 {
			arrow = "↘️" // Negative trend.
		}
		return pct, arrow
	}

	var b strings.Builder
	b.WriteString("# Channel Period Comparison\n\n")
	b.WriteString(fmt.Sprintf("**Period 1**: %s to %s\n", p1Start, p1End))
	b.WriteString(fmt.Sprintf("**Period 2**: %s to %s\n\n", p2Start, p2End))

	b.WriteString("## Performance Comparison\n\n")
	b.WriteString("| Metric | Period 1 | Period 2 | Change | % Change |\n")
	b.WriteString("|--------|----------|----------|--------|----------|\n")

	// writeRow formats and appends a single comparison table row to b.
	// Set isTime=true to render values as hours; set isPercent=true for % suffix.
	writeRow := func(label string, val1, val2 float64, isTime, isPercent bool) {
		var s1, s2, delta string

		switch {
		case isTime:
			s1 = fmt.Sprintf("%.1f hrs", val1/60)
			s2 = fmt.Sprintf("%.1f hrs", val2/60)
			delta = fmt.Sprintf("%+.1f hrs", (val2-val1)/60)
		case isPercent:
			s1 = fmt.Sprintf("%.1f%%", val1)
			s2 = fmt.Sprintf("%.1f%%", val2)
			delta = fmt.Sprintf("%+.1f%%", val2-val1)
		default:
			s1 = formatNumber(uint64(val1))
			s2 = formatNumber(uint64(val2))
			delta = fmt.Sprintf("%+s", formatNumber(uint64(val2-val1)))
		}

		pctChange, arrow := calcPercentChange(val1, val2)
		b.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %.1f%% %s |\n",
			label, s1, s2, delta, pctChange, arrow))
	}

	// Write one row per tracked metric.
	writeRow("Views", p1Data["views"], p2Data["views"], false, false)
	writeRow("Watch Time", p1Data["watchTime"], p2Data["watchTime"], true, false)
	writeRow("Likes", p1Data["likes"], p2Data["likes"], false, false)
	writeRow("Dislikes", p1Data["dislikes"], p2Data["dislikes"], false, false)
	writeRow("Comments", p1Data["comments"], p2Data["comments"], false, false)
	writeRow("Shares", p1Data["shares"], p2Data["shares"], false, false)
	writeRow("Subs Gained", p1Data["subsGained"], p2Data["subsGained"], false, false)
	writeRow("Subs Lost", p1Data["subsLost"], p2Data["subsLost"], false, false)

	// Net subscribers = gained − lost for each period.
	netSubs1 := p1Data["subsGained"] - p1Data["subsLost"]
	netSubs2 := p2Data["subsGained"] - p2Data["subsLost"]
	writeRow("Net Subscribers", netSubs1, netSubs2, false, false)

	writeRow("Avg View Duration", p1Data["avgDuration"], p2Data["avgDuration"], false, false)
	writeRow("Avg View %", p1Data["avgPercent"], p2Data["avgPercent"], false, true)
	writeRow("Engagement Rate", p1Data["engagement"], p2Data["engagement"], false, true)

	// Generate plain-English insights based on configurable thresholds.
	b.WriteString("\n## Key Insights\n\n")
	var insights []string

	// Views: flag significant growth or decline (>10% threshold).
	viewChange, _ := calcPercentChange(p1Data["views"], p2Data["views"])
	switch {
	case viewChange > 10:
		insights = append(insights, fmt.Sprintf("✅ **Strong view growth** of %.1f%% in period 2", viewChange))
	case viewChange < -10:
		insights = append(insights, fmt.Sprintf("⚠️ **Views declined** by %.1f%% - investigate content strategy", -viewChange))
	}

	// Engagement: flag meaningful shifts in interaction rate (>5% threshold).
	engChange, _ := calcPercentChange(p1Data["engagement"], p2Data["engagement"])
	switch {
	case engChange > 5:
		insights = append(insights, fmt.Sprintf("✅ **Engagement improved** by %.1f%% - audience is more interactive", engChange))
	case engChange < -5:
		insights = append(insights, fmt.Sprintf("⚠️ **Engagement dropped** by %.1f%% - content may need adjustment", -engChange))
	}

	// Retention: flag changes in how much of each video viewers watch (>5% threshold).
	retentionChange, _ := calcPercentChange(p1Data["avgPercent"], p2Data["avgPercent"])
	switch {
	case retentionChange > 5:
		insights = append(insights, "✅ **Better audience retention** - viewers watching more of each video")
	case retentionChange < -5:
		insights = append(insights, "⚠️ **Lower retention** - viewers leaving videos earlier")
	}

	// Provide a neutral fallback when no threshold was breached.
	if len(insights) == 0 {
		insights = append(insights, "📊 Performance is relatively stable between periods")
	}

	for _, insight := range insights {
		b.WriteString(fmt.Sprintf("- %s\n", insight))
	}

	out := b.String()
	services.Cache.Set(cacheKey, out)
	return nil, MarkdownOutput{Content: out}, nil
}

// CompareVideosHandler compares performance metrics side-by-side for two or more
// videos supplied as a comma-separated list of video IDs. It fetches both public
// statistics and owner-level analytics (watch time, retention, etc.) and highlights
// the top performer in each key category.
func CompareVideosHandler(ctx context.Context, req *mcp.CallToolRequest, input CompareVideosInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	startDate := getDateOrDefault(input.StartDate, -30)
	endDate := getDateOrDefault(input.EndDate, 0)

	cacheKey := fmt.Sprintf("compare_videos_%s_%s_%s", input.VideoIDs, startDate, endDate)
	if !input.ForceRefresh {
		if cached, ok := services.Cache.Get(cacheKey); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	// Trim whitespace from each individual video ID in the comma-separated list.
	ids := strings.Split(input.VideoIDs, ",")
	for i := range ids {
		ids[i] = strings.TrimSpace(ids[i])
	}

	// Batch-fetch metadata for all requested videos in a single API call.
	videoResp, err := services.YoutubeService.Videos.List([]string{"snippet", "statistics", "contentDetails"}).
		Id(strings.Join(ids, ",")).
		Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	type videoStats struct {
		title                   string
		id                      string
		views                   uint64
		likes                   uint64
		dislikes                uint64
		comments                uint64
		duration                string
		watchTimeMinutes        float64 // Minutes watched during the analysis period.
		avgDurationSecs         float64 // Average seconds watched per view.
		avgViewPercent          float64 // Percentage of the video watched on average.
		shares                  float64
		engagementRate          float64 // (likes + comments + shares) / views * 100.
		annotationCTR           float64 // Annotation click-through rate.
		annotationImpressions   float64 // Clickable annotation impression count.
		thumbnailImpressions    float64 // Thumbnail impression count.
		thumbnailImpressionsCTR float64 // Thumbnail click-through rate.
	}

	var videos []videoStats

	for _, v := range videoResp.Items {
		vs := videoStats{
			title:    v.Snippet.Title,
			id:       v.Id,
			views:    v.Statistics.ViewCount,
			likes:    v.Statistics.LikeCount,
			dislikes: v.Statistics.DislikeCount,
			comments: v.Statistics.CommentCount,
			duration: v.ContentDetails.Duration,
		}

		metrics := strings.Join(videoMetrics, ",")

		// Enrich with owner-level analytics for the specified date range.
		analyticsResp, err := services.AnalyticsService.Reports.Query().
			Ids("channel==MINE").
			StartDate(startDate).
			EndDate(endDate).
			Metrics(metrics).
			Filters("video==" + v.Id).
			Do()

		// Populate analytics fields only when the API call succeeds and returns rows.
		if err == nil && len(analyticsResp.Rows) > 0 {
			row := analyticsResp.Rows[0]
			vs.watchTimeMinutes = row[0].(float64)
			vs.avgDurationSecs = row[1].(float64)
			vs.avgViewPercent = row[2].(float64)
			vs.shares = row[3].(float64)
			vs.annotationCTR = row[4].(float64)
			vs.annotationImpressions = row[5].(float64)
			vs.thumbnailImpressions = row[6].(float64)
			vs.thumbnailImpressionsCTR = row[7].(float64)

			// Engagement rate: normalised interactions as a percentage of total views.
			if vs.views > 0 {
				vs.engagementRate = (float64(vs.likes) + float64(vs.comments) + vs.shares) / float64(vs.views) * 100
			}
		}

		videos = append(videos, vs)
	}

	var b strings.Builder
	b.WriteString("# Video Performance Comparison\n\n")
	b.WriteString(fmt.Sprintf("**Period**: %s to %s\n", startDate, endDate))
	b.WriteString(fmt.Sprintf("**Videos Analyzed**: %d\n\n", len(videos)))

	// Summary table: one row per video with key metrics only.
	b.WriteString("## Performance Overview\n\n")
	b.WriteString("| Video | Views | Watch Time | Avg Duration | Avg View % | Engagement |\n")
	b.WriteString("|-------|-------|------------|--------------|------------|------------|\n")

	for _, vs := range videos {
		title := vs.title
		// Truncate long titles to keep the Markdown table readable.
		if len(title) > 40 {
			title = title[:37] + "..."
		}
		b.WriteString(fmt.Sprintf("| %s | %s | %.1f hrs | %.0fs | %.1f%% | %.2f%% |\n",
			title,
			formatNumber(vs.views),
			vs.watchTimeMinutes/60, // Convert minutes to hours.
			vs.avgDurationSecs,
			vs.avgViewPercent,
			vs.engagementRate,
		))
	}

	// Detailed section: full metrics breakdown per video.
	b.WriteString("\n## Detailed Metrics\n\n")
	for i, vs := range videos {
		b.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, vs.title))
		b.WriteString(fmt.Sprintf("- **Video ID**: `%s`\n", vs.id))
		b.WriteString(fmt.Sprintf("- **Views**: %s\n", formatNumber(vs.views)))
		b.WriteString(fmt.Sprintf("- **Likes**: %s\n", formatNumber(vs.likes)))
		b.WriteString(fmt.Sprintf("- **Dislikes**: %s\n", formatNumber(vs.dislikes)))
		b.WriteString(fmt.Sprintf("- **Comments**: %s\n", formatNumber(vs.comments)))
		b.WriteString(fmt.Sprintf("- **Duration**: `%s`\n", vs.duration))
		b.WriteString(fmt.Sprintf("- **Watch Time**: %.1f hours\n", vs.watchTimeMinutes/60))
		b.WriteString(fmt.Sprintf("- **Avg Duration**: %.0f seconds\n", vs.avgDurationSecs))
		b.WriteString(fmt.Sprintf("- **Avg View %%**: %.1f%%\n", vs.avgViewPercent))
		b.WriteString(fmt.Sprintf("- **Engagement Rate**: %.2f%%\n\n", vs.engagementRate))
		b.WriteString(fmt.Sprintf("- **Annotation Click Through Rate**: %.2f%%\n\n", vs.annotationCTR))
		b.WriteString(fmt.Sprintf("- **Annotation Clickable Impressions**: %.2f%%\n\n", vs.annotationImpressions))
		b.WriteString(fmt.Sprintf("- **Video Thumbnail Impressions**: %.2f%%\n\n", vs.thumbnailImpressions))
		b.WriteString(fmt.Sprintf("- **Video Thumbnail Impressions Click Rate**: %.2f%%\n\n", vs.thumbnailImpressionsCTR))
	}

	// Determine and surface the top performer in three key dimensions.
	if len(videos) > 1 {
		b.WriteString("## Top Performers\n\n")

		// Seed all three winners with the first video to simplify the comparison loop.
		bestViews := videos[0]
		bestEngagement := videos[0]
		bestRetention := videos[0]

		for _, vs := range videos {
			if vs.views > bestViews.views {
				bestViews = vs
			}
			if vs.engagementRate > bestEngagement.engagementRate {
				bestEngagement = vs
			}
			if vs.avgViewPercent > bestRetention.avgViewPercent {
				bestRetention = vs
			}
		}

		b.WriteString(fmt.Sprintf("- **Most Views**: \"%s\" (%s views)\n", bestViews.title, formatNumber(bestViews.views)))
		b.WriteString(fmt.Sprintf("- **Highest Engagement**: \"%s\" (%.2f%%)\n", bestEngagement.title, bestEngagement.engagementRate))
		b.WriteString(fmt.Sprintf("- **Best Retention**: \"%s\" (%.1f%% avg view)\n", bestRetention.title, bestRetention.avgViewPercent))
	}

	out := b.String()
	services.Cache.Set(cacheKey, out)
	return nil, MarkdownOutput{Content: out}, nil
}

// ComparePublishingScheduleHandler analyses which days of the week yield the highest
// average views and engagement for the specified channel.
func ComparePublishingScheduleHandler(ctx context.Context, req *mcp.CallToolRequest, input CompareScheduleInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	channelID := input.ChannelID
	lookbackDays := input.LookbackDays
	maxVideos := input.MaxVideos

	cacheKey := fmt.Sprintf("compare_schedule_%s_%d_%d", channelID, lookbackDays, maxVideos)
	if !input.ForceRefresh {
		if cached, ok := services.Cache.Get(cacheKey); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	// Resolve the uploads playlist for the channel.
	chResp, err := services.YoutubeService.Channels.List([]string{"contentDetails"}).Id(channelID).Do()
	if err != nil || len(chResp.Items) == 0 {
		return nil, MarkdownOutput{}, fmt.Errorf("channel not found")
	}

	uploadsPlaylistID := chResp.Items[0].ContentDetails.RelatedPlaylists.Uploads

	resp, err := services.YoutubeService.PlaylistItems.List([]string{"snippet", "contentDetails"}).
		PlaylistId(uploadsPlaylistID).
		MaxResults(int64(maxVideos)).
		Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	// dayStats accumulates per-day-of-week totals used to compute averages later.
	type dayStats struct {
		count      int
		totalViews uint64
		avgViews   float64
		totalEng   float64
		avgEng     float64
	}

	// Pre-populate all seven days to guarantee every day has an entry in the map.
	statsByDay := map[string]*dayStats{
		"Sunday": {}, "Monday": {}, "Tuesday": {}, "Wednesday": {},
		"Thursday": {}, "Friday": {}, "Saturday": {},
	}

	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -lookbackDays).Format("2006-01-02")

	for _, item := range resp.Items {
		// Parse the publication timestamp to extract the day of the week.
		publishedAt, err := time.Parse(time.RFC3339, item.ContentDetails.VideoPublishedAt)
		if err != nil {
			continue // Skip videos with malformed publish timestamps.
		}

		dayName := publishedAt.Weekday().String()
		videoID := item.ContentDetails.VideoId

		// Fetch public view/like/comment counts for this video.
		videoResp, err := services.YoutubeService.Videos.List([]string{"statistics"}).Id(videoID).Do()
		if err != nil || len(videoResp.Items) == 0 {
			continue
		}

		views := videoResp.Items[0].Statistics.ViewCount
		likes := videoResp.Items[0].Statistics.LikeCount
		comments := videoResp.Items[0].Statistics.CommentCount

		// Fetch share count from the Analytics API (not available via the Data API).
		analyticsResp, _ := services.AnalyticsService.Reports.Query().
			Ids("channel==MINE").
			StartDate(startDate).
			EndDate(endDate).
			Metrics("shares").
			Filters("video==" + videoID).
			Do()

		shares := float64(0)
		if analyticsResp != nil && len(analyticsResp.Rows) > 0 && len(analyticsResp.Rows[0]) > 0 {
			shares = analyticsResp.Rows[0][0].(float64)
		}

		// Compute per-video engagement rate.
		engagementRate := float64(0)
		if views > 0 {
			engagementRate = (float64(likes) + float64(comments) + shares) / float64(views) * 100
		}

		// Accumulate totals into the day-of-week bucket.
		ds := statsByDay[dayName]
		ds.count++
		ds.totalViews += views
		ds.totalEng += engagementRate
	}

	// Convert accumulated totals to per-day averages.
	for _, ds := range statsByDay {
		if ds.count > 0 {
			ds.avgViews = float64(ds.totalViews) / float64(ds.count)
			ds.avgEng = ds.totalEng / float64(ds.count)
		}
	}

	// dayPerf pairs a day name with its computed stats to allow slice sorting.
	type dayPerf struct {
		day   string
		stats *dayStats
	}

	dayList := []dayPerf{
		{"Sunday", statsByDay["Sunday"]},
		{"Monday", statsByDay["Monday"]},
		{"Tuesday", statsByDay["Tuesday"]},
		{"Wednesday", statsByDay["Wednesday"]},
		{"Thursday", statsByDay["Thursday"]},
		{"Friday", statsByDay["Friday"]},
		{"Saturday", statsByDay["Saturday"]},
	}

	// Bubble-sort descending by average views (small N makes this acceptable).
	for i := 0; i < len(dayList)-1; i++ {
		for j := i + 1; j < len(dayList); j++ {
			if dayList[j].stats.avgViews > dayList[i].stats.avgViews {
				dayList[i], dayList[j] = dayList[j], dayList[i]
			}
		}
	}

	var b strings.Builder
	b.WriteString("# Publishing Schedule Analysis\n\n")
	b.WriteString(fmt.Sprintf("**Analysis Period**: Last %d days\n", lookbackDays))
	b.WriteString(fmt.Sprintf("**Videos Analyzed**: %d\n\n", len(resp.Items)))

	b.WriteString("## Performance by Day of Week\n\n")
	b.WriteString("| Day | Videos | Avg Views | Avg Engagement | Recommendation |\n")
	b.WriteString("|-----|--------|-----------|----------------|----------------|\n")

	for i, dp := range dayList {
		if dp.stats.count == 0 {
			continue // Skip days with no upload data.
		}

		// Assign a tier-based emoji and label according to performance ranking.
		var emoji, rec string
		switch {
		case i == 0:
			emoji, rec = "⭐", "Best"
		case i <= 2:
			emoji, rec = "✅", "Good"
		case i <= 4:
			emoji, rec = "📊", "OK"
		default:
			emoji, rec = "⚠️", "Lower"
		}

		b.WriteString(fmt.Sprintf("| %s | %d | %s | %.2f%% | %s %s |\n",
			dp.day,
			dp.stats.count,
			formatNumber(uint64(dp.stats.avgViews)),
			dp.stats.avgEng,
			emoji, rec,
		))
	}

	// Summarise the best, second-best, and worst publishing days.
	b.WriteString("\n## Recommendations\n\n")
	if dayList[0].stats.count > 0 {
		b.WriteString(fmt.Sprintf("- **Best Day**: %s with average %s views per video\n",
			dayList[0].day, formatNumber(uint64(dayList[0].stats.avgViews))))
	}
	if dayList[1].stats.count > 0 {
		b.WriteString(fmt.Sprintf("- **Second Best**: %s with average %s views per video\n",
			dayList[1].day, formatNumber(uint64(dayList[1].stats.avgViews))))
	}
	if dayList[len(dayList)-1].stats.count > 0 {
		b.WriteString(fmt.Sprintf("- **Avoid**: %s shows lowest average performance\n",
			dayList[len(dayList)-1].day))
	}

	out := b.String()
	services.Cache.Set(cacheKey, out)
	return nil, MarkdownOutput{Content: out}, nil
}

// CompareVideoFormatsHandler categorises videos by format keywords found in their
// titles and compares average performance across each format. Videos whose titles
// do not match any keyword are grouped under "other".
//
// This is useful for understanding:-
// whether tutorial-style, vlog, or short-form content performs best on a given channel.
func CompareVideoFormatsHandler(ctx context.Context, req *mcp.CallToolRequest, input CompareFormatsInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	channelID := input.ChannelID
	keywords := input.FormatKeywords
	lookbackDays := input.LookbackDays
	maxVideos := input.MaxVideos

	cacheKey := fmt.Sprintf("compare_formats_%s_%s_%d_%d", channelID, keywords, lookbackDays, maxVideos)
	if !input.ForceRefresh {
		if cached, ok := services.Cache.Get(cacheKey); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	// Normalise keywords to lowercase for case-insensitive title matching.
	formatKeywords := strings.Split(keywords, ",")
	for i := range formatKeywords {
		formatKeywords[i] = strings.TrimSpace(strings.ToLower(formatKeywords[i]))
	}

	// Resolve the uploads playlist for the channel.
	chResp, err := services.YoutubeService.Channels.List([]string{"contentDetails"}).Id(channelID).Do()
	if err != nil || len(chResp.Items) == 0 {
		return nil, MarkdownOutput{}, fmt.Errorf("channel not found")
	}

	uploadsPlaylistID := chResp.Items[0].ContentDetails.RelatedPlaylists.Uploads

	resp, err := services.YoutubeService.PlaylistItems.List([]string{"snippet", "contentDetails"}).
		PlaylistId(uploadsPlaylistID).
		MaxResults(int64(maxVideos)).
		Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	// formatStats accumulates per-format totals across all videos in that category.
	type formatStats struct {
		count      int
		totalViews uint64
		avgViews   float64
		totalWatch float64 // Total minutes watched.
		avgWatch   float64
		totalDur   float64 // Sum of average durations in seconds.
		avgDur     float64
		totalPct   float64 // Sum of average view percentages.
		avgPct     float64
		totalEng   float64 // Sum of per-video engagement rates.
		avgEng     float64
	}

	// Initialise a bucket for each keyword plus a catch-all "other" bucket.
	statsByFormat := make(map[string]*formatStats)
	for _, kw := range formatKeywords {
		statsByFormat[kw] = &formatStats{}
	}
	statsByFormat["other"] = &formatStats{}

	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -lookbackDays).Format("2006-01-02")

	for _, item := range resp.Items {
		title := strings.ToLower(item.Snippet.Title)
		videoID := item.ContentDetails.VideoId

		// Assign the video to the first matching keyword bucket (first-match wins).
		format := "other"
		for _, kw := range formatKeywords {
			if strings.Contains(title, kw) {
				format = kw
				break
			}
		}

		// Fetch public statistics for this video.
		videoResp, err := services.YoutubeService.Videos.List([]string{"statistics"}).Id(videoID).Do()
		if err != nil || len(videoResp.Items) == 0 {
			continue
		}

		views := videoResp.Items[0].Statistics.ViewCount
		likes := videoResp.Items[0].Statistics.LikeCount
		comments := videoResp.Items[0].Statistics.CommentCount

		// Fetch richer watch-time and retention analytics from the Analytics API.
		analyticsResp, _ := services.AnalyticsService.Reports.Query().
			Ids("channel==MINE").
			StartDate(startDate).
			EndDate(endDate).
			Metrics("estimatedMinutesWatched,averageViewDuration,averageViewPercentage,shares").
			Filters("video==" + videoID).
			Do()

		var watchTime, avgDur, avgPct, shares float64
		if analyticsResp != nil && len(analyticsResp.Rows) > 0 && len(analyticsResp.Rows[0]) >= 4 {
			row := analyticsResp.Rows[0]
			watchTime = row[0].(float64)
			avgDur = row[1].(float64)
			avgPct = row[2].(float64)
			shares = row[3].(float64)
		}

		// Compute per-video engagement rate before accumulating into the format bucket.
		engagementRate := float64(0)
		if views > 0 {
			engagementRate = (float64(likes) + float64(comments) + shares) / float64(views) * 100
		}

		fs := statsByFormat[format]
		fs.count++
		fs.totalViews += views
		fs.totalWatch += watchTime
		fs.totalDur += avgDur
		fs.totalPct += avgPct
		fs.totalEng += engagementRate
	}

	// Calculate per-format averages now that all videos have been processed.
	for _, fs := range statsByFormat {
		if fs.count > 0 {
			fs.avgViews = float64(fs.totalViews) / float64(fs.count)
			fs.avgWatch = fs.totalWatch / float64(fs.count)
			fs.avgDur = fs.totalDur / float64(fs.count)
			fs.avgPct = fs.totalPct / float64(fs.count)
			fs.avgEng = fs.totalEng / float64(fs.count)
		}
	}

	// formatPerf enables slice-sorting of formats by average views.
	type formatPerf struct {
		format string
		stats  *formatStats
	}

	// Collect only formats that contain at least one video.
	var formatList []formatPerf
	for format, fs := range statsByFormat {
		if fs.count > 0 {
			formatList = append(formatList, formatPerf{format, fs})
		}
	}

	// Bubble-sort descending by average views.
	for i := 0; i < len(formatList)-1; i++ {
		for j := i + 1; j < len(formatList); j++ {
			if formatList[j].stats.avgViews > formatList[i].stats.avgViews {
				formatList[i], formatList[j] = formatList[j], formatList[i]
			}
		}
	}

	var b strings.Builder
	b.WriteString("# Video Format Comparison\n\n")
	b.WriteString(fmt.Sprintf("**Analysis Period**: Last %d days\n", lookbackDays))
	b.WriteString(fmt.Sprintf("**Videos Analyzed**: %d\n", len(resp.Items)))
	b.WriteString(fmt.Sprintf("**Format Keywords**: %s\n\n", keywords))

	b.WriteString("## Performance by Format\n\n")
	b.WriteString("| Format | Count | Avg Views | Avg Watch Time | Avg Duration | Avg View % | Engagement |\n")
	b.WriteString("|--------|-------|-----------|----------------|--------------|------------|------------|\n")

	for _, fp := range formatList {
		b.WriteString(fmt.Sprintf("| %s | %d | %s | %.1f hrs | %.0fs | %.1f%% | %.2f%% |\n",
			strings.ToTitle(fp.format),
			fp.stats.count,
			formatNumber(uint64(fp.stats.avgViews)),
			fp.stats.avgWatch/60, // Convert minutes to hours.
			fp.stats.avgDur,
			fp.stats.avgPct,
			fp.stats.avgEng,
		))
	}

	// Surface actionable insights based on known performance thresholds.
	b.WriteString("\n## Insights & Recommendations\n\n")
	if len(formatList) > 0 {
		best := formatList[0]
		b.WriteString(fmt.Sprintf("- **Best Performing Format**: \"%s\" with %s avg views\n",
			strings.ToTitle(best.format), formatNumber(uint64(best.stats.avgViews))))

		for _, fp := range formatList {
			// Highlight formats where viewers watch the majority of the video.
			if fp.stats.avgPct > 70 {
				b.WriteString(fmt.Sprintf("- **High Retention**: \"%s\" format keeps %.1f%% of viewers\n",
					strings.ToTitle(fp.format), fp.stats.avgPct))
			}
			// Highlight formats that drive strong audience interaction.
			if fp.stats.avgEng > 8 {
				b.WriteString(fmt.Sprintf("- **High Engagement**: \"%s\" format has %.2f%% engagement rate\n",
					strings.ToTitle(fp.format), fp.stats.avgEng))
			}
		}

		b.WriteString(fmt.Sprintf("\n**Strategy Tip**: Focus on creating more \"%s\" content as it shows the strongest performance.\n",
			strings.ToTitle(best.format)))
	}

	out := b.String()
	services.Cache.Set(cacheKey, out)
	return nil, MarkdownOutput{Content: out}, nil
}

// GetVideoCommentsHandler retrieves up to maxResults top-level comments for a given
// video, including the author, like count, publish time, and reply count.
func GetVideoCommentsHandler(ctx context.Context, req *mcp.CallToolRequest, input VideoCommentsInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	videoID := input.VideoID

	maxResults := input.MaxResults
	if maxResults == 0 {
		maxResults = 50 // Default to 50 comments when not specified.
	}
	if maxResults > 100 {
		maxResults = 100 // Cap at 100 to respect YouTube API quota limits.
	}

	cacheKey := fmt.Sprintf("comments_%s_%d", videoID, maxResults)
	if !input.ForceRefresh {
		if cached, ok := services.Cache.Get(cacheKey); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	// Fetch the video title to include in the Markdown heading.
	videoResp, err := services.YoutubeService.Videos.List([]string{"snippet"}).Id(videoID).Do()
	if err != nil || len(videoResp.Items) == 0 {
		return nil, MarkdownOutput{}, fmt.Errorf("video not found")
	}

	videoTitle := videoResp.Items[0].Snippet.Title

	// Retrieve comment threads in plain-text format for easier downstream parsing.
	commentsResp, err := services.YoutubeService.CommentThreads.List([]string{"snippet"}).
		VideoId(videoID).
		MaxResults(maxResults).
		TextFormat("plainText").
		Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Comments for Video: %s\n\n", videoTitle))
	b.WriteString(fmt.Sprintf("**Video ID**: `%s`\n", videoID))
	b.WriteString(fmt.Sprintf("**Total Comments Retrieved**: %d\n\n", len(commentsResp.Items)))

	if len(commentsResp.Items) == 0 {
		b.WriteString("No comments found for this video.")
		out := b.String()
		services.Cache.Set(cacheKey, out)
		return nil, MarkdownOutput{Content: out}, nil
	}

	b.WriteString("## Comments\n\n")

	for i, thread := range commentsResp.Items {
		topComment := thread.Snippet.TopLevelComment.Snippet

		b.WriteString(fmt.Sprintf("### %d. %s\n", i+1, topComment.AuthorDisplayName))
		b.WriteString(fmt.Sprintf("- **Likes**: %d\n", topComment.LikeCount))
		b.WriteString(fmt.Sprintf("- **Published At**: %s\n", topComment.PublishedAt))
		b.WriteString(fmt.Sprintf("- **Comment**: %s\n", topComment.TextDisplay))

		// Include reply count only when replies exist to reduce visual noise.
		if thread.Snippet.TotalReplyCount > 0 {
			b.WriteString(fmt.Sprintf("- **Replies**: %d\n", thread.Snippet.TotalReplyCount))
		}
		b.WriteString("\n")
	}

	out := b.String()
	services.Cache.Set(cacheKey, out)
	return nil, MarkdownOutput{Content: out}, nil
}
