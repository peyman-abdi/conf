package nemo

import (
	"testing"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type testEvalFunction struct {
}
var _ EvaluatorFunction = (*testEvalFunction)(nil)

func (_ *testEvalFunction) GetFunctionName() string {
	return "paramsJoin"
}
func (_ *testEvalFunction) Eval(params []string, def interface{}) interface{} {
	return strings.Join(params, ":")
}

func TestInit(t *testing.T) {
	root, err := os.Executable()
	if err != nil {
		panic(err)
	}

	rootDir := filepath.Dir(root + "/../../")

	New(rootDir + "/configs", rootDir, []EvaluatorFunction {
		new(testEvalFunction),
	})
}

func TestNestedObjects(t *testing.T) {
	root, err := os.Executable()
	if err != nil {
		panic(err)
	}

	rootDir := filepath.Dir(root + "/../../")

	config := New(rootDir + "/configs", rootDir, []EvaluatorFunction {
		new(testEvalFunction),
	})

	testStrings := map[string]string {
		"nested.objects[0].name": "First",
		"nested.objects[1].name": "Second",
		"nested.objects[0].role": "Object",
		"nested.objects[1].role": "Object",
		"nested.vars.app.inner.string": "Some string goes here",
		"nested.vars.app.array[1]": "string element 1",
	}
	for key, val := range testStrings {
		checkString(config, key, val, t)
	}

	testFloats := map[string]float64 {
		"nested.vars.app.inner.float": 13.333,
		"nested.objects[0].float": 103.33,
		"nested.objects[1].float": 203.33,
	}
	for key, val := range testFloats {
		if result := config.GetFloat(key, 0); result != val {
			t.Error( fmt.Sprintf("Looking for \"%f\" in key %s but found %f", val, key, result))
		}
	}

	testInts := map[string]int {
		"nested.vars.app.inner.integer": 10,
		"nested.objects[0].integer": 100,
		"nested.objects[1].integer": 200,
	}
	for key, val := range testInts {
		if result := config.GetInt(key, 0); result != val {
			t.Error( fmt.Sprintf("Looking for \"%f\" in key %s but found %f", val, key, result))
		}
	}

	testBools := map[string]bool {
		"nested.vars.app.boolean1": false,
		"nested.vars.app.boolean2": true,
		"nested.vars.app.boolean3": false,
		"nested.vars.app.boolean4": true,
	}
	for key, val := range testBools {
		if result := config.GetBoolean(key, false); result != val {
			t.Error( fmt.Sprintf("Looking for \"%v\" in key %s but found %v", val, key, result))
		}
	}

	checkString(config,"dir.inner.inside.value", "Nested conf in directories", t)
}

func TestEvaluators(t *testing.T) {
	root, err := os.Executable()
	if err != nil {
		panic(err)
	}

	rootDir := filepath.Dir(root + "/../../")

	config := New(rootDir + "/configs", rootDir, []EvaluatorFunction {
		new(testEvalFunction),
	})

	checkString(config, "evaluators.env.instance_in_conf_default", "in conf default", t)
	checkString(config, "evaluators.env.instance", "TEST", t)
	checkString(config, "evaluators.env.sample", "2000", t)
	checkString(config, "evaluators.env.host", "testhost", t)
	checkString(config, "evaluators.testEval", "1:2:3:4:5", t)
}

func checkString(config *Config, key string, expected string, t *testing.T) {
	if value := config.GetString(key, "not found"); value != expected {
		t.Error(fmt.Sprintf("Looking for \"%s\" but found \"%s\" in key: \"%s\"", expected, value, key))
	}
}

