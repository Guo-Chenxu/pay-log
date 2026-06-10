package req

type TriggerAIReq struct {
	StartYear  int `json:"start_year"  binding:"required"`
	StartMonth int `json:"start_month" binding:"required"`
	EndYear    int `json:"end_year"    binding:"required"`
	EndMonth   int `json:"end_month"   binding:"required"`
}

type GetAISummaryReq struct {
	StartYear  int `form:"start_year"`
	StartMonth int `form:"start_month"`
	EndYear    int `form:"end_year"`
	EndMonth   int `form:"end_month"`
}
