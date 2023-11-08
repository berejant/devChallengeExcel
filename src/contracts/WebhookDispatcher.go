package contracts

type WebhookDispatcher interface {
	SetWebhookUrl(canonicalSheetId string, canonicalCellId string, webhookUrl string)
	GetWebhookUrl(canonicalSheetId string, canonicalCellId string) string
	Notify(canonicalSheetId string, cells []*Cell)
	Start()
	Close()
}
