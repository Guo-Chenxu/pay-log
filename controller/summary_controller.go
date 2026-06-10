package controller

import (
	"github.com/gin-gonic/gin"

	"github.com/Guo-Chenxu/pay-log/pkg/auth"
	"github.com/Guo-Chenxu/pay-log/pkg/response"
	"github.com/Guo-Chenxu/pay-log/service"
	voreq "github.com/Guo-Chenxu/pay-log/vo/req"
)

type SummaryController struct{ svc *service.SummaryService }

func NewSummaryController() *SummaryController {
	return &SummaryController{svc: service.NewSummaryService()}
}

func (c *SummaryController) Overview(ctx *gin.Context) {
	userID, err := auth.GetLoginID(ctx)
	if err != nil {
		response.UnauthorizedResponse(ctx)
		return
	}
	var req voreq.OverviewReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.HandleInvalidParamsResponse(ctx, err.Error())
		return
	}
	data, err := c.svc.GetMonthOverview(ctx.Request.Context(), userID, req.Page, req.PageSize)
	response.HandleCommonResponse(ctx, data, err)
}

func (c *SummaryController) MonthDetail(ctx *gin.Context) {
	userID, err := auth.GetLoginID(ctx)
	if err != nil {
		response.UnauthorizedResponse(ctx)
		return
	}
	var req voreq.MonthReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.HandleInvalidParamsResponse(ctx, err.Error())
		return
	}
	data, err := c.svc.GetMonthDetail(ctx.Request.Context(), userID, req.Year, req.Month)
	response.HandleCommonResponse(ctx, data, err)
}

func (c *SummaryController) RangeDetail(ctx *gin.Context) {
	userID, err := auth.GetLoginID(ctx)
	if err != nil {
		response.UnauthorizedResponse(ctx)
		return
	}
	var req voreq.RangeReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.HandleInvalidParamsResponse(ctx, err.Error())
		return
	}
	if req.StartYear == 0 || req.StartMonth == 0 || req.EndYear == 0 || req.EndMonth == 0 {
		response.HandleInvalidParamsResponse(ctx, "start_year, start_month, end_year, end_month required")
		return
	}
	data, err := c.svc.GetRangeDetail(ctx.Request.Context(), userID, req.StartYear, req.StartMonth, req.EndYear, req.EndMonth, req.Page, req.PageSize)
	response.HandleCommonResponse(ctx, data, err)
}
