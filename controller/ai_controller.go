package controller

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/Guo-Chenxu/pay-log/consts"
	"github.com/Guo-Chenxu/pay-log/pkg/auth"
	"github.com/Guo-Chenxu/pay-log/pkg/bizerr"
	"github.com/Guo-Chenxu/pay-log/pkg/logger"
	"github.com/Guo-Chenxu/pay-log/pkg/response"
	"github.com/Guo-Chenxu/pay-log/service"
	voreq "github.com/Guo-Chenxu/pay-log/vo/req"
	voresp "github.com/Guo-Chenxu/pay-log/vo/resp"
)

type AIController struct{ svc *service.AIService }

func NewAIController() *AIController {
	return &AIController{svc: service.NewAIService()}
}

func (c *AIController) Trigger(ctx *gin.Context) {
	userID, err := auth.GetLoginID(ctx)
	if err != nil {
		response.FailResponse(ctx, bizerr.Unauthorized.Code, bizerr.Unauthorized.Message)
		return
	}
	var req voreq.TriggerAIReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleInvalidParamsResponse(ctx, err.Error())
		return
	}
	if err := c.svc.ValidateSummaryRange(req.StartYear, req.StartMonth, req.EndYear, req.EndMonth); err != nil {
		response.HandleCommonResponse(ctx, nil, err)
		return
	}
	// async — return immediately, process in background
	requestCtx := context.WithoutCancel(ctx.Request.Context())
	go func() {
		if err := c.svc.TriggerAISummary(requestCtx, userID,
			req.StartYear, req.StartMonth, req.EndYear, req.EndMonth, int8(consts.AITriggerManual)); err != nil {
			logger.Errorf("async AI trigger failed for user %d: %v", userID, err)
		}
	}()
	response.HandleCommonResponse(ctx, nil, nil)
}

func (c *AIController) GetSummary(ctx *gin.Context) {
	userID, err := auth.GetLoginID(ctx)
	if err != nil {
		response.FailResponse(ctx, bizerr.Unauthorized.Code, bizerr.Unauthorized.Message)
		return
	}
	var req voreq.GetAISummaryReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.HandleInvalidParamsResponse(ctx, err.Error())
		return
	}
	data, err := c.svc.GetSummary(ctx.Request.Context(), userID, req.StartYear, req.StartMonth, req.EndYear, req.EndMonth)
	response.HandleCommonResponse(ctx, voresp.FromAISummary(data), err)
}
