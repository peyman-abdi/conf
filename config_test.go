package conf_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/peyman-abdi/conf"
)

type testEvalFunction struct {
}

var _ conf.EvaluatorFunction = (*testEvalFunction)(nil)

func (t *testEvalFunction) GetFunctionName() string {
	return "paramsJoin"
}
func (t *testEvalFunction) Eval(params []string, def interface{}) interface{} {
	return strings.Join(params, ":")
}

func TestNew(t *testing.T) {
	root, err := os.Executable()
	if err != nil {
		panic(err)
	}

	rootDir := filepath.Join(filepath.Dir(root), "..")
	t.Log("Searching Config files at: " + rootDir)

	_, err = conf.New(filepath.Join(rootDir, "test_configs/valids"), rootDir, []conf.EvaluatorFunction{
		new(testEvalFunction),
	})

	if err != nil {
		t.Error(err)
	}
}

func TestNotFoundFile(t *testing.T) {
	root, err := os.Executable()
	if err != nil {
		panic(err)
	}

	rootDir := filepath.Join(filepath.Dir(root), "../does not exist/")
	t.Log("Searching Config files at: " + rootDir)

	_, err = conf.New(filepath.Join(rootDir, "test_configs"), rootDir, []conf.EvaluatorFunction{
		new(testEvalFunction),
	})

	if err == nil {
		t.Error("File not exist but got no error!")
	}
}


func TestInvalidFile(t *testing.T) {
	root, err := os.Executable()
	if err != nil {
		panic(err)
	}

	rootDir := filepath.Join(filepath.Dir(root), "..")
	t.Log("Searching Config files at: " + rootDir)

	_, err = conf.New(filepath.Join(rootDir, "test_configs/invalids"), rootDir, []conf.EvaluatorFunction{
		new(testEvalFunction),
	})

	if err == nil {
		t.Error("File not exist but got no error!")
	}

	t.Log(err)
}

func TestNestedString(t *testing.T) {
	root, err := os.Executable()
	if err != nil {
		panic(err)
	}

	rootDir := filepath.Join(filepath.Dir(root), "..")
	t.Log("Searching Config files at: " + rootDir)

	configure, err := conf.New(filepath.Join(rootDir, "test_configs/valids"), rootDir, []conf.EvaluatorFunction{
		new(testEvalFunction),
	})

	if err != nil {
		t.Error(err)
	}

	testStrings := map[string]string{
		"nested.objects[0].name":       "First",
		"nested.objects[1].name":       "Second",
		"nested.objects[0].role":       "Object",
		"nested.objects[1].role":       "Object",
		"nested.vars.app.inner.string": "Some string goes here",
		"nested.vars.app.array[1]":     "string element 1",
	}
	for key, val := range testStrings {
		checkString(configure, key, val, t)
	}

	checkString(configure, "dir.inner.inside.value", "Nested conf in directories", t)

	testStrArray := configure.GetStringArray("nested.vars.app.array", []string{})
	if len(testStrArray) != 4 || testStrArray[0] != "string element 0" {
		t.Errorf("Failed for StringArray at key nested.vars.app.array: %v with len: %d", testStrArray, len(testStrArray))
	}
}

func TestNestedIntAndFloats(t *testing.T) {
	root, err := os.Executable()
	if err != nil {
		panic(err)
	}

	rootDir := filepath.Join(filepath.Dir(root), "..")
	t.Log("Searching Config files at: " + rootDir)

	configure, err := conf.New(filepath.Join(rootDir, "test_configs/valids"), rootDir, []conf.EvaluatorFunction{
		new(testEvalFunction),
	})

	if err != nil {
		t.Error(err)
	}

	testFloats := map[string]float64{
		"nested.vars.app.inner.float": 13.333,
		"nested.objects[0].float":     103.33,
		"nested.objects[1].float":     203.33,
	}
	for key, val := range testFloats {
		if result := configure.GetFloat(key, 0); result != val {
			t.Error(fmt.Sprintf("Looking for \"%f\" in key %s but found %f", val, key, result))
		}
	}

	testInts := map[string]int{
		"nested.vars.app.inner.integer": 10,
		"nested.objects[0].integer":     100,
		"nested.objects[1].integer":     200,
	}
	for key, val := range testInts {
		if result := configure.GetInt(key, 0); result != val {
			t.Error(fmt.Sprintf("Looking for \"%d\" in key %s but found %d", val, key, result))
		}
	}

	testBools := map[string]bool{
		"nested.vars.app.boolean1": false,
		"nested.vars.app.boolean2": true,
		"nested.vars.app.boolean3": false,
		"nested.vars.app.boolean4": true,
	}
	for key, val := range testBools {
		if result := configure.GetBoolean(key, false); result != val {
			t.Error(fmt.Sprintf("Looking for \"%v\" in key %s but found %v", val, key, result))
		}
	}

	arrInt := configure.GetIntArray("nested.vars.intArray", []int{})
	if len(arrInt) != 7 || arrInt[0] != 1 || arrInt[1] != 2 {
		t.Error(fmt.Sprintf("Looking for int array in key %s but found %v", "nested.vars.intArray", arrInt))
	}

	arrFloat := configure.GetFloatArray("nested.vars.floatArray", []float64{})
	if len(arrFloat) != 5 || arrFloat[0] != 1.1 || arrFloat[1] != 1.2 {
		t.Error(fmt.Sprintf("Looking for float array in key %s but found %v", "nested.vars.intArray", arrFloat))
	}

	if bigInt := configure.GetInt64("nested.vars.bigInt", 0); bigInt != 9223372036854775806 {
		t.Errorf("Failed reading big interger int64: %d", bigInt)
	}

	if bigInt := configure.GetAsString("nested.vars.bigInt", ""); bigInt != "18446744073709551615" {
		t.Errorf("Failed reading big interger unsigned int64: %s", bigInt)
	}
}

func TestConfig_GetMap(t *testing.T) {
	root, err := os.Executable()
	if err != nil {
		t.Error(err)
	}

	rootDir := filepath.Join(filepath.Dir(root), "..")
	t.Log("Searching Config files at: " + rootDir)

	configure, err := conf.New(filepath.Join(rootDir, "test_configs/valids"), rootDir, []conf.EvaluatorFunction{
		new(testEvalFunction),
	})

	if err != nil {
		t.Error(err)
	}

	testMap := configure.GetMap("nested.vars.app", map[string]interface{} {})
	if !testMap["boolean2"].(bool) {
		t.Error("Testing GetMap Failed, expecting true for map nested.vars.app on key boolean2")
	}
}

func TestConfig_GetAsString(t *testing.T) {
	root, err := os.Executable()
	if err != nil {
		t.Error(err)
	}

	rootDir := filepath.Join(filepath.Dir(root), "..")
	t.Log("Searching Config files at: " + rootDir)

	configure, err := conf.New(filepath.Join(rootDir, "test_configs/valids"), rootDir, []conf.EvaluatorFunction{
		new(testEvalFunction),
	})

	if err != nil {
		t.Error(err)
	}

	if strIntArr := configure.GetAsString("nested.vars.intArray[1]", "not found"); strIntArr != "2" {
		t.Error("Failed for GetAsString on int array returned: " + strIntArr)
	}

	if strFlArr := configure.GetAsString("nested.vars.floatArray[1]", "not found"); strFlArr != "1.2" {
		t.Error("Failed for GetAsString on float array returned: " + strFlArr)
	}

	if str := configure.GetAsString("nested.vars.small", "not found"); str != "map[v:1 d:2]" && str != "map[d:2 v:1]" {
		t.Error("Failed for GetAsString on obj returned: " + str)
	}

	if str := configure.GetAsString("nested.vars.app.inner.string", "not found"); str != "Some string goes here" {
		t.Error("Failed for GetAsString on string returned: " + str)
	}
}

func TestConfig_GetAndIsSet(t *testing.T) {
	root, err := os.Executable()
	if err != nil {
		t.Error(err)
	}

	rootDir := filepath.Join(filepath.Dir(root), "..")
	t.Log("Searching Config files at: " + rootDir)

	configure, err := conf.New(filepath.Join(rootDir, "test_configs/valids"), rootDir, []conf.EvaluatorFunction{
		new(testEvalFunction),
	})

	if err != nil {
		t.Error(err)
	}

	if configure.IsSet("nested.vars.app.inner.string") {
		if val := configure.Get("nested.vars.app.inner.string", "").(string); val != "Some string goes here" {
			t.Error("Testing Get failed")
		}
	} else {
		t.Error("Testing IsSet failed; key nested.vars.app.inner.string is present but reported as not found!")
	}
}

func TestEnvEvaluator_GetFunctionName(t *testing.T) {
	root, err := os.Executable()
	if err != nil {
		t.Error(err)
	}

	rootDir := filepath.Join(filepath.Dir(root), "..")
	t.Log("Searching Config files at: " + rootDir)

	configure, err := conf.New(filepath.Join(rootDir, "test_configs/valids"), rootDir, []conf.EvaluatorFunction{
		new(testEvalFunction),
	})

	if err != nil {
		t.Error(err)
	}

	checkString(configure, "evaluators.env.instance_in_conf_default", "in conf default", t)
	checkString(configure, "evaluators.env.server", "Github", t)
	checkString(configure, "evaluators.env.port", "2020", t)
	checkString(configure, "evaluators.env.val", "vendor", t)
	checkString(configure, "evaluators.env.instance", "TEST", t)
	checkString(configure, "evaluators.env.sample", "2000", t)
	checkString(configure, "evaluators.env.host", "testhost", t)
	checkString(configure, "evaluators.testEval", "1:2:3:4:5", t)
	checkString(configure, "evaluators.not_to_be_found", "not found", t)
	checkString(configure, "evaluators.env.noParam", "not found", t)
}

func Test_DefaultValues(t *testing.T)  {
	root, err := os.Executable()
	if err != nil {
		t.Error(err)
	}

	rootDir := filepath.Join(filepath.Dir(root), "..")
	t.Log("Searching Config files at: " + rootDir)

	configure, err := conf.New(filepath.Join(rootDir, "test_configs/valids"), rootDir, []conf.EvaluatorFunction{
		new(testEvalFunction),
	})

	if dArr := configure.GetString("nested.vars.app.array[not_a_number]", "default val"); dArr != "default val" {
		t.Error("Failed for returning default string for invalid array indexer")
	}
	if dStr := configure.GetString("do.not.exist", "default string"); dStr != "default string" {
		t.Error("Failed for returning default string")
	}
	if dInt := configure.GetInt("do.not.exist", 2020); dInt != 2020 {
		t.Error("Failed for returning default string")
	}
	if dFloat := configure.GetFloat("do.not.exist", 12.12); dFloat != 12.12 {
		t.Error("Failed for returning default string")
	}
	if dMap := configure.GetMap("do.not.exist", map[string]interface{} { "key": "value" }); dMap["key"] != "value" {
		t.Error("Failed for returning default map")
	}
}

func checkString(configure *conf.Config, key string, expected string, t *testing.T) {
	if value := configure.GetString(key, "not found"); value != expected {
		t.Error(fmt.Sprintf("Looking for \"%s\" but found \"%s\" in key: \"%s\"", expected, value, key))
	}
}
