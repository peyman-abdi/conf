package conf

import (
	"strings"
	"os"
	"fmt"
)

type envEvaluator struct {
}
var _ EvaluatorFunction = (*envEvaluator)(nil)

func (_ *envEvaluator) GetFunctionName() string {
	return "env"
}

func (_ *envEvaluator) Eval(params []string, def interface{}) interface{}  {
	if len(params) == 2 {
		envVal := os.Getenv(params[0])
		if envVal != "" {
			fmt.Println(envVal)
			return strings.Trim(envVal, " \"'")
		} else {
			return strings.Trim(params[1], " \"'")
		}
	} else {
		envVal := os.Getenv(params[0])
		if envVal != "" {
			return envVal
		} else {
			return def
		}
	}
}
