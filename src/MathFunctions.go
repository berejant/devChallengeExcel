package main

import (
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm/runtime"
)

var calculateMax = func(args ...any) (any, error) {
	var maxValue any
	for _, arg := range args {
		if maxValue == nil || runtime.Less(maxValue, arg) {
			maxValue = arg
		}
	}
	return maxValue, nil
}

var calculateMin = func(args ...any) (any, error) {
	var minValue any
	for _, arg := range args {
		if minValue == nil || runtime.More(minValue, arg) {
			minValue = arg
		}
	}
	return minValue, nil
}

var calculateSum = func(args ...any) (any, error) {
	sum := args[0]
	for i := 1; i < len(args); i++ {
		sum = runtime.Add(sum, args[i])
	}
	return sum, nil
}

var calculateAvg = func(args ...any) (any, error) {
	sum, err := calculateSum(args...)
	if err != nil {
		return nil, err
	}
	return runtime.Divide(sum, len(args)), nil
}

var maxFunction = expr.Function("max", calculateMax)
var minFunction = expr.Function("min", calculateMin)
var sumFunction = expr.Function("sum", calculateSum)
var avgFunction = expr.Function("avg", calculateAvg)
