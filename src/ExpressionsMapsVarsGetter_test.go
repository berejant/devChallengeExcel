package main

import (
	"devChallengeExcel/contracts"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExpressionsMap_GetCellValues(t *testing.T) {
	getter := NewExpressionsMapsValuesGetter(&contracts.ExpressionsMap{
		"cell1": _makeStringRef("value1"),
		"cell2": _makeStringRef("value2"),
		"cell3": _makeStringRef("value3"),
	})

	actual := getter([]string{"cell3", "not-exists1", "cell1"})

	assert.Len(t, actual, 3)
	assert.Equal(t, "value3", *(actual[0]))
	assert.Nil(t, actual[1])
	assert.Equal(t, "value1", *(actual[2]))

}

func _makeStringRef(value string) *string {
	return &value
}
