package resp

import (
	"github.com/Guo-Chenxu/pay-log/dal/mysql/manager"
	"github.com/Guo-Chenxu/pay-log/dal/mysql/model"
	"github.com/Guo-Chenxu/pay-log/pkg/money"
)

type MonthSummaryResp struct {
	ID               int64  `json:"id,string"`
	Year             int    `json:"year"`
	Month            int    `json:"month"`
	TotalIncome      string `json:"total_income"`
	TotalExpense     string `json:"total_expense"`
	AlipayIncome     string `json:"alipay_income"`
	AlipayExpense    string `json:"alipay_expense"`
	WechatIncome     string `json:"wechat_income"`
	WechatExpense    string `json:"wechat_expense"`
	InvestmentAmount string `json:"investment_amount"`
}

func FromMonthSummary(m *model.MonthSummary) *MonthSummaryResp {
	return &MonthSummaryResp{
		ID:               m.ID,
		Year:             m.Year,
		Month:            m.Month,
		TotalIncome:      money.CentsToYuanString(m.TotalIncome),
		TotalExpense:     money.CentsToYuanString(m.TotalExpense),
		AlipayIncome:     money.CentsToYuanString(m.AlipayIncome),
		AlipayExpense:    money.CentsToYuanString(m.AlipayExpense),
		WechatIncome:     money.CentsToYuanString(m.WechatIncome),
		WechatExpense:    money.CentsToYuanString(m.WechatExpense),
		InvestmentAmount: money.CentsToYuanString(m.InvestmentAmount),
	}
}

func FromPeriodAggMonthSummary(year, month int, agg *manager.PeriodAgg) *MonthSummaryResp {
	return &MonthSummaryResp{
		Year:             year,
		Month:            month,
		TotalIncome:      money.CentsToYuanString(agg.TotalIncome),
		TotalExpense:     money.CentsToYuanString(agg.TotalExpense),
		AlipayIncome:     money.CentsToYuanString(agg.AlipayIncome),
		AlipayExpense:    money.CentsToYuanString(agg.AlipayExpense),
		WechatIncome:     money.CentsToYuanString(agg.WechatIncome),
		WechatExpense:    money.CentsToYuanString(agg.WechatExpense),
		InvestmentAmount: money.CentsToYuanString(agg.InvestmentAmount),
	}
}

func FromMonthSummaries(ms []*model.MonthSummary) []*MonthSummaryResp {
	out := make([]*MonthSummaryResp, len(ms))
	for i, m := range ms {
		out[i] = FromMonthSummary(m)
	}
	return out
}
