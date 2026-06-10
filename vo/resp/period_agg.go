package resp

import (
	"github.com/Guo-Chenxu/pay-log/dal/mysql/manager"
	"github.com/Guo-Chenxu/pay-log/pkg/money"
)

type PeriodAggResp struct {
	TotalIncome      string `json:"total_income"`
	TotalExpense     string `json:"total_expense"`
	AlipayIncome     string `json:"alipay_income"`
	AlipayExpense    string `json:"alipay_expense"`
	WechatIncome     string `json:"wechat_income"`
	WechatExpense    string `json:"wechat_expense"`
	InvestmentAmount string `json:"investment_amount"`
}

func FromPeriodAgg(a *manager.PeriodAgg) PeriodAggResp {
	return PeriodAggResp{
		TotalIncome:      money.CentsToYuanString(a.TotalIncome),
		TotalExpense:     money.CentsToYuanString(a.TotalExpense),
		AlipayIncome:     money.CentsToYuanString(a.AlipayIncome),
		AlipayExpense:    money.CentsToYuanString(a.AlipayExpense),
		WechatIncome:     money.CentsToYuanString(a.WechatIncome),
		WechatExpense:    money.CentsToYuanString(a.WechatExpense),
		InvestmentAmount: money.CentsToYuanString(a.InvestmentAmount),
	}
}
