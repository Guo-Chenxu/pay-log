package controller

import (
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Guo-Chenxu/pay-log/consts"
	"github.com/Guo-Chenxu/pay-log/pkg/auth"
	"github.com/Guo-Chenxu/pay-log/pkg/bizerr"
	"github.com/Guo-Chenxu/pay-log/pkg/response"
	"github.com/Guo-Chenxu/pay-log/service"
	voreq "github.com/Guo-Chenxu/pay-log/vo/req"
)

type BillController struct{ svc *service.BillService }

func NewBillController() *BillController {
	return &BillController{svc: service.NewBillService()}
}

func (c *BillController) Upload(ctx *gin.Context) {
	userID, err := auth.GetLoginID(ctx)
	if err != nil {
		response.FailResponse(ctx, bizerr.Unauthorized.Code, bizerr.Unauthorized.Message)
		return
	}
	var req voreq.UploadBillReq
	if err := ctx.ShouldBind(&req); err != nil {
		response.HandleInvalidParamsResponse(ctx, err.Error())
		return
	}
	if !consts.IsValidChannel(req.Channel) {
		response.HandleInvalidParamsResponse(ctx, "channel must be 1 (alipay) or 2 (wechat)")
		return
	}
	tmp, err := os.CreateTemp("", "bill-*")
	if err != nil {
		response.HandleCommonResponse(ctx, nil, err)
		return
	}
	tmpPath := tmp.Name()
	tmp.Close()
	if err = ctx.SaveUploadedFile(req.File, tmpPath); err != nil {
		os.Remove(tmpPath)
		response.HandleCommonResponse(ctx, nil, err)
		return
	}
	result, err := c.svc.ImportFile(ctx.Request.Context(), userID, tmpPath, req.Channel)
	response.HandleCommonResponse(ctx, result, err)
}

func (c *BillController) AddManual(ctx *gin.Context) {
	userID, err := auth.GetLoginID(ctx)
	if err != nil {
		response.FailResponse(ctx, bizerr.Unauthorized.Code, bizerr.Unauthorized.Message)
		return
	}
	var req voreq.AddManualBillReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleInvalidParamsResponse(ctx, err.Error())
		return
	}
	response.HandleCommonResponse(ctx, nil, c.svc.AddManual(ctx.Request.Context(), userID, req))
}

func (c *BillController) List(ctx *gin.Context) {
	userID, err := auth.GetLoginID(ctx)
	if err != nil {
		response.FailResponse(ctx, bizerr.Unauthorized.Code, bizerr.Unauthorized.Message)
		return
	}
	var req voreq.ListBillsReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.HandleInvalidParamsResponse(ctx, err.Error())
		return
	}
	if req.Year == 0 || req.Month == 0 {
		response.HandleInvalidParamsResponse(ctx, "year and month required")
		return
	}
	data, err := c.svc.ListByMonth(ctx.Request.Context(), userID, req.Year, req.Month, req.Page, req.PageSize, req.SortOrder)
	response.HandleCommonResponse(ctx, data, err)
}

func (c *BillController) ListRange(ctx *gin.Context) {
	userID, err := auth.GetLoginID(ctx)
	if err != nil {
		response.FailResponse(ctx, bizerr.Unauthorized.Code, bizerr.Unauthorized.Message)
		return
	}
	var req voreq.ListRangeBillsReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.HandleInvalidParamsResponse(ctx, err.Error())
		return
	}
	if req.StartYear == 0 || req.StartMonth == 0 || req.EndYear == 0 || req.EndMonth == 0 {
		response.HandleInvalidParamsResponse(ctx, "start_year, start_month, end_year, end_month required")
		return
	}
	data, err := c.svc.ListByRange(ctx.Request.Context(), userID, req.StartYear, req.StartMonth, req.EndYear, req.EndMonth, req.Page, req.PageSize, req.SortOrder)
	response.HandleCommonResponse(ctx, data, err)
}

func (c *BillController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		response.HandleInvalidParamsResponse(ctx, "invalid id")
		return
	}
	userID, err := auth.GetLoginID(ctx)
	if err != nil {
		response.FailResponse(ctx, bizerr.Unauthorized.Code, bizerr.Unauthorized.Message)
		return
	}
	response.HandleCommonResponse(ctx, nil, c.svc.Delete(ctx.Request.Context(), userID, id))
}
