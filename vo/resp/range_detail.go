package resp

type RangeDetailResp struct {
	Agg    PeriodAggResp      `json:"agg"`
	Months []*MonthSummaryResp `json:"months"`
}
