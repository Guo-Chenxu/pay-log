package controller

import (
	"github.com/gin-gonic/gin"

	"github.com/Guo-Chenxu/pay-log/config"
	"github.com/Guo-Chenxu/pay-log/pkg/response"
	"github.com/Guo-Chenxu/pay-log/service"
	voreq "github.com/Guo-Chenxu/pay-log/vo/req"
	voresp "github.com/Guo-Chenxu/pay-log/vo/resp"
)

type AuthController struct{ svc *service.AuthService }

func NewAuthController() *AuthController {
	return &AuthController{svc: service.NewAuthService()}
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req voreq.LoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.HandleInvalidParamsResponse(ctx, err.Error())
		return
	}
	token, err := c.svc.Login(ctx.Request.Context(), req.Username, req.Password)
	response.HandleCommonResponse(ctx, voresp.LoginResp{
		Token:     token,
		TokenName: config.GetAuthConfig().TokenName,
	}, err)
}

func (c *AuthController) Logout(ctx *gin.Context) {
	token := ctx.GetHeader(config.GetAuthConfig().TokenName)
	err := c.svc.Logout(ctx.Request.Context(), token)
	response.HandleCommonResponse(ctx, nil, err)
}
