package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	openai "github.com/sashabaranov/go-openai"

	"github.com/Guo-Chenxu/pay-log/config"
	"github.com/Guo-Chenxu/pay-log/consts"
	"github.com/Guo-Chenxu/pay-log/dal/mysql/manager"
	"github.com/Guo-Chenxu/pay-log/dal/mysql/model"
	"github.com/Guo-Chenxu/pay-log/pkg/logger"
	"github.com/Guo-Chenxu/pay-log/pkg/money"
	"github.com/Guo-Chenxu/pay-log/pkg/snowflake"
)

type AIService struct {
	aiMgr    *manager.AISummaryManager
	monthMgr *manager.MonthSummaryManager
	billMgr  *manager.BillRecordManager
}

var (
	aiSvc     *AIService
	aiSvcOnce sync.Once
)

func NewAIService() *AIService {
	aiSvcOnce.Do(func() {
		aiSvc = &AIService{
			aiMgr:    manager.NewAISummaryManager(),
			monthMgr: manager.NewMonthSummaryManager(),
			billMgr:  manager.NewBillRecordManager(),
		}
	})
	return aiSvc
}

func validateAISummaryRange(startYear, startMonth, endYear, endMonth int) error {
	return validateYearMonthRange(startYear, startMonth, endYear, endMonth)
}

func (s *AIService) ValidateSummaryRange(startYear, startMonth, endYear, endMonth int) error {
	return validateAISummaryRange(startYear, startMonth, endYear, endMonth)
}

// TriggerAISummary generates an AI analysis for the given month range.
func (s *AIService) TriggerAISummary(ctx context.Context, userID int64, startYear, startMonth, endYear, endMonth int, triggerType int8) error {
	if err := validateAISummaryRange(startYear, startMonth, endYear, endMonth); err != nil {
		return err
	}

	// Collect monthly summaries in range
	var summaries []*model.MonthSummary
	y, m := startYear, startMonth
	for {
		ms, err := s.monthMgr.Get(userID, y, m)
		if err == nil {
			summaries = append(summaries, ms)
		}
		if y == endYear && m == endMonth {
			break
		}
		m++
		if m > 12 {
			m = 1
			y++
		}
	}

	// Collect bill details for the range
	start := time.Date(startYear, time.Month(startMonth), 1, 0, 0, 0, 0, time.Local)
	end := time.Date(endYear, time.Month(endMonth), 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)
	bills, _, err := s.billMgr.ListByRange(userID, start, end, 0, 1000)
	if err != nil {
		logger.CtxErrorf(ctx, "list bills for AI prompt failed: %v", err)
		bills = nil
	}

	prompt := buildPrompt(summaries, bills, startYear, startMonth, endYear, endMonth)
	analysis := callAI(ctx, prompt)

	if err := s.aiMgr.SoftDeleteByRange(userID, startYear, startMonth, endYear, endMonth); err != nil {
		logger.CtxErrorf(ctx, "soft delete ai_summary failed: %v", err)
	}
	return s.aiMgr.Create(&model.AISummary{
		ID:          snowflake.GenerateID(),
		UserID:      userID,
		StartYear:   startYear,
		StartMonth:  startMonth,
		EndYear:     endYear,
		EndMonth:    endMonth,
		AIAnalysis:  analysis,
		TriggerType: triggerType,
	})
}

func (s *AIService) GetSummary(ctx context.Context, userID int64, startYear, startMonth, endYear, endMonth int) (*model.AISummary, error) {
	return s.aiMgr.GetByRange(userID, startYear, startMonth, endYear, endMonth)
}

func buildPrompt(summaries []*model.MonthSummary, bills []*model.BillRecord, sy, sm, ey, em int) string {
	buf := &strings.Builder{}
	fmt.Fprintf(buf, "以下是我 %d年%d月 至 %d年%d月 的账单数据，请简要分析消费结构和趋势：\n\n", sy, sm, ey, em)

	if len(summaries) > 0 {
		buf.WriteString("【月度汇总】\n")
		for _, s := range summaries {
			fmt.Fprintf(buf, "%d年%d月：总收入%s元，总支出%s元，支付宝支出%s元，微信支出%s元，理财支出%s元\n",
				s.Year, s.Month,
				money.CentsToYuanString(s.TotalIncome),
				money.CentsToYuanString(s.TotalExpense),
				money.CentsToYuanString(s.AlipayExpense),
				money.CentsToYuanString(s.WechatExpense),
				money.CentsToYuanString(s.InvestmentAmount))
		}
		buf.WriteString("\n")
	}

	if len(bills) > 0 {
		buf.WriteString("【账单明细】\n")
		for _, b := range bills {
			fmt.Fprintf(buf, "%s %s %s %s %s元\n",
				b.TransactionTime.Format("2006-01-02"),
				consts.ChannelLabel(b.Channel),
				consts.BillTypeLabel(b.BillType),
				b.Description,
				money.CentsToYuanString(b.Amount),
			)
		}
	}

	return buf.String()
}

func callAI(ctx context.Context, prompt string) string {
	cfg := config.GetModelConfig()
	ocfg := openai.DefaultConfig(cfg.APIKey)
	ocfg.BaseURL = cfg.BaseURL
	client := openai.NewClientWithConfig(ocfg)

	reqCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	resp, err := client.CreateChatCompletion(reqCtx, openai.ChatCompletionRequest{
		Model: cfg.ModelID,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
	})
	if err != nil {
		logger.CtxErrorf(ctx, "AI call failed: %v", err)
		return "AI 分析暂时不可用，请稍后再试。"
	}
	if len(resp.Choices) == 0 {
		return "AI 分析返回为空。"
	}
	return resp.Choices[0].Message.Content
}
