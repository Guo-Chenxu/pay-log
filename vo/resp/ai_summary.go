package resp

import "github.com/Guo-Chenxu/pay-log/dal/mysql/model"

type AISummaryResp struct {
	ID          int64  `json:"id,string"`
	StartYear   int    `json:"start_year"`
	StartMonth  int    `json:"start_month"`
	EndYear     int    `json:"end_year"`
	EndMonth    int    `json:"end_month"`
	AIAnalysis  string `json:"ai_analysis"`
	TriggerType int8   `json:"trigger_type"`
}

func FromAISummary(a *model.AISummary) *AISummaryResp {
	if a == nil {
		return nil
	}
	return &AISummaryResp{
		ID:          a.ID,
		StartYear:   a.StartYear,
		StartMonth:  a.StartMonth,
		EndYear:     a.EndYear,
		EndMonth:    a.EndMonth,
		AIAnalysis:  a.AIAnalysis,
		TriggerType: a.TriggerType,
	}
}
