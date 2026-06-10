package resp

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Guo-Chenxu/pay-log/consts"
	"github.com/Guo-Chenxu/pay-log/dal/mysql/model"
)

func TestBillRecordRespJSONContract(t *testing.T) {
	txTime := time.Date(2026, 6, 11, 21, 30, 0, 0, time.FixedZone("CST", 8*60*60))
	resp := FromBillRecord(&model.BillRecord{
		ID:              922337203685477000,
		Channel:         int8(consts.ChannelAlipay),
		TransactionTime: txTime,
		Category:        "餐饮",
		Description:     "晚餐",
		BillType:        int8(consts.BillTypeExpense),
		Amount:          1234,
	})

	body, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	jsonText := string(body)

	if !strings.Contains(jsonText, `"id":"922337203685477000"`) {
		t.Fatalf("id must be encoded as string, got %s", jsonText)
	}
	if !strings.Contains(jsonText, `"transaction_time":"2026-06-11T21:30:00+08:00"`) {
		t.Fatalf("transaction_time must be RFC3339, got %s", jsonText)
	}
	if !strings.Contains(jsonText, `"amount":"12.34"`) {
		t.Fatalf("amount must be yuan string, got %s", jsonText)
	}
}
