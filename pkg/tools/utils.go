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
