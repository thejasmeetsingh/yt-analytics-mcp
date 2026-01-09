package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/thejasmeetsingh/yt-analytics-mcp/pkg/services"
)

var (
	cache            = services.Cache
	youtubeService   = services.YoutubeService
	analyticsService = services.AnalyticsService
)

func ListChannelsHandler(ctx context.Context, req *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	key := "channels_list"
	if cached, ok := cache.Get(key); ok {
		return nil, MarkdownOutput{Content: cached}, nil
	}

	resp, err := youtubeService.Channels.List([]string{"snippet", "statistics"}).Mine(true).Do()
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
	cache.Set(key, result)
	return nil, MarkdownOutput{Content: result}, nil
}

func GetChannelDetailsHandler(ctx context.Context, req *mcp.CallToolRequest, input ChannelIDInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	key := "ch_" + input.ChannelID
	if !input.ForceRefresh {
		if cached, ok := cache.Get(key); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	resp, err := youtubeService.Channels.List([]string{"snippet", "statistics"}).Id(input.ChannelID).Do()
	if err != nil || len(resp.Items) == 0 {
		return nil, MarkdownOutput{}, fmt.Errorf("channel not found")
	}

	ch := resp.Items[0]
	md := fmt.Sprintf("# %s\n\n## Statistics\n\n- **Subscribers**: %s\n- **Videos**: %s\n- **Views**: %s\n",
		ch.Snippet.Title, formatNumber(ch.Statistics.SubscriberCount),
		formatNumber(ch.Statistics.VideoCount), formatNumber(ch.Statistics.ViewCount))

	cache.Set(key, md)
	return nil, MarkdownOutput{Content: md}, nil
}

func GetChannelAnalyticsHandler(ctx context.Context, req *mcp.CallToolRequest, input ChannelAnalyticsInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	start := getDateOrDefault(input.StartDate, -30)
	end := getDateOrDefault(input.EndDate, 0)

	key := fmt.Sprintf("analytics_%s_%s_%s", input.ChannelID, start, end)
	if !input.ForceRefresh {
		if cached, ok := cache.Get(key); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	metrics := "views,estimatedMinutesWatched,likes,comments,shares,subscribersGained,subscribersLost,averageViewDuration,averageViewPercentage"
	resp, err := analyticsService.Reports.Query().
		Ids("channel==" + input.ChannelID).StartDate(start).EndDate(end).
		Metrics(metrics).Dimensions("day").Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	var views, watchTime, likes, comments, shares, subsGained, subsLost int64
	var avgDuration, avgPercent float64
	for _, row := range resp.Rows {
		if len(row) >= 9 {
			views += int64(row[1].(float64))
			watchTime += int64(row[2].(float64))
			likes += int64(row[3].(float64))
			comments += int64(row[4].(float64))
			shares += int64(row[5].(float64))
			subsGained += int64(row[6].(float64))
			subsLost += int64(row[7].(float64))
			avgDuration += row[8].(float64)
			if len(row) >= 10 {
				avgPercent += row[9].(float64)
			}
		}
	}
	if len(resp.Rows) > 0 {
		avgDuration /= float64(len(resp.Rows))
		avgPercent /= float64(len(resp.Rows))
	}

	md := fmt.Sprintf("# Channel Analytics (%s to %s)\n\n"+
		"- **Views**: %s\n- **Watch Time**: %.1f hours\n- **Likes**: %s\n"+
		"- **Comments**: %s\n- **Shares**: %s\n- **Subs Gained**: %s\n"+
		"- **Subs Lost**: %s\n- **Net Subs**: %s\n- **Avg Duration**: %.0fs\n"+
		"- **Avg View %%**: %.1f%%\n- **Engagement Rate**: %.2f%%\n",
		start, end, formatNumber(uint64(views)), float64(watchTime)/60,
		formatNumber(uint64(likes)), formatNumber(uint64(comments)),
		formatNumber(uint64(shares)), formatNumber(uint64(subsGained)),
		formatNumber(uint64(subsLost)), formatNumber(uint64(subsGained-subsLost)),
		avgDuration, avgPercent, float64(likes+comments+shares)/float64(views)*100)

	cache.Set(key, md)
	return nil, MarkdownOutput{Content: md}, nil
}

func GetVideoListHandler(ctx context.Context, req *mcp.CallToolRequest, input VideoListInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	maxResults := input.MaxResults
	if maxResults == 0 {
		maxResults = 50
	}

	key := fmt.Sprintf("videos_%s_%d", input.ChannelID, maxResults)
	if !input.ForceRefresh {
		if cached, ok := cache.Get(key); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	chResp, err := youtubeService.Channels.List([]string{"contentDetails"}).Id(input.ChannelID).Do()
	if err != nil || len(chResp.Items) == 0 {
		return nil, MarkdownOutput{}, fmt.Errorf("channel not found")
	}

	playlistID := chResp.Items[0].ContentDetails.RelatedPlaylists.Uploads
	resp, err := youtubeService.PlaylistItems.List([]string{"snippet", "contentDetails"}).
		PlaylistId(playlistID).MaxResults(maxResults).Do()
	if err != nil {
		return nil, MarkdownOutput{}, err
	}

	var md strings.Builder
	md.WriteString(fmt.Sprintf("# Videos (%d)\n\n", len(resp.Items)))
	for i, item := range resp.Items {
		md.WriteString(fmt.Sprintf("## %d. %s\n- **ID**: `%s`\n- **Published**: %s\n\n",
			i+1, item.Snippet.Title, item.ContentDetails.VideoId, item.ContentDetails.VideoPublishedAt))
	}

	result := md.String()
	cache.Set(key, result)
	return nil, MarkdownOutput{Content: result}, nil
}

func GetVideoAnalyticsHandler(ctx context.Context, req *mcp.CallToolRequest, input VideoAnalyticsInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	start := getDateOrDefault(input.StartDate, -30)
	end := getDateOrDefault(input.EndDate, 0)

	key := fmt.Sprintf("video_analytics_%s_%s_%s", input.VideoIDs, start, end)
	if !input.ForceRefresh {
		if cached, ok := cache.Get(key); ok {
			return nil, MarkdownOutput{Content: cached}, nil
		}
	}

	ids := strings.Split(input.VideoIDs, ",")
	for i := range ids {
		ids[i] = strings.TrimSpace(ids[i])
	}

	videoResp, err := youtubeService.Videos.List([]string{"snippet", "statistics"}).
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
	cache.Set(key, result)
	return nil, MarkdownOutput{Content: result}, nil
}

func CompareChannelPeriodsHandler(ctx context.Context, req *mcp.CallToolRequest, input ComparePeriodsInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	md := "# Channel Period Comparison\n\n[Comparison functionality - implement full version from original code]"
	return nil, MarkdownOutput{Content: md}, nil
}

func CompareVideosHandler(ctx context.Context, req *mcp.CallToolRequest, input CompareVideosInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	md := "# Video Performance Comparison\n\n[Comparison functionality - implement full version from original code]"
	return nil, MarkdownOutput{Content: md}, nil
}

func ComparePublishingScheduleHandler(ctx context.Context, req *mcp.CallToolRequest, input CompareScheduleInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	md := "# Publishing Schedule Analysis\n\n[Schedule analysis - implement full version from original code]"
	return nil, MarkdownOutput{Content: md}, nil
}

func CompareVideoFormatsHandler(ctx context.Context, req *mcp.CallToolRequest, input CompareFormatsInput) (*mcp.CallToolResult, MarkdownOutput, error) {
	md := "# Video Format Comparison\n\n[Format comparison - implement full version from original code]"
	return nil, MarkdownOutput{Content: md}, nil
}
