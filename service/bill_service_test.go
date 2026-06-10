package service

import (
	"testing"
	"time"

	"github.com/Guo-Chenxu/pay-log/consts"
	voreq "github.com/Guo-Chenxu/pay-log/vo/req"
)

func TestNormalizeManualBillParsesAmountAndForcesInvestmentNeutral(t *testing.T) {
	req := voreq.AddManualBillReq{
		Channel:         int8(consts.ChannelAlipay),
		TransactionTime: time.Date(2026, 6, 11, 21, 30, 0, 0, time.Local),
		BillType:        int8(consts.BillTypeExpense),
		Amount:          "1000.05",
		IsInvestment:    true,
	}

	got, err := normalizeManualBill(req)
	if err != nil {
		t.Fatalf("normalizeManualBill returned error: %v", err)
	}
	if got.Amount != 100005 {
		t.Fatalf("Amount = %d, want 100005", got.Amount)
	}
	if got.BillType != int8(consts.BillTypeNeutral) {
		t.Fatalf("BillType = %d, want neutral", got.BillType)
	}
	if !got.IsInvestment {
		t.Fatalf("IsInvestment = false, want true")
	}
}

func TestNormalizeManualBillRejectsInvalidEnumsAndAmount(t *testing.T) {
	tests := []struct {
		name string
		req  voreq.AddManualBillReq
	}{
		{name: "invalid channel", req: voreq.AddManualBillReq{Channel: 9, BillType: int8(consts.BillTypeExpense), Amount: "1.00"}},
		{name: "invalid bill type", req: voreq.AddManualBillReq{Channel: int8(consts.ChannelAlipay), BillType: 9, Amount: "1.00"}},
		{name: "invalid amount", req: voreq.AddManualBillReq{Channel: int8(consts.ChannelAlipay), BillType: int8(consts.BillTypeExpense), Amount: "1.234"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := normalizeManualBill(tt.req); err == nil {
				t.Fatalf("normalizeManualBill returned nil error and %+v, want error", got)
			}
		})
	}
}
