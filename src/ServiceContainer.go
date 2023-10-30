package main

import (
	"devChallengeExcel/contracts"
	"github.com/gin-gonic/gin"
	"go.etcd.io/bbolt"
)

type ServiceContainer struct {
	Database           *bbolt.DB
	ApiController      contracts.ApiController
	SheetRepository    contracts.SheetRepository
	ExpressionExecutor contracts.ExpressionExecutor
	Router             *gin.Engine
}

func BuildServiceContainer(configDbPath string) (container ServiceContainer, err error) {
	container.Database, err = bbolt.Open(configDbPath, 0600, nil)
	serializer := NewCellBinarySerializer()
	canonicalizer := NewCanonicalizer()

	container.ExpressionExecutor = NewExpressionExecutor(canonicalizer)
	container.SheetRepository = NewSheetRepository(container.Database, container.ExpressionExecutor, serializer, canonicalizer)
	container.ApiController = NewApiController(container.SheetRepository)

	container.Router = SetupRouter(container.ApiController)

	return
}
