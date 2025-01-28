package main

import (
	"bytes"
	"devChallengeExcel/contracts"
	"devChallengeExcel/mocks"
	"errors"
	json "github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApiController_GetCellAction(t *testing.T) {
	gin.SetMode(gin.TestMode)

	requestToGetCellAction := func(apiController contracts.ApiController) *httptest.ResponseRecorder {
		router := SetupRouter(apiController)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/"+ApiVersion+"/sheet1/cell1", nil)
		router.ServeHTTP(w, req)
		return w
	}

	t.Run("should return cell value", func(t *testing.T) {
		sheetRepository := mocks.NewSheetRepository(t)
		sheetRepository.On("GetCell", "sheet1", "cell1").
			Return(&contracts.Cell{
				Value:  "value1",
				Result: "value1",
			}, nil)

		apiController := NewApiController(sheetRepository, nil, nil)

		w := requestToGetCellAction(apiController)
		response, err := _parseJsonBody(w)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, response, "value")
		assert.Contains(t, response, "result")
		assert.Equal(t, response["value"], "value1")
		assert.Equal(t, response["result"], "value1")
	})

	t.Run("cell not found", func(t *testing.T) {
		sheetRepository := mocks.NewSheetRepository(t)
		sheetRepository.On("GetCell", "sheet1", "cell1").Return(nil, contracts.CellNotFoundError)

		apiController := NewApiController(sheetRepository, nil, nil)

		w := requestToGetCellAction(apiController)
		response, err := _parseJsonBody(w)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, response, "error")
		assert.Equal(t, response["error"], contracts.CellNotFoundError.Error())
	})

	t.Run("sheet not found", func(t *testing.T) {
		sheetRepository := mocks.NewSheetRepository(t)
		sheetRepository.On("GetCell", "sheet1", "cell1").Return(nil, contracts.SheetNotFoundError)

		apiController := NewApiController(sheetRepository, nil, nil)

		w := requestToGetCellAction(apiController)
		response, err := _parseJsonBody(w)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, response, "error")
		assert.Equal(t, response["error"], contracts.SheetNotFoundError.Error())
	})

	t.Run("custom error", func(t *testing.T) {
		sheetRepository := mocks.NewSheetRepository(t)
		sheetRepository.On("GetCell", "sheet1", "cell1").Return(nil, errors.New("test"))

		apiController := NewApiController(sheetRepository, nil, nil)

		w := requestToGetCellAction(apiController)

		response, err := _parseJsonBody(w)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, response, "error")
		assert.Equal(t, response["error"], "test")
	})
}

func TestApiController_SetCellAction(t *testing.T) {
	gin.SetMode(gin.TestMode)

	requestToSetCellAction := func(apiController contracts.ApiController, data map[string]string) *httptest.ResponseRecorder {
		jsonBody, _ := json.Marshal(data)
		bodyReader := bytes.NewReader(jsonBody)

		router := SetupRouter(apiController)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/"+ApiVersion+"/sheet1/cell1", bodyReader)
		router.ServeHTTP(w, req)
		return w
	}

	t.Run("success write", func(t *testing.T) {
		sheetRepository := mocks.NewSheetRepository(t)
		sheetRepository.On("SetCell", "sheet1", "cell1", "value1", true).
			Return(&contracts.Cell{Value: "value1"}, nil, false)

		apiController := NewApiController(sheetRepository, nil, nil)

		w := requestToSetCellAction(apiController, map[string]string{"value": "value1"})
		response, err := _parseJsonBody(w)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, response, "value")
		assert.Equal(t, response["value"], "value1")
	})

	t.Run("error", func(t *testing.T) {
		sheetRepository := mocks.NewSheetRepository(t)
		sheetRepository.On("SetCell", "sheet1", "cell1", "value1", true).
			Return(nil, errors.New("test"), false)

		apiController := NewApiController(sheetRepository, nil, nil)

		w := requestToSetCellAction(apiController, map[string]string{"value": "value1"})
		response, err := _parseJsonBody(w)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assert.Contains(t, response, "result")
		assert.Contains(t, response, "value")
		assert.Equal(t, response["value"], "value1")
		assert.Equal(t, response["result"], "test")
	})

}

func TestApiController_GetSheetAction(t *testing.T) {
	gin.SetMode(gin.TestMode)

	requestToGetSheetAction := func(apiController contracts.ApiController) *httptest.ResponseRecorder {
		router := SetupRouter(apiController)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/"+ApiVersion+"/sheet1", nil)
		router.ServeHTTP(w, req)
		return w
	}

	t.Run("success", func(t *testing.T) {
		list := &contracts.CellList{
			"cell1": {
				Value:  "value1",
				Result: "value1",
			},
			"cell2": {
				Value:  "value2",
				Result: "value2",
			},
		}

		sheetRepository := mocks.NewSheetRepository(t)
		sheetRepository.On("GetCellList", "sheet1").Return(list, nil)

		apiController := NewApiController(sheetRepository, nil, nil)

		w := requestToGetSheetAction(apiController)
		response, err := _parseJsonBody(w)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		assert.Contains(t, response, "cell1")
		assert.Contains(t, response, "cell2")

		for key, cell := range *list {
			assert.Contains(t, response, key)

			responseCell := response[key].(map[string]any)
			assert.Equal(t, responseCell["value"], cell.Value)
			assert.Equal(t, responseCell["result"], cell.Result)
		}
	})

	t.Run("not_found_sheet", func(t *testing.T) {
		sheetRepository := mocks.NewSheetRepository(t)
		sheetRepository.On("GetCellList", "sheet1").Return(nil, contracts.SheetNotFoundError)

		apiController := NewApiController(sheetRepository, nil, nil)

		w := requestToGetSheetAction(apiController)
		response, err := _parseJsonBody(w)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)

		assert.Contains(t, response, "error")
		assert.Equal(t, response["error"], contracts.SheetNotFoundError.Error())
	})

	t.Run("error", func(t *testing.T) {
		sheetRepository := mocks.NewSheetRepository(t)
		sheetRepository.On("GetCellList", "sheet1").Return(nil, errors.New("test"))

		apiController := NewApiController(sheetRepository, nil, nil)

		w := requestToGetSheetAction(apiController)
		response, err := _parseJsonBody(w)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		assert.Contains(t, response, "error")
		assert.Equal(t, response["error"], "test")
	})
}

func _parseJsonBody(w *httptest.ResponseRecorder) (response map[string]any, err error) {
	err = json.Unmarshal(w.Body.Bytes(), &response)
	return
}
