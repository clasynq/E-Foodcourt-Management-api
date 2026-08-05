package service

import (
	"context"
	"fmt"
	"strings"

	"manager-dashboard/internal/model"
	"manager-dashboard/internal/repository"

	"github.com/redis/go-redis/v9"
)

type DashboardService struct {
	repo *repository.DashboardRepository
	rdb  *redis.Client
}

func NewDashboardService(repo *repository.DashboardRepository, rdb *redis.Client) *DashboardService {
	return &DashboardService{repo: repo, rdb: rdb}
}

func (s *DashboardService) GetOverview(ctx context.Context) (*model.ManagerOverviewResponse, error) {
	// 1. Fetch pending orders dynamically
	pending, err := s.repo.GetPendingAndPreparingOrders()
	if err != nil {
		return nil, err
	}

	// 2. Fetch low inventory dynamically
	lowStockItems, err := s.repo.GetLowStockItems()
	if err != nil {
		return nil, err
	}

	// 3. Compile low inventory warning messages based on query results
	var alert *model.LowInventoryAlert
	if len(lowStockItems) > 0 {
		var names []string
		for _, item := range lowStockItems {
			names = append(names, item.Name)
		}

		var alertMsg string
		if len(names) == 1 {
			alertMsg = fmt.Sprintf("%s is running low. Consider restocking before dinner rush.", names[0])
		} else if len(names) == 2 {
			alertMsg = fmt.Sprintf("%s and %s are running low. Consider restocking before dinner rush.", names[0], names[1])
		} else {
			lastIdx := len(names) - 1
			alertMsg = fmt.Sprintf("%s, and %s are running low. Consider restocking before dinner rush.",
				strings.Join(names[:lastIdx], ", "), names[lastIdx])
		}

		alert = &model.LowInventoryAlert{
			Message: alertMsg,
			Items:   names,
		}
	}

	// 4. Aggregate stats dynamically
	todayCount, todayRevenue, todayAvgPrep, err := s.repo.GetTodayStats()
	if err != nil {
		return nil, err
	}

	yesterdayCount, yesterdayRevenue, yesterdayAvgPrep, err := s.repo.GetYesterdayStats()
	if err != nil {
		return nil, err
	}

	// 5. Query active user sessions from Redis
	activeUsersCount := s.getActiveUsers(ctx)
	yesterdayActiveUsers := s.getYesterdayActiveUsers(ctx) // compare Redis logs or default fallback

	// Calculate changes dynamically
	countChange := s.calculatePctChange(float64(todayCount), float64(yesterdayCount))
	revenueChange := s.calculatePctChange(todayRevenue, yesterdayRevenue)
	prepChange := todayAvgPrep - yesterdayAvgPrep
	activeChange := s.calculatePctChange(float64(activeUsersCount), float64(yesterdayActiveUsers))

	// Build stats card schema mapping
	stats := []model.StatsCard{
		{
			Title:     "Today's Orders",
			Value:     fmt.Sprintf("%d", todayCount),
			Change:    s.formatChange(countChange),
			Icon:      "ShoppingBag",
			BgColor:   "bg-blue-500/10",
			TextColor: "text-blue-600 dark:text-blue-400",
		},
		{
			Title:     "Today's Revenue",
			Value:     fmt.Sprintf("₹%s", s.formatCurrency(todayRevenue)),
			Change:    s.formatChange(revenueChange),
			Icon:      "DollarSign",
			BgColor:   "bg-emerald-500/10",
			TextColor: "text-emerald-600 dark:text-emerald-400",
		},
		{
			Title:     "Avg Prep Time",
			Value:     fmt.Sprintf("%.0f min", todayAvgPrep),
			Change:    s.formatPrepTimeChange(prepChange),
			Icon:      "Clock",
			BgColor:   "bg-amber-500/10",
			TextColor: "text-amber-600 dark:text-amber-400",
		},
		{
			Title:     "Active Users",
			Value:     fmt.Sprintf("%d", activeUsersCount),
			Change:    s.formatChange(activeChange),
			Icon:      "Users",
			BgColor:   "bg-purple-500/10",
			TextColor: "text-purple-600 dark:text-purple-400",
		},
	}

	return &model.ManagerOverviewResponse{
		Stats:             stats,
		PendingOrders:     pending,
		LowInventoryAlert: alert,
	}, nil
}

// getActiveUsers queries Redis for active user session keys
func (s *DashboardService) getActiveUsers(ctx context.Context) int64 {
	keys, err := s.rdb.Keys(ctx, "user:email:*").Result()
	if err != nil {
		return 0
	}
	return int64(len(keys))
}

// getYesterdayActiveUsers queries historical records in Redis or defaults to active counts
func (s *DashboardService) getYesterdayActiveUsers(ctx context.Context) int64 {
	val, err := s.rdb.Get(ctx, "stats:active_users:yesterday").Int64()
	if err != nil {
		return 0
	}
	return val
}

func (s *DashboardService) calculatePctChange(current, previous float64) float64 {
	if previous == 0 {
		if current > 0 {
			return 100.0
		}
		return 0.0
	}
	return ((current - previous) / previous) * 100.0
}

func (s *DashboardService) formatChange(pct float64) string {
	if pct == 0 {
		return "0%"
	}
	sign := "+"
	if pct < 0 {
		sign = ""
	}
	return fmt.Sprintf("%s%.0f%%", sign, pct)
}

func (s *DashboardService) formatPrepTimeChange(diff float64) string {
	if diff == 0 {
		return "0 min"
	}
	sign := "+"
	if diff < 0 {
		sign = ""
	}
	return fmt.Sprintf("%s%.0f min", sign, diff)
}

func (s *DashboardService) formatCurrency(val float64) string {
	str := fmt.Sprintf("%.0f", val)
	if len(str) <= 3 {
		return str
	}
	lastThree := str[len(str)-3:]
	rest := str[:len(str)-3]
	if len(rest) == 0 {
		return lastThree
	}
	return fmt.Sprintf("%s,%s", rest, lastThree)
}
