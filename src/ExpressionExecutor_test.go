package main

import (
	"devChallengeExcel/contracts"
	"devChallengeExcel/mocks"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestExpressionExecutor_Evaluate(t *testing.T) {
	t.Run("simple_expression", func(t *testing.T) {
		t.Run("non_formula_expression", func(t *testing.T) {
			executor := NewExpressionExecutor(NewCanonicalizer())
			actual, err := executor.Evaluate("5", nil)

			assert.NoError(t, err)
			assert.Equal(t, "5", actual)

			actual, err = executor.Evaluate("awesome", nil)
			assert.NoError(t, err)
			assert.Equal(t, "awesome", actual)

		})

		t.Run("simple_formula", func(t *testing.T) {
			valuesGetter := mocks.NewCellValuesGetter(t)
			valuesGetter.On("Execute", []string{"1", "2"}).Return([]*string{nil, nil})

			executor := NewExpressionExecutor(NewCanonicalizer())
			actual, err := executor.Evaluate("=1+2", valuesGetter.Execute)

			assert.NoError(t, err)
			assert.Equal(t, "3", actual)
		})

		t.Run("simple_formula_variables", func(t *testing.T) {
			getValuesNames := []string{"a1", "a2"}
			getValues := []*string{
				_makeStringRef("110"),
				_makeStringRef("20.5"),
			}

			valuesGetter := mocks.NewCellValuesGetter(t)
			valuesGetter.On("Execute", getValuesNames).Return(getValues)

			executor := NewExpressionExecutor(NewCanonicalizer())
			actual, err := executor.Evaluate("=A1+A2", valuesGetter.Execute)

			assert.NoError(t, err)
			assert.Equal(t, "130.5", actual)
		})

	})

	t.Run("recursive_formula", func(t *testing.T) {
		t.Run("deep_with_two_recursive", func(t *testing.T) {
			varsGetting := map[*[]string][]*string{
				&[]string{"a1", "a2"}: {
					_makeStringRef("= A10 - A20"),
					_makeStringRef("=A30*A31"),
				},

				&[]string{"a10", "a20"}: {
					_makeStringRef("-10"),
					_makeStringRef("50"),
				},

				&[]string{"a30", "a31"}: {
					_makeStringRef("2"),
					_makeStringRef("3"),
				},
			}
			valuesGetter := mocks.NewCellValuesGetter(t)
			for getValuesNames, getValues := range varsGetting {
				valuesGetter.On("Execute", *getValuesNames).Return(getValues)
			}

			executor := NewExpressionExecutor(NewCanonicalizer())
			// =A1+A2 = (A10-A20) + (A30*A31) = (-10-50) + (2*3) = -60 + 6 = -54
			actual, err := executor.Evaluate("=    A1   +    A2    ", valuesGetter.Execute)

			assert.NoError(t, err)
			assert.Equal(t, "-54", actual)
		})

		t.Run("circular", func(t *testing.T) {
			varsGetting := map[*[]string][]*string{
				&[]string{"a1", "a2"}: {
					_makeStringRef("= A10 - A20"),
					_makeStringRef("=A30*A31"),
				},

				&[]string{"a10", "a20"}: {
					_makeStringRef("-10"),
					_makeStringRef("=A1"),
				},
			}
			valuesGetter := mocks.NewCellValuesGetter(t)
			for getValuesNames, getValues := range varsGetting {
				valuesGetter.On("Execute", *getValuesNames).Return(getValues)
			}

			executor := NewExpressionExecutor(NewCanonicalizer())
			// =A1+A2 = (A10-A20) + (A30*A31) = (-10-50) + (2*3) = -60 + 6 = -54
			actual, err := executor.Evaluate("=    A1   +    A2    ", valuesGetter.Execute)

			assert.Error(t, err)
			assert.ErrorIs(t, err, CircularReferenceError)
			assert.True(t, strings.HasPrefix(actual, "ERROR: "))
		})
	})

	t.Run("override_numbers", func(t *testing.T) {
		t.Run("simple", func(t *testing.T) {
			getValuesNames := []string{"1", "2.2", "10", "a3"}
			getValues := []*string{
				_makeStringRef("40.5"),
				_makeStringRef("6"),
				nil,
				_makeStringRef("5.5"),
			}

			valuesGetter := mocks.NewCellValuesGetter(t)
			valuesGetter.On("Execute", getValuesNames).Return(getValues)

			executor := NewExpressionExecutor(NewCanonicalizer())
			actual, err := executor.Evaluate("=1+2.2+10+A3", valuesGetter.Execute)

			assert.NoError(t, err)
			assert.Equal(t, "62", actual)
		})

		t.Run("recursive", func(t *testing.T) {
			getValuesNames := []string{"1", "2.2", "10", "a3"}
			getValues := []*string{
				_makeStringRef("40.5"),
				_makeStringRef("6"),
				nil,
				_makeStringRef("=A4"),
			}

			valuesGetter := mocks.NewCellValuesGetter(t)
			valuesGetter.On("Execute", getValuesNames).Return(getValues)

			valuesGetter.On("Execute", []string{"a4"}).Return([]*string{
				_makeStringRef("=" + getValuesNames[0]),
			})

			executor := NewExpressionExecutor(NewCanonicalizer())
			// =1+2.2+10+A3 = 40.5 + 6 + 10 + 40.5 = 97
			actual, err := executor.Evaluate("=1+2.2+10+A3", valuesGetter.Execute)

			assert.NoError(t, err)
			assert.Equal(t, "97", actual)
		})

		t.Run("recursive_digit_value", func(t *testing.T) {
			// cell have digit value which match with cell name - do not override it
			getValuesNames := []string{"1", "2.2", "10", "a3"}
			getValues := []*string{
				_makeStringRef("40.5"),
				_makeStringRef("6"),
				nil,
				_makeStringRef("=A4"),
			}

			valuesGetter := mocks.NewCellValuesGetter(t)
			valuesGetter.On("Execute", getValuesNames).Return(getValues)

			valuesGetter.On("Execute", []string{"a4"}).Return([]*string{
				&getValuesNames[0],
			})

			executor := NewExpressionExecutor(NewCanonicalizer())
			// =1+2.2+10+A3 = 40.5 + 6 + 10 + 1 = 57.5
			actual, err := executor.Evaluate("=1+2.2+10+A3", valuesGetter.Execute)

			assert.NoError(t, err)
			assert.Equal(t, "57.5", actual)
		})
	})

	t.Run("errors", func(t *testing.T) {
		t.Run("compile_errors", func(t *testing.T) {
			executor := NewExpressionExecutor(NewCanonicalizer())
			actual, err := executor.Evaluate("=(value1+value2", nil)

			assert.Error(t, err)
			assert.True(t, strings.HasPrefix(actual, "ERROR: "))
			assert.True(t, strings.Contains(actual, err.Error()))
		})

		t.Run("runtime_error", func(t *testing.T) {
			getValuesNames := []string{"a1", "a2"}
			getValues := []*string{
				_makeStringRef("110"),
				_makeStringRef("string"),
			}

			valuesGetter := mocks.NewCellValuesGetter(t)
			valuesGetter.On("Execute", getValuesNames).Return(getValues)

			executor := NewExpressionExecutor(NewCanonicalizer())
			actual, err := executor.Evaluate("=A1+A2", valuesGetter.Execute)

			assert.Error(t, err)
			assert.True(t, strings.HasPrefix(actual, "ERROR: "))
			assert.True(t, strings.Contains(actual, err.Error()))
		})
	})
}

func TestExpressionExecutor_MultiEvaluate(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		expressions := contracts.ExpressionsMap{
			"h10": _makeStringRef("5"),
			"h11": _makeStringRef("awesome"),
			//			"H20": _makeStringRef("=NOT_EXISTING + 1"),
			"h12": _makeStringRef("=1+2"),
			"h13": _makeStringRef("=H10+H12"),
		}

		executor := NewExpressionExecutor(NewCanonicalizer())
		err := executor.MultiEvaluate(expressions, nil, false)

		assert.NoError(t, err)

		assert.Equal(t, "5", *expressions["h10"])
		assert.Equal(t, "awesome", *expressions["h11"])
		assert.Equal(t, "3", *expressions["h12"])
		assert.Equal(t, "8", *expressions["h13"])
	})

	t.Run("errors", func(t *testing.T) {
		expressions := contracts.ExpressionsMap{
			"h10": _makeStringRef("5"),
			"h11": _makeStringRef("awesome"),
			"h20": _makeStringRef("=NOT_EXISTING + 1"),
			"h12": _makeStringRef("=1+2"),
			"h13": _makeStringRef("=H10+H11"),
		}

		executor := NewExpressionExecutor(NewCanonicalizer())
		err := executor.MultiEvaluate(expressions, nil, false)

		assert.Error(t, err)
		assert.True(t, strings.HasPrefix(err.Error(), "cell "), err.Error())

		assert.Equal(t, "5", *expressions["h10"])
		assert.Equal(t, "awesome", *expressions["h11"])
		assert.Equal(t, "3", *expressions["h12"])

		assert.True(t, strings.HasPrefix(*expressions["h20"], "ERROR: "))
		assert.True(t, strings.HasPrefix(*expressions["h13"], "ERROR: "))
	})

	t.Run("break_on_first_error", func(t *testing.T) {
		expressions := contracts.ExpressionsMap{
			"a1":  _makeStringRef("=1+2"),
			"h10": _makeStringRef("5"),
			"h11": _makeStringRef("awesome"),
			"h20": _makeStringRef("=NOT_EXISTING + 1"),
			"h13": _makeStringRef("=H10+H11"),
			"h12": _makeStringRef("=1+2"),

			"hhhh220": _makeStringRef("=1+2"),
		}

		executor := NewExpressionExecutor(NewCanonicalizer())
		err := executor.MultiEvaluate(expressions, nil, true)

		assert.Error(t, err)
		assert.True(t, strings.HasPrefix(err.Error(), "cell "), err.Error())

		assert.Equal(t, "5", *expressions["h10"])
		assert.Equal(t, "awesome", *expressions["h11"])
		// last expression is not executed
		atLeastOneNotExecuted := *expressions["h13"] == "=H10+H11" || *expressions["hhhh220"] == "=1+2" || *expressions["a1"] == "=1+2"
		assert.True(t, atLeastOneNotExecuted)

		assert.True(t, strings.HasPrefix(*expressions["h20"], "ERROR: ") || strings.HasPrefix(*expressions["h13"], "ERROR: "))
	})

	t.Run("numeric_cell_id", func(t *testing.T) {
		t.Run("string_value", func(t *testing.T) {
			expressions := contracts.ExpressionsMap{
				"123": _makeStringRef("awesome123"),
			}

			executor := NewExpressionExecutor(NewCanonicalizer())
			err := executor.MultiEvaluate(expressions, nil, true)

			assert.Error(t, err)
			assert.True(t, strings.HasPrefix(err.Error(), "cell 123"), err.Error())
			assert.ErrorIs(t, err, contracts.CellIdNumericError)
		})

		t.Run("string_in_referenced", func(t *testing.T) {
			expressions := contracts.ExpressionsMap{
				"123": _makeStringRef("=A1"),
				"a1":  _makeStringRef("=awesome"),
			}

			executor := NewExpressionExecutor(NewCanonicalizer())
			err := executor.MultiEvaluate(expressions, nil, true)

			assert.Error(t, err)
			assert.True(t, strings.HasPrefix(err.Error(), "cell 123"), err.Error())
			assert.ErrorIs(t, err, contracts.CellIdNumericError)
		})

		t.Run("correct_numeric_value_by_ref", func(t *testing.T) {
			expressions := contracts.ExpressionsMap{
				"123": _makeStringRef("=A1"),
				"a1":  _makeStringRef("=1234"),
			}

			executor := NewExpressionExecutor(NewCanonicalizer())
			err := executor.MultiEvaluate(expressions, nil, true)

			assert.NoError(t, err)
			assert.Equal(t, "1234", *expressions["123"])
		})

		t.Run("correct_float_value", func(t *testing.T) {
			expressions := contracts.ExpressionsMap{
				"123": _makeStringRef("=123.45"),
			}

			executor := NewExpressionExecutor(NewCanonicalizer())
			err := executor.MultiEvaluate(expressions, nil, true)

			assert.NoError(t, err)
			assert.Equal(t, "123.45", *expressions["123"])
		})

		t.Run("correct_float_value_by_ref", func(t *testing.T) {
			expressions := contracts.ExpressionsMap{
				"123": _makeStringRef("=A1"),
				"a1":  _makeStringRef("=123.34"),
			}

			executor := NewExpressionExecutor(NewCanonicalizer())
			err := executor.MultiEvaluate(expressions, nil, true)

			assert.NoError(t, err)
			assert.Equal(t, "123.34", *expressions["123"])
		})
	})
}

func TestExpressionExecutor_outputToString(t *testing.T) {
	executor := NewExpressionExecutor(NewCanonicalizer())

	assert.Equal(t, "", executor.outputToString(nil, nil))
	assert.Equal(t, "text", executor.outputToString("text", nil))
	assert.Equal(t, "5", executor.outputToString(5, nil))
	assert.Equal(t, "5.5", executor.outputToString(5.5, nil))
}

func TestExpressionExecutor_ExtractDependingOnList(t *testing.T) {
	executor := NewExpressionExecutor(NewCanonicalizer())

	assert.Equal(t, []string{"a1", "a2"}, executor.ExtractDependingOnList("=A1+A2"))
	assert.Equal(t, []string{"a1", "a2"}, executor.ExtractDependingOnList("=4+A1+A2+3"))

	assert.Equal(t, []string{}, executor.ExtractDependingOnList("=4"))

	// not formula
	assert.Equal(t, []string{}, executor.ExtractDependingOnList("4"))
	assert.Equal(t, []string{}, executor.ExtractDependingOnList("test1"))

	// compile error
	assert.Equal(t, []string{}, executor.ExtractDependingOnList("=(4+(4 * (1"))

}

func TestIsNumeric(t *testing.T) {
	assert.True(t, isNumeric(_makeStringRef("123")))
	assert.True(t, isNumeric("123"))
	assert.True(t, isNumeric("123e23"))

	assert.True(t, isNumeric(_makeStringRef("123.23")))
	assert.True(t, isNumeric("123.23"))
	assert.True(t, isNumeric(_makeStringRef("123.23e2")))

	assert.False(t, isNumeric(_makeStringRef("123.23e")))
	assert.False(t, isNumeric("123.23e"))

	assert.True(t, isNumeric(123.56))
	assert.True(t, isNumeric(123))

	int1 := 999
	float2 := 999.999
	assert.True(t, isNumeric(&int1))
	assert.True(t, isNumeric(&float2))
}
