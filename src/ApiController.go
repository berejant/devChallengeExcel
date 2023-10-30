package main

import (
	"devChallengeExcel/contracts"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ApiController struct {
	SheetRepository contracts.SheetRepository
}

type CellEndpointParams struct {
	SheetId string `uri:"sheet_id" binding:"required"`
	CellId  string `uri:"cell_id" binding:"required"`
}

type SheetEndpointParams struct {
	SheetId string `uri:"sheet_id" binding:"required"`
}

type SetCellRequest struct {
	Value string `json:"value" binding:"required"`
}

// https://regex101.com/r/N5SLnV/2

func NewApiController(sheetRepository contracts.SheetRepository) *ApiController {
	return &ApiController{SheetRepository: sheetRepository}
}

func (api *ApiController) GetCellAction(c *gin.Context) {
	params := CellEndpointParams{}
	var response *contracts.Cell

	err := c.ShouldBindUri(&params)

	if err == nil {
		response, err = api.SheetRepository.GetCell(params.SheetId, params.CellId)
	}

	if errors.Is(err, contracts.CellNotFoundError) || errors.Is(err, contracts.SheetNotFoundError) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, response)
	}
}

func (api *ApiController) SetCellAction(c *gin.Context) {
	params := CellEndpointParams{}
	request := SetCellRequest{}
	var response *contracts.Cell

	err := c.ShouldBindUri(&params)
	if err == nil {
		err = c.ShouldBindJSON(&request)
	}

	if err == nil {
		response, err = api.SheetRepository.SetCell(params.SheetId, params.CellId, request.Value)
	}

	if err != nil {
		if response == nil {
			response = &contracts.Cell{}
		}
		response.Value = request.Value
		response.Result = err.Error()
		c.JSON(http.StatusUnprocessableEntity, response)
	} else {
		c.JSON(http.StatusCreated, response)
	}
}

func (api *ApiController) GetSheetAction(c *gin.Context) {
	params := SheetEndpointParams{}
	response := &contracts.CellList{}

	err := c.ShouldBindUri(&params)

	if err == nil {
		response, err = api.SheetRepository.GetCellList(params.SheetId)
	}

	if errors.Is(err, contracts.SheetNotFoundError) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})

	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, response)
	}
}
