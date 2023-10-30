package main

import (
	"devChallengeExcel/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expectedApiRoutes := [][3]string{
		{http.MethodPost, "/:sheet_id/:cell_id", "SetCellAction"},
		{http.MethodGet, "/:sheet_id/:cell_id", "GetCellAction"},
		{http.MethodGet, "/:sheet_id", "GetSheetAction"},
	}

	for _, expectedRoute := range expectedApiRoutes {
		t.Run("Route "+expectedRoute[2], func(t *testing.T) {
			apiController := mocks.NewApiController(t)
			router := SetupRouter(apiController)

			apiController.On(expectedRoute[2], mock.Anything).Return()

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(expectedRoute[0], "/api/"+ApiVersion+expectedRoute[1], nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			apiController.AssertNumberOfCalls(t, expectedRoute[2], 1)
		})
	}

	t.Run("healthcheck", func(t *testing.T) {
		apiController := mocks.NewApiController(t)
		router := SetupRouter(apiController)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/healthcheck", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "health", w.Body.String())
	})
}
