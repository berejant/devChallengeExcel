package main

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
	"os"
	"testing"
)

func TestBuildServiceContainer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	f, err := os.CreateTemp("", "db_*.db")
	defer os.Remove(f.Name())

	serviceContainer, err := BuildServiceContainer(f.Name())

	assert.NoError(t, err)

	// check database
	assert.NotNil(t, serviceContainer.Database)
	assert.IsType(t, &bbolt.DB{}, serviceContainer.Database)
	assert.NoError(t, serviceContainer.Database.Close())

	// check expression executor
	assert.NotNil(t, serviceContainer.ExpressionExecutor)
	assert.IsType(t, &ExpressionExecutor{}, serviceContainer.ExpressionExecutor)

	expressionExecutor := serviceContainer.ExpressionExecutor.(*ExpressionExecutor)
	assert.IsType(t, &Canonicalizer{}, expressionExecutor.canonicalizer)

	// check webhook dispatcher
	assert.NotNil(t, serviceContainer.WebhookDispatcher)
	assert.IsType(t, &WebhookDispatcher{}, serviceContainer.WebhookDispatcher)

	// check sheet repository
	assert.NotNil(t, serviceContainer.SheetRepository)
	assert.IsType(t, &SheetRepository{}, serviceContainer.SheetRepository)

	sheetRepository := serviceContainer.SheetRepository.(*SheetRepository)
	assert.NotNil(t, sheetRepository.db)
	assert.Equal(t, serviceContainer.Database, sheetRepository.db)
	assert.Equal(t, serviceContainer.ExpressionExecutor, sheetRepository.executor)
	assert.Equal(t, serviceContainer.WebhookDispatcher, sheetRepository.webhookDispatcher)

	assert.NotNil(t, sheetRepository.serializer)
	assert.IsType(t, &CellBinarySerializer{}, sheetRepository.serializer)

	assert.NotNil(t, sheetRepository.canonicalizer)
	assert.IsType(t, &Canonicalizer{}, sheetRepository.canonicalizer)
	assert.Equal(t, expressionExecutor.canonicalizer, sheetRepository.canonicalizer)

	// check api controller
	assert.NotNil(t, serviceContainer.ApiController)
	assert.IsType(t, &ApiController{}, serviceContainer.ApiController)

	apiController := serviceContainer.ApiController.(*ApiController)
	assert.NotNil(t, apiController.SheetRepository)
	assert.Equal(t, serviceContainer.SheetRepository, apiController.SheetRepository)
	assert.NotNil(t, apiController.WebhookDispatcher)
	assert.Equal(t, serviceContainer.WebhookDispatcher, apiController.WebhookDispatcher)

	// check router
	assert.NotNil(t, serviceContainer.Router)
	assert.IsType(t, &gin.Engine{}, serviceContainer.Router)

	// check routes
	routes := serviceContainer.Router.Routes()
	assert.NotNil(t, routes)
	// 3 api route + health check
	assert.GreaterOrEqual(t, len(routes), 4)
}
