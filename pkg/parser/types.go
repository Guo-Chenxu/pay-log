package parser

import "time"

// ParsedRecord is the normalized bill record after parsing either Alipay or WeChat bill.
type ParsedRecord struct {
	TransactionTime time.Time
	Counterparty    string
	Description     string
	Category        string
	BillType        int8 // see consts.BillType*
	Amount          int64
	PaymentMethod   string
	Status          string
	OrderNo         string
	MerchantOrderNo string
	Remark          string
	IsInvestment    bool
}
