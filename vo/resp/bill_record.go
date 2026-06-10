package resp

import (
	"time"

	"github.com/Guo-Chenxu/pay-log/dal/mysql/model"
	"github.com/Guo-Chenxu/pay-log/pkg/money"
)

type BillRecordResp struct {
	ID              int64     `json:"id,string"`
	Channel         int8      `json:"channel"`
	TransactionTime time.Time `json:"transaction_time"`
	Category        string    `json:"category"`
	Description     string    `json:"description"`
	BillType        int8      `json:"bill_type"`
	Amount          string    `json:"amount"`
}

func FromBillRecord(b *model.BillRecord) *BillRecordResp {
	return &BillRecordResp{
		ID:              b.ID,
		Channel:         b.Channel,
		TransactionTime: b.TransactionTime,
		Category:        b.Category,
		Description:     b.Description,
		BillType:        b.BillType,
		Amount:          money.CentsToYuanString(b.Amount),
	}
}

func FromBillRecords(records []*model.BillRecord) []*BillRecordResp {
	out := make([]*BillRecordResp, len(records))
	for i, b := range records {
		out[i] = FromBillRecord(b)
	}
	return out
}
