package req

import "github.com/Guo-Chenxu/pay-log/pkg/pageutil"

type OverviewReq struct {
	pageutil.PageReq
}

type MonthReq struct {
	Year  int `form:"year"`
	Month int `form:"month"`
}

type RangeReq struct {
	StartYear  int `form:"start_year"`
	StartMonth int `form:"start_month"`
	EndYear    int `form:"end_year"`
	EndMonth   int `form:"end_month"`
	pageutil.PageReq
}
