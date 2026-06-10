package router

import (
	"github.com/gin-gonic/gin"

	"github.com/Guo-Chenxu/pay-log/controller"
)

func Register(r *gin.Engine) {
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	v1 := r.Group("/api/v1")

	authCtrl := controller.NewAuthController()
	v1.POST("/auth/login", authCtrl.Login)
	v1.POST("/auth/logout", authCtrl.Logout)

	billCtrl := controller.NewBillController()
	v1.POST("/bill/upload", billCtrl.Upload)
	v1.POST("/bill/manual", billCtrl.AddManual)
	v1.GET("/bill/list", billCtrl.List)
	v1.GET("/bill/list-range", billCtrl.ListRange)
	v1.DELETE("/bill/:id", billCtrl.Delete)

	sumCtrl := controller.NewSummaryController()
	v1.GET("/summary/overview", sumCtrl.Overview)
	v1.GET("/summary/month", sumCtrl.MonthDetail)
	v1.GET("/summary/range", sumCtrl.RangeDetail)

	aiCtrl := controller.NewAIController()
	v1.POST("/ai/trigger", aiCtrl.Trigger)
	v1.GET("/ai/summary", aiCtrl.GetSummary)
}
