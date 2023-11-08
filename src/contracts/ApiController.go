package contracts

import "github.com/gin-gonic/gin"

type ApiController interface {
	SetCellAction(c *gin.Context)
	GetCellAction(c *gin.Context)
	GetSheetAction(c *gin.Context)
	SubscribeAction(c *gin.Context)
	ExternalRefWebhookAction(c *gin.Context)
}
