package main

import "devChallengeExcel/contracts"

func NewExpressionsMapsValuesGetter(vars *contracts.ExpressionsMap) contracts.CellValuesGetter {
	return func(cellIds []string) []*string {
		values := make([]*string, len(cellIds))

		var ok bool
		var value *string
		for index, cellId := range cellIds {
			if value, ok = (*vars)[cellId]; ok {
				values[index] = value
			}
		}

		return values
	}
}
