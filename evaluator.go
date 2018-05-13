package conf

type EvaluatorFunction interface {
	Eval(params []string, def interface{}) interface{}
	GetFunctionName() string
}
