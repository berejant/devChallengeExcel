package main

import (
	"devChallengeExcel/contracts"
	"github.com/gin-gonic/gin"
	"net/http"
)

const ApiVersion = "v1"

func SetupRouter(controller contracts.ApiController) *gin.Engine {
	router := gin.New()

	apiRouterGroup := router.Group("/api/" + ApiVersion)
	apiRouterGroup.POST("/:sheet_id/:cell_id", controller.SetCellAction)
	apiRouterGroup.GET("/:sheet_id/:cell_id", controller.GetCellAction)
	apiRouterGroup.GET("/:sheet_id", controller.GetSheetAction)

	router.GET("/healthcheck", func(c *gin.Context) {
		c.String(http.StatusOK, "health")
	})

	return router
}
