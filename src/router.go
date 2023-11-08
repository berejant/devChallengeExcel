package main

import (
	"devChallengeExcel/contracts"
	"github.com/gin-gonic/gin"
	"net/http"
)

const ApiVersion = "v1"

const externalRefWebhookPath = "externalRefWebhook"
const subscribePath = "subscribe"

func SetupRouter(controller contracts.ApiController) *gin.Engine {
	router := gin.New()

	apiRouterGroup := router.Group("/api/" + ApiVersion)
	apiRouterGroup.POST("/:sheet_id/:cell_id/"+subscribePath, controller.SubscribeAction)
	apiRouterGroup.POST("/:sheet_id/:cell_id/"+externalRefWebhookPath, controller.ExternalRefWebhookAction)

	apiRouterGroup.POST("/:sheet_id/:cell_id", controller.SetCellAction)
	apiRouterGroup.GET("/:sheet_id/:cell_id", controller.GetCellAction)
	apiRouterGroup.GET("/:sheet_id", controller.GetSheetAction)

	router.GET("/healthcheck", func(c *gin.Context) {
		c.String(http.StatusOK, "health")
	})

	return router
}
