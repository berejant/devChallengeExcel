package main

import (
	"bytes"
	"devChallengeExcel/contracts"
	"errors"
	"fmt"
	json "github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strings"
	"time"
)

type ApiController struct {
	SheetRepository   contracts.SheetRepository
	WebhookDispatcher contracts.WebhookDispatcher
	Executor          contracts.ExpressionExecutor
	Hostname          string
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

type WebhookConfig struct {
	WebhookUrl string `json:"webhook_url" binding:"required"`
}

// https://regex101.com/r/N5SLnV/2

func NewApiController(sheetRepository contracts.SheetRepository, webhookDispatcher contracts.WebhookDispatcher, executor contracts.ExpressionExecutor) *ApiController {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	return &ApiController{
		SheetRepository:   sheetRepository,
		WebhookDispatcher: webhookDispatcher,
		Executor:          executor,
		Hostname:          hostname + ListenPort,
	}
}

func (api *ApiController) GetCellAction(c *gin.Context) {
	params := CellEndpointParams{}
	var response *contracts.Cell

	err := c.ShouldBindUri(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err = api.SheetRepository.GetCell(params.SheetId, params.CellId)

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
	var isUpdated bool

	err := c.ShouldBindUri(&params)
	if err == nil {
		err = c.ShouldBindJSON(&request)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err, isUpdated = api.SheetRepository.SetCell(params.SheetId, params.CellId, request.Value, true)

	if isUpdated {
		go api.SubscribeExternalRefsToWebhook(&params, response)
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
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err = api.SheetRepository.GetCellList(params.SheetId)

	if errors.Is(err, contracts.SheetNotFoundError) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})

	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, response)
	}
}

func (api *ApiController) SubscribeAction(c *gin.Context) {
	params := CellEndpointParams{}
	webhookRequestConfig := WebhookConfig{}

	err := c.ShouldBindUri(&params)
	if err == nil {
		err = c.ShouldBindJSON(&webhookRequestConfig)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var cell *contracts.Cell
	cell, err = api.SheetRepository.GetCell(params.SheetId, params.CellId)
	if errors.Is(err, contracts.CellNotFoundError) || errors.Is(err, contracts.SheetNotFoundError) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	canonicalSheetId := api.SheetRepository.GetCanonicalSheetId(params.SheetId)
	api.WebhookDispatcher.SetWebhookUrl(canonicalSheetId, cell.CanonicalKey, webhookRequestConfig.WebhookUrl)

	webhookResponseConfig := WebhookConfig{
		WebhookUrl: api.WebhookDispatcher.GetWebhookUrl(canonicalSheetId, cell.CanonicalKey),
	}

	c.JSON(http.StatusCreated, webhookResponseConfig)
}

func (api *ApiController) SubscribeExternalRefsToWebhook(params *CellEndpointParams, cell *contracts.Cell) {
	// parse external refs
	externalsRefs := api.Executor.ExtractExternalRefs(cell.Value)
	if externalsRefs == nil || len(externalsRefs) == 0 {
		return
	}

	webhookUrl := "http://" + api.Hostname + "/api/" + ApiVersion + "/" + params.SheetId + "/" + params.CellId + "/" + externalRefWebhookPath

	webhookConfig := WebhookConfig{
		WebhookUrl: webhookUrl,
	}
	payload, _ := json.Marshal(webhookConfig)

	client := http.Client{
		Timeout: time.Second * 4,
	}
	for _, externalRef := range externalsRefs {
		externalRefSubscribeEndpoint := strings.TrimSuffix(externalRef, "/") + "/" + subscribePath
		response, err := client.Post(externalRefSubscribeEndpoint, "application/json", bytes.NewReader(payload))
		if err != nil {
			fmt.Println("failed to subscribe:", err)
		} else if response.StatusCode != http.StatusCreated {
			fmt.Println("failed to create subscribe:", response.Status)
		} else {
			fmt.Printf("subscribed to %s (webhook %s)\n", externalRefSubscribeEndpoint, webhookUrl)
		}
	}

}

func (api *ApiController) ExternalRefWebhookAction(c *gin.Context) {
	params := CellEndpointParams{}
	var response *contracts.Cell

	err := c.ShouldBindUri(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cell, _ := api.SheetRepository.GetCell(params.SheetId, params.CellId)

	response, err, _ = api.SheetRepository.SetCell(params.SheetId, params.CellId, cell.Value, false)

	if errors.Is(err, contracts.CellNotFoundError) || errors.Is(err, contracts.SheetNotFoundError) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, response)
	}
}
