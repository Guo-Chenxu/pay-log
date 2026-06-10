package req

import (
	"mime/multipart"
	"time"

	"github.com/Guo-Chenxu/pay-log/pkg/pageutil"
)

type UploadBillReq struct {
	Channel int8                  `form:"channel" binding:"required"`
	File    *multipart.FileHeader `form:"file"    binding:"required"`
}

type ListBillsReq struct {
	Year      int    `form:"year"`
	Month     int    `form:"month"`
	SortOrder string `form:"sort_order"` // asc | desc (by amount), empty = time desc
	pageutil.PageReq
}

type ListRangeBillsReq struct {
	StartYear  int    `form:"start_year"`
	StartMonth int    `form:"start_month"`
	EndYear    int    `form:"end_year"`
	EndMonth   int    `form:"end_month"`
	SortOrder  string `form:"sort_order"`
	pageutil.PageReq
}

type AddManualBillReq struct {
	Channel         int8      `json:"channel"          binding:"required"`
	TransactionTime time.Time `json:"transaction_time" binding:"required"`
	Counterparty    string    `json:"counterparty"`
	Description     string    `json:"description"`
	Category        string    `json:"category"`
	BillType        int8      `json:"bill_type"        binding:"required"`
	Amount          string    `json:"amount"           binding:"required"`
	PaymentMethod   string    `json:"payment_method"`
	Remark          string    `json:"remark"`
	IsInvestment    bool      `json:"is_investment"`
}
