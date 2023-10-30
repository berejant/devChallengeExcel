package main

import (
	"devChallengeExcel/contracts"
	"errors"
	"fmt"
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"strconv"
	"strings"
	"sync"
)

type ExpressionExecutor struct {
	canonicalizer   contracts.Canonicalizer
	compilerOptions []expr.Option
	vmPool          sync.Pool
}

const FormulaPrefix = "="

const FormulaExecutionInProcess = '\n'

var ExpressionError = errors.New("expression error")

var CircularReferenceError = fmt.Errorf("%w: %s", ExpressionError, "circular reference detected")

func NewExpressionExecutor(canonicalizer contracts.Canonicalizer) *ExpressionExecutor {
	return &ExpressionExecutor{
		canonicalizer: canonicalizer,
		compilerOptions: []expr.Option{
			expr.Env(map[string]any{}),
			expr.AllowUndefinedVariables(),
			expr.Optimize(false),
			expr.DisableAllBuiltins(),
		},

		vmPool: sync.Pool{
			New: func() any {
				return new(vm.VM)
			},
		},
	}
}

func (e *ExpressionExecutor) MultiEvaluate(expressions contracts.ExpressionsMap, sheetGetter contracts.CellValuesGetter, breakOnError bool) error {
	vars := make(map[string]any)
	var currentErr error
	var firstErr error

	cellValuesFromExpression := NewCellValuesGetterChain(NewExpressionsMapsValuesGetter(&expressions), sheetGetter)

	for cellId, expression := range expressions {
		currentErr = nil
		if e.IsFormula(*expression) {
			vars[cellId] = FormulaExecutionInProcess
			vars[cellId], currentErr = e.doEvaluate(*expression, cellValuesFromExpression, vars)
			*expressions[cellId] = e.outputToString(vars[cellId], currentErr)
		}

		if currentErr == nil && isNumeric(cellId) && !isNumeric(expression) {
			currentErr = contracts.CellIdNumericError
		}

		if firstErr == nil && currentErr != nil {
			firstErr = fmt.Errorf("cell %s: %w", cellId, currentErr)
			if breakOnError {
				break
			}
		}
	}

	return firstErr
}

func (e *ExpressionExecutor) Evaluate(expression string, sheet contracts.CellValuesGetter) (string, error) {
	// not formula
	if !e.IsFormula(expression) {
		return expression, nil
	}

	vars := make(map[string]any)
	output, err := e.doEvaluate(expression, sheet, vars)
	if err != nil {
		err = fmt.Errorf("%s: %w", expression, err)
	}
	return e.outputToString(output, err), err
}

func (e *ExpressionExecutor) ExtractDependingOnList(expression string) []string {
	dependants := make([]string, 0)
	// not formula
	if !e.IsFormula(expression) {
		return dependants
	}

	program, err := e.compile(expression)
	if err == nil {
		for _, constantValue := range program.Constants {
			if variableName, ok := constantValue.(string); ok {
				dependants = append(dependants, variableName)
			}
		}
	}

	return dependants
}

func (e *ExpressionExecutor) compile(expression string) (*vm.Program, error) {
	return expr.Compile(
		e.canonicalizer.Canonicalize(strings.TrimPrefix(expression, FormulaPrefix)),
		e.compilerOptions...,
	)
}

func (e *ExpressionExecutor) doEvaluate(expression string, sheet contracts.CellValuesGetter, vars map[string]any) (out any, err error) {
	program, err := e.compile(expression)
	if err != nil {
		return "", err
	}

	err = e.lookupAndFillVars(program, sheet, vars)
	if err != nil {
		return "", err
	}

	v := e.vmPool.Get().(*vm.VM)
	out, err = v.Run(program, vars)
	e.vmPool.Put(v)
	return
}

func (e *ExpressionExecutor) IsFormula(expression string) bool {
	return strings.HasPrefix(expression, FormulaPrefix)
}

/**
 * Retrieve variable which are used in expression and still not filled
 * @param program
 */
func (e *ExpressionExecutor) lookupAndFillVars(program *vm.Program, valuesGetter contracts.CellValuesGetter, vars map[string]any) error {
	variablesNamesToFetch := make([]string, 0, len(program.Constants))
	constantIndexes := make([]int, 0, len(program.Constants))

	var ok bool
	var constantIndex int
	var constantValue any
	var variableName string
	var index int

	for constantIndex, constantValue = range program.Constants {
		variableName = e.toString(constantValue)
		if _, ok = vars[variableName]; !ok {
			variablesNamesToFetch = append(variablesNamesToFetch, variableName)
			constantIndexes = append(constantIndexes, constantIndex)
		} else if vars[variableName] == FormulaExecutionInProcess {
			return fmt.Errorf("%s: %w", variableName, CircularReferenceError)
		} else {
			e.overrideNumberConstant(program, constantIndex, vars[variableName])
		}
	}

	if len(variablesNamesToFetch) == 0 {
		return nil
	}
	fetchedValues := valuesGetter(variablesNamesToFetch)

	var stringValueRef *string
	var floatValue float64
	var intValue int64
	var err error
	for index, variableName = range variablesNamesToFetch {
		constantIndex = constantIndexes[index]
		stringValueRef = fetchedValues[index]
		if stringValueRef == nil {
			continue

		} else if e.IsFormula(*stringValueRef) {
			// prevent recursive call - mark this variable as in process
			vars[variableName] = FormulaExecutionInProcess
			vars[variableName], err = e.doEvaluate(*stringValueRef, valuesGetter, vars)
			if err != nil {
				return err
			}

		} else if intValue, err = strconv.ParseInt(*stringValueRef, 10, 64); err == nil {
			vars[variableName] = intValue
		} else if floatValue, err = strconv.ParseFloat(*stringValueRef, 64); err == nil {
			vars[variableName] = floatValue
		} else {
			vars[variableName] = *stringValueRef
		}

		e.overrideNumberConstant(program, constantIndex, vars[variableName])
	}

	return nil
}

// overrideNumberConstant According to judge comments, we should override number constants in case when there is `cell_id` with same digit value
func (e *ExpressionExecutor) overrideNumberConstant(program *vm.Program, constantIndex int, overrideValue any) {
	switch (program.Constants[constantIndex]).(type) {
	case int, int64, float64:
		program.Constants[constantIndex] = overrideValue
	}
}

func (e *ExpressionExecutor) outputToString(output any, err error) string {
	if err != nil {
		return "ERROR: " + err.Error()
	}

	return e.toString(output)
}

func (e *ExpressionExecutor) toString(input any) string {
	switch input.(type) {
	case int:
		return strconv.Itoa(input.(int))
	case float64:
		return strconv.FormatFloat(input.(float64), 'f', -1, 64)
	case string:
		return input.(string)
	default:
		return ""
	}
}

func isNumeric(input any) bool {

	switch (input).(type) {
	case string:
		var err error
		if _, err = strconv.ParseInt(input.(string), 10, 64); err == nil {
			return true
		} else if _, err = strconv.ParseFloat(input.(string), 64); err == nil {
			return true
		}

	case *string:
		return isNumeric(*input.(*string))

	case int, float64, *int, *float64:
		return true
	}

	return false
}
