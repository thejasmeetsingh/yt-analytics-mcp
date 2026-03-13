package tools

import (
	"fmt"
	"time"
)

func getDateOrDefault(val string, daysOffset int) string {
	if val != "" {
		return val
	}
	return time.Now().AddDate(0, 0, daysOffset).Format("2006-01-02")
}

func formatNumber(n uint64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	if n < 1000000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	return fmt.Sprintf("%.1fB", float64(n)/1000000000)
}

func formatAnalyticsData(rows [][]interface{}) map[string]float64 {
	totals := map[string]float64{}

	for _, row := range rows {
		totals["views"] += row[1].(float64)
		totals["watchTime"] += row[2].(float64)
		totals["likes"] += row[3].(float64)
		totals["dislikes"] += row[4].(float64)
		totals["comments"] += row[4].(float64)
		totals["shares"] += row[5].(float64)
		totals["subsGained"] += row[6].(float64)
		totals["subsLost"] += row[7].(float64)
		totals["avgDuration"] += row[8].(float64)
		totals["avgPercent"] += row[9].(float64)
		totals["clickThroughRate"] += row[11].(float64)
		totals["clickImpressions"] += row[12].(float64)
		totals["thumbImpressions"] += row[13].(float64)
		totals["thumbImpressionCtr"] += row[14].(float64)
	}

	if len(rows) > 0 {
		totals["avgDuration"] /= float64(len(rows))
		totals["avgPercent"] /= float64(len(rows))
	}

	totals["engagement"] = 0
	if totals["views"] > 0 {
		totals["engagement"] = (totals["likes"] + totals["comments"] + totals["shares"]) / totals["views"] * 100
	}

	return totals
}
