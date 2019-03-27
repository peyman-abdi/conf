package conf

import (
	"os"
	"strings"
)

type envEvaluator struct {
}

var _ EvaluatorFunction = (*envEvaluator)(nil)

func (e *envEvaluator) GetFunctionName() string {
	return "env"
}

func (e *envEvaluator) Eval(params []string, def interface{}) interface{} {
	if len(params) > 0 {
		envVal := os.Getenv(params[0])
		if envVal != "" {
			return envVal
		}

		if len(params) == 2 {
			return strings.Trim(params[1], " \"'")
		}

	}

	return def
}
