package main

import (
	"bytes"
	"devChallengeExcel/contracts"
	"fmt"
	json "github.com/bytedance/sonic"
	"net/http"
	"time"
)

const WebhookWorkersCount = 5

type SheetWebhooks map[string]string

type WebhookSendCommand struct {
	Webhook string
	Cell    *contracts.Cell
}

type WebhookDispatcher struct {
	queue    chan WebhookSendCommand
	webhooks map[string]SheetWebhooks
}

func NewWebhookDispatcher() *WebhookDispatcher {
	return &WebhookDispatcher{
		queue:    make(chan WebhookSendCommand, 20),
		webhooks: map[string]SheetWebhooks{},
	}
}

func (manager *WebhookDispatcher) SetWebhookUrl(canonicalSheetId string, canonicalCellId string, webhookUrl string) {
	if _, ok := manager.webhooks[canonicalSheetId]; !ok {
		manager.webhooks[canonicalSheetId] = SheetWebhooks{}
	}

	if webhookUrl == "" {
		delete(manager.webhooks[canonicalSheetId], canonicalCellId)
	} else {
		manager.webhooks[canonicalSheetId][canonicalCellId] = webhookUrl
	}
}

func (manager *WebhookDispatcher) GetWebhookUrl(canonicalSheetId string, canonicalCellId string) string {
	if _, ok := manager.webhooks[canonicalSheetId]; !ok {
		return ""
	}

	if webhook, ok := manager.webhooks[canonicalSheetId][canonicalCellId]; ok {
		return webhook
	}

	return ""
}

func (manager *WebhookDispatcher) Notify(canonicalSheetId string, cells []*contracts.Cell) {
	if _, ok := manager.webhooks[canonicalSheetId]; !ok {
		return
	}

	go manager.addToQueue(canonicalSheetId, cells)
}

func (manager *WebhookDispatcher) addToQueue(canonicalSheetId string, cells []*contracts.Cell) {
	var ok bool
	if _, ok = manager.webhooks[canonicalSheetId]; ok {
		var webhook string
		for _, cell := range cells {
			if webhook, ok = manager.webhooks[canonicalSheetId][cell.CanonicalKey]; ok {
				manager.queue <- WebhookSendCommand{
					Webhook: webhook,
					Cell:    cell,
				}
			}
		}
	}
}

func (manager *WebhookDispatcher) Start() {
	for i := 0; i < WebhookWorkersCount; i++ {
		go manager.runWebhookSenderWorker()
	}
}

func (manager *WebhookDispatcher) Close() {
	close(manager.queue)
}

func (manager *WebhookDispatcher) runWebhookSenderWorker() {
	client := &http.Client{
		Timeout: time.Second * 5,
	}

	var response *http.Response
	var err error

	for command := range manager.queue {
		payload, _ := json.Marshal(command.Cell)
		response, err = client.Post(command.Webhook, "application/json", bytes.NewBuffer(payload))

		if err != nil {
			fmt.Printf("Webhook send error: %s\n", err)
		} else if response.StatusCode >= 300 {
			fmt.Printf("Unexpect wbhook response HTTP status: %s\n", response.Status)
		}
	}
}
