package config

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
	t.Log("Test started at path: " + rootDir)

	Initialize(rootDir + "/configs", rootDir, []EvaluatorFunction {
		new(testEvalFunction),
	})
}

func TestNestedObjects(t *testing.T) {
	testStrings := map[string]string {
		"nested.objects[0].name": "First",
		"nested.objects[1].name": "Second",
		"nested.objects[0].role": "Object",
		"nested.objects[1].role": "Object",
		"nested.vars.app.inner.string": "Some string goes here",
		"nested.vars.app.array[1]": "string element 1",
	}
	for key, val := range testStrings {
		checkString(key, val, t)
	}

	testFloats := map[string]float64 {
		"nested.vars.app.inner.float": 13.333,
		"nested.objects[0].float": 103.33,
		"nested.objects[1].float": 203.33,
	}
	for key, val := range testFloats {
		if result := GetFloat(key, 0); result != val {
			t.Error( fmt.Sprintf("Looking for \"%f\" in key %s but found %f", val, key, result))
		}
	}

	testInts := map[string]float64 {
		"nested.vars.app.inner.integer": 10,
		"nested.objects[0].integer": 100,
		"nested.objects[1].integer": 200,
	}
	for key, val := range testInts {
		if result := GetFloat(key, 0); result != val {
			t.Error( fmt.Sprintf("Looking for \"%f\" in key %s but found %f", val, key, result))
		}
	}

	checkString("dir.inner.inside.value", "Nested conf in directories", t)
}

func TestEvaluators(t *testing.T) {
	checkString("evaluators.env.instance_in_conf_default", "in conf default", t)
	checkString("evaluators.env.instance", "TEST", t)
	checkString("evaluators.env.sample", "2000", t)
	checkString("evaluators.env.host", "testhost", t)
	checkString("evaluators.testEval", "1:2:3:4:5", t)
}

func checkString(key string, expected string, t *testing.T) {
	if value := GetString(key, "not found"); value != expected {
		t.Error(fmt.Sprintf("Looking for \"%s\" but found \"%s\" in key: \"%s\"", expected, value, key))
	}
}

