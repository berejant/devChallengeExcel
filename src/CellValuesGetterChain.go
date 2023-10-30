package main

import "devChallengeExcel/contracts"

func NewCellValuesGetterChain(first contracts.CellValuesGetter, second contracts.CellValuesGetter) contracts.CellValuesGetter {
	if second == nil {
		return first
	}

	if first == nil {
		return second
	}

	return func(cellIds []string) []*string {
		result := first(cellIds)

		secondCellIds := make([]string, 0, len(cellIds))
		for index, value := range result {
			if value == nil {
				secondCellIds = append(secondCellIds, cellIds[index])
			}
		}

		if len(secondCellIds) != 0 {
			secondResult := second(secondCellIds)

			searchInSecondsCellIdsIndex := 0
			for index, value := range result {
				if value == nil {
					result[index] = secondResult[searchInSecondsCellIdsIndex]
					searchInSecondsCellIdsIndex++
				}
			}
		}

		return result
	}
}
