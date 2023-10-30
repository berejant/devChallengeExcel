package contracts

import (
	"errors"
	"fmt"
	"strings"
)

type Cell struct {
	Value  string `json:"value"`
	Result string `json:"result"`
}

// CellIdBlacklist deny charset which associate with operators
const CellIdBlacklist = "+-*/%^()<>!=&|\t\n\r\v\f"

var CellNotFoundError = errors.New("cell not found")

var CellIdBlacklistError = fmt.Errorf("cell id contains invalid characters (%s)", strings.Join(strings.Split(CellIdBlacklist, ""), ", "))

var CellIdNumericError = errors.New("cell with numeric key should has numeric value")
