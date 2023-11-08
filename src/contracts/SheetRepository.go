package contracts

import "errors"

type SheetRepository interface {
	SetCell(sheetId string, cellId string, value string, skipNotChanged bool) (*Cell, error, bool)
	GetCell(sheetId string, cellId string) (*Cell, error)
	GetCellList(sheetId string) (*CellList, error)
	GetCanonicalSheetId(sheetId string) string
}

var SheetNotFoundError = errors.New("sheet not found")
