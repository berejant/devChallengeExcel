package main

import (
	"devChallengeExcel/contracts"
	"devChallengeExcel/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCellValuesGetterChain(t *testing.T) {
	t.Run("only_second", func(t *testing.T) {
		v := []string{"cell1", "cell2", "cell3"}
		second := mocks.NewCellValuesGetter(t)
		second.On("Execute", v).Return([]*string{nil, nil, nil})

		NewCellValuesGetterChain(nil, second.Execute)(v)
	})

	t.Run("only_first", func(t *testing.T) {
		v := []string{"cell1", "cell2", "cell3"}
		first := mocks.NewCellValuesGetter(t)
		first.On("Execute", v).Return([]*string{nil, nil, nil})

		NewCellValuesGetterChain(first.Execute, nil)(v)
	})

	t.Run("both", func(t *testing.T) {

		first := NewExpressionsMapsValuesGetter(&contracts.ExpressionsMap{
			"cell1": _makeStringRef("value1"),
			"cell2": _makeStringRef("value2"),
			"cell3": _makeStringRef("value3"),
		})

		second := NewExpressionsMapsValuesGetter(&contracts.ExpressionsMap{
			"cell25": _makeStringRef("value25"),
			"cell26": _makeStringRef("value26"),
			"cell27": _makeStringRef("value27"),
		})

		getter := NewCellValuesGetterChain(first, second)

		actual := getter([]string{"cell3", "not-exists1", "cell1", "cell26", "not-exists2", "cell25"})

		assert.Len(t, actual, 6)
		assert.Equal(t, "value3", *(actual[0]))
		assert.Nil(t, actual[1])
		assert.Equal(t, "value1", *(actual[2]))
		assert.Equal(t, "value26", *(actual[3]))
		assert.Nil(t, actual[4])
		assert.Equal(t, "value25", *(actual[5]))

	})
}