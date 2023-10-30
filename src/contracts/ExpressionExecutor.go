package contracts

type ExpressionExecutor interface {
	Evaluate(expression string, sheet CellValuesGetter) (string, error)
	MultiEvaluate(expressions ExpressionsMap, sheet CellValuesGetter, breakOnError bool) error
	ExtractDependingOnList(expression string) (dependingOnCellIds []string)
}
