package mcp

import (
	"context"
	"fmt"
	"strings"
)

func (s *MCPServer) listChannels(ctx context.Context) (string, error) {
	key := "channels_list"
	if cached, ok := s.cache.Get(key); ok {
		return cached.(string), nil
	}
	call := s.youtubeService.Channels.List([]string{"snippet", "statistics"}).Mine(true)
	resp, err := call.Do()
	if err != nil {
		return "", err
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
	s.cache.Set(key, result)
	return result, nil
}

func (s *MCPServer) getChannelDetails(ctx context.Context, channelID string, force bool) (string, error) {
	key := "ch_" + channelID
	if !force {
		if cached, ok := s.cache.Get(key); ok {
			return cached.(string), nil
		}
	}
	resp, err := s.youtubeService.Channels.List([]string{"snippet", "statistics"}).Id(channelID).Do()
	if err != nil || len(resp.Items) == 0 {
		return "", fmt.Errorf("channel not found")
	}
	ch := resp.Items[0]
	md := fmt.Sprintf("# %s\n\n## Statistics\n\n- **Subscribers**: %s\n- **Videos**: %s\n- **Views**: %s\n",
		ch.Snippet.Title, formatNumber(ch.Statistics.SubscriberCount),
		formatNumber(ch.Statistics.VideoCount), formatNumber(ch.Statistics.ViewCount))
	s.cache.Set(key, md)
	return md, nil
}

func (s *MCPServer) getChannelAnalytics(ctx context.Context, channelID, start, end string, force bool) (string, error) {
	key := fmt.Sprintf("analytics_%s_%s_%s", channelID, start, end)
	if !force {
		if cached, ok := s.cache.Get(key); ok {
			return cached.(string), nil
		}
	}
	metrics := "views,estimatedMinutesWatched,likes,comments,shares,subscribersGained,subscribersLost,averageViewDuration,averageViewPercentage"
	resp, err := s.analyticsService.Reports.Query().
		Ids("channel==" + channelID).StartDate(start).EndDate(end).
		Metrics(metrics).Dimensions("day").Do()
	if err != nil {
		return "", err
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
	s.cache.Set(key, md)
	return md, nil
}

func (s *MCPServer) getVideoList(ctx context.Context, channelID string, max int64, force bool) (string, error) {
	key := fmt.Sprintf("videos_%s_%d", channelID, max)
	if !force {
		if cached, ok := s.cache.Get(key); ok {
			return cached.(string), nil
		}
	}
	chResp, err := s.youtubeService.Channels.List([]string{"contentDetails"}).Id(channelID).Do()
	if err != nil || len(chResp.Items) == 0 {
		return "", fmt.Errorf("channel not found")
	}
	playlistID := chResp.Items[0].ContentDetails.RelatedPlaylists.Uploads
	resp, err := s.youtubeService.PlaylistItems.List([]string{"snippet", "contentDetails"}).
		PlaylistId(playlistID).MaxResults(max).Do()
	if err != nil {
		return "", err
	}
	var md strings.Builder
	md.WriteString(fmt.Sprintf("# Videos (%d)\n\n", len(resp.Items)))
	for i, item := range resp.Items {
		md.WriteString(fmt.Sprintf("## %d. %s\n- **ID**: `%s`\n- **Description**: '%s'\n- **Published**: %s\n\n",
			i+1, item.Snippet.Title, item.ContentDetails.VideoId, item.Snippet.Description, item.ContentDetails.VideoPublishedAt))
	}
	result := md.String()
	s.cache.Set(key, result)
	return result, nil
}

func (s *MCPServer) getVideoAnalytics(ctx context.Context, videoIDs, start, end string, force bool) (string, error) {
	key := fmt.Sprintf("video_analytics_%s_%s_%s", videoIDs, start, end)
	if !force {
		if cached, ok := s.cache.Get(key); ok {
			return cached.(string), nil
		}
	}
	ids := strings.Split(videoIDs, ",")
	for i := range ids {
		ids[i] = strings.TrimSpace(ids[i])
	}
	videoResp, err := s.youtubeService.Videos.List([]string{"snippet", "statistics"}).
		Id(strings.Join(ids, ",")).Do()
	if err != nil {
		return "", err
	}
	var md strings.Builder
	md.WriteString(fmt.Sprintf("# Video Analytics (%s to %s)\n\n", start, end))
	for _, video := range videoResp.Items {
		md.WriteString(fmt.Sprintf("## %s\n\n- **ID**: `%s`\n- **Views**: %s\n- **Likes**: %s\n- **Comments**: %s\n\n",
			video.Snippet.Title, video.Id, formatNumber(video.Statistics.ViewCount),
			formatNumber(video.Statistics.LikeCount), formatNumber(video.Statistics.CommentCount)))
		metrics := "views,estimatedMinutesWatched,averageViewDuration,averageViewPercentage,likes,comments,shares"
		analyticsResp, err := s.analyticsService.Reports.Query().
			Ids("channel==MINE").StartDate(start).EndDate(end).
			Metrics(metrics).Filters("video==" + video.Id).Do()
		if err == nil && len(analyticsResp.Rows) > 0 {
			row := analyticsResp.Rows[0]
			if len(row) >= 7 {
				md.WriteString(fmt.Sprintf("### Period Stats\n- **Views**: %s\n- **Watch Time**: %.1f hrs\n"+
					"- **Avg Duration**: %.0fs\n- **Avg View %%**: %.1f%%\n"+
					"- **Likes**: %s\n- **Comments**: %s\n- **Shares**: %s\n\n",
					formatNumber(uint64(row[0].(float64))), row[1].(float64)/60,
					row[2].(float64), row[3].(float64),
					formatNumber(uint64(row[4].(float64))),
					formatNumber(uint64(row[5].(float64))),
					formatNumber(uint64(row[6].(float64)))))
			}
		}
	}
	result := md.String()
	s.cache.Set(key, result)
	return result, nil
}
