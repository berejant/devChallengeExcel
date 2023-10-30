package contracts

import "errors"

type SheetRepository interface {
	SetCell(sheetId string, cellId string, value string) (*Cell, error)
	GetCell(sheetId string, cellId string) (*Cell, error)
	GetCellList(sheetId string) (*CellList, error)
}

var SheetNotFoundError = errors.New("sheet not found")
