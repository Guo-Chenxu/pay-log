package manager

import (
	"testing"

	"github.com/Guo-Chenxu/pay-log/consts"
)

func TestApplyPeriodAggRowUsesCentsAndSkipsInvestmentFromExpense(t *testing.T) {
	agg := &PeriodAgg{}

	applyPeriodAggRow(agg, int8(consts.ChannelAlipay), int8(consts.BillTypeExpense), false, 1234)
	applyPeriodAggRow(agg, int8(consts.ChannelWechat), int8(consts.BillTypeIncome), false, 5000)
	applyPeriodAggRow(agg, int8(consts.ChannelAlipay), int8(consts.BillTypeNeutral), true, 100000)
	applyPeriodAggRow(agg, int8(consts.ChannelWechat), int8(consts.BillTypeNeutral), false, 999999)
	applyPeriodAggRow(agg, int8(consts.ChannelWechat), int8(consts.BillTypeExpense), true, 4321)
	applyPeriodAggRow(agg, int8(consts.ChannelAlipay), int8(consts.BillTypeIncome), true, 7654)

	if agg.AlipayIncome != 0 {
		t.Fatalf("AlipayIncome = %d, want 0", agg.AlipayIncome)
	}
	if agg.AlipayExpense != 1234 {
		t.Fatalf("AlipayExpense = %d, want 1234", agg.AlipayExpense)
	}
	if agg.WechatIncome != 5000 {
		t.Fatalf("WechatIncome = %d, want 5000", agg.WechatIncome)
	}
	if agg.WechatExpense != 0 {
		t.Fatalf("WechatExpense = %d, want 0", agg.WechatExpense)
	}
	if agg.InvestmentAmount != 111975 {
		t.Fatalf("InvestmentAmount = %d, want 111975", agg.InvestmentAmount)
	}
	if agg.TotalExpense != 1234 {
		t.Fatalf("TotalExpense = %d, want 1234", agg.TotalExpense)
	}
	if agg.TotalIncome != 5000 {
		t.Fatalf("TotalIncome = %d, want 5000", agg.TotalIncome)
	}
}
