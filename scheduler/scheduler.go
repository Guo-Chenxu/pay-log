package scheduler

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/Guo-Chenxu/pay-log/consts"
	"github.com/Guo-Chenxu/pay-log/dal/mysql/manager"
	"github.com/Guo-Chenxu/pay-log/pkg/logger"
	"github.com/Guo-Chenxu/pay-log/service"
)

// Start registers cron jobs and starts the scheduler.
func Start() {
	c := cron.New()

	// Run on the 1st of every month at 01:00 — summarize all unsummarized months for all users
	c.AddFunc("0 1 1 * *", func() { //nolint:errcheck
		summarizeAllUsers()
	})

	c.Start()
}

func summarizeAllUsers() {
	ctx := context.Background()
	sumSvc := service.NewSummaryService()
	aiSvc := service.NewAIService()
	userMgr := manager.NewUserManager()

	users, err := userMgr.ListAll()
	if err != nil {
		logger.Errorf("scheduler: list users failed: %v", err)
		return
	}
	for _, u := range users {
		if err := sumSvc.SummarizeAllUnsummarized(ctx, u.ID); err != nil {
			logger.Errorf("scheduler: summarize unsummarized months for user %d: %v", u.ID, err)
		}
		// Trigger AI for previous month
		now := time.Now()
		prev := now.AddDate(0, -1, 0)
		if err := aiSvc.TriggerAISummary(ctx, u.ID, prev.Year(), int(prev.Month()), prev.Year(), int(prev.Month()), int8(consts.AITriggerAuto)); err != nil {
			logger.Errorf("scheduler: ai summary %d/%d for user %d: %v", prev.Year(), int(prev.Month()), u.ID, err)
		}
	}
}
