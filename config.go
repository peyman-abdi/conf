// Package conf provides all functionality required for parsing and accessing
// configuration files.
// You can use a hjson/json/env files as configurations and access them recursively
// with dots
package conf

import (
	"flag"
	"fmt"
	"github.com/hjson/hjson-go"
	"github.com/joho/godotenv"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

// New creates a new config parser with config files at path configDir.
// Argument envDir can be an empty string which will cause the function to ignore .env parsing
// Config files with extensions .hjson and .json will be opened at this point and all files will be parsed
// resulting in a fast access time
// If there are any Evaluation needed those will be applied when accessing variables
// All folders inside configDir will recursively scanned for .hjson and .json files and
// any config will be accessible by its relative path connected with dots
// An error may happen during reading files like access denied
// if the error causes
func New(configDir string, envDir string, evalFunctions []EvaluatorFunction) (config *Config, err error) {
	config = new(Config)
	err = nil

	var configFiles []string

	configFiles = iterateForConfig(configDir, configFiles)

	config.ConfigsMap = make(map[string]interface{})
	for _, file := range configFiles {
		if !strings.HasSuffix(file, ".hjson") && !strings.HasSuffix(file, ".json") {
			continue
		}

		content, errF := ioutil.ReadFile(file)
		if errF != nil {
			config = nil
			err = errF
			return
		}

		var conf map[string]interface{}
		err = hjson.Unmarshal(content, &conf)
		if err != nil {
			config = nil
			return
		}

		dirname, filename := filepath.Split(file)
		filename = filename[:strings.Index(filename, ".hjson")]
		if strings.Trim(dirname, "/\\") != strings.Trim(configDir, "/\\") {
			dirname = dirname[len(configDir):]
			dirname = strings.Replace(strings.Trim(dirname, "/\\"), "/", ".", -1) + "." + filename
			folders := strings.Split(dirname, ".")
			lenFolders := len(folders)
			filename = folders[0]
			for index := lenFolders - 1; index > 0; index-- {
				conf = map[string]interface{}{
					folders[index]: conf,
				}
			}
		}

		config.ConfigsMap[filename] = conf
	}

	if envDir != "" {
		err = godotenv.Load(filepath.Join(envDir, ".env"))
		if err == nil {
			if flag.Lookup("test.v") != nil {
				err = godotenv.Overload(filepath.Join(envDir, ".env.test"))
			}
		}
	}

	envEval := new(envEvaluator)
	config.EvaluatorFunctionsMap = map[string]EvaluatorFunction{
		envEval.GetFunctionName(): envEval,
	}
	if evalFunctions != nil {
		for _, evalFunc := range evalFunctions {
			config.EvaluatorFunctionsMap[evalFunc.GetFunctionName()] = evalFunc
		}
	}

	return
}

func iterateForConfig(topPath string, configFiles []string) []string {
	filepath.Walk(topPath, func(path string, info os.FileInfo, err error) error {
		if info != nil {
			if info.IsDir() && path != topPath {
				if info.Name() == "test" {
					if flag.Lookup("test.v") == nil {
						return nil
					}
				}
				iterateForConfig(path, configFiles)
			} else {
				configFiles = append(configFiles, path)
			}
		}
		return nil
	})
	return configFiles
}
func get(config *Config, key string, def interface{}) interface{} {
	keys := strings.Split(key, ".")
	if len(keys) < 1 {
		return def
	}

	return iterateForKey(config, &keys, config.ConfigsMap, def)
}
func iterateForKey(config *Config, keys *[]string, conf map[string]interface{}, def interface{}) interface{} {
	if len(*keys) < 1 {
		return def
	}

	key := (*keys)[0]
	isArray := false
	var arrayIndex int
	var err error
	if strings.Contains(key, "[") && strings.Contains(key, "]") {
		indexStart := strings.Index(key, "[")
		indexEnd := strings.Index(key, "]")
		keyName := key[:indexStart]

		isArray = true
		arrayIndex, err = strconv.Atoi(key[indexStart+1 : indexEnd])
		if err != nil {
			panic(err)
		}

		key = keyName
	}

	if len(*keys) == 1 {
		if conf[key] != nil {
			if isArray {
				return conf[key].([]interface{})[arrayIndex]
			}

			if reflect.TypeOf(conf[key]).Kind() == reflect.String {
				strKey := conf[key].(string)
				return evalStringValue(config, strKey, def)
			}
			return conf[key]
		}

		return def
	} else {
		if conf[key] != nil {
			if isArray {
				newKeys := (*keys)[1:]
				arr := conf[key].([]interface{})
				return iterateForKey(config, &newKeys, arr[arrayIndex].(map[string]interface{}), def)
			}

			newKeys := (*keys)[1:]
			return iterateForKey(config, &newKeys, conf[key].(map[string]interface{}), def)
		}

		return def
	}
}
func evalStringValue(config *Config, content string, def interface{}) interface{} {
	evalStartIndex := strings.Index(content, "(")
	evalEndIndex := strings.Index(content, ")")
	if evalStartIndex > 0 && evalEndIndex > 0 {
		methodName := strings.Trim(content[:evalStartIndex], "\"\t' ")
		if config.EvaluatorFunctionsMap[methodName] != nil {
			evalParamsString := content[evalStartIndex+1 : evalEndIndex]
			evalParams := strings.Split(evalParamsString, ",")
			var evalParamsSanitized []string
			for _, param := range evalParams {
				evalParamsSanitized = append(evalParamsSanitized, strings.Trim(param, "\"\t' "))
			}

			return config.EvaluatorFunctionsMap[methodName].Eval(evalParamsSanitized, def)
		}
	}
	return content
}

// EvaluatorFunction lets you create dynamic config values
// they act like functions inside your hjson/json files
// each function when called inside config files can have any number of
// arguments and they are passed to the Eval function.
type EvaluatorFunction interface {
	// Evaluate the input arguments and return the value
	// if an error happened just return the def value
	Eval(params []string, def interface{}) interface{}

	// Get the name of the function
	// Returned value of this function is the string
	// that is used inside config files to call this evaluator
	GetFunctionName() string
}

// Config object for accessing configurations as a map
// all files are parsed at Creation time and recursively added
// to the ConfigsMap.
// Its much better and easier to use Getter functions of the struct.
// But when accessing full config objects are needed they are available with GetMap function or
// throw the ConfigsMap object
type Config struct {
	ConfigsMap            map[string]interface{}
	EvaluatorFunctionsMap map[string]EvaluatorFunction
}

// IsSet returns true if there is value for key, false otherwise
func (c *Config) IsSet(key string) bool {
	return get(c, key, nil) != nil
}

// Get returns the raw interface{} value of a key
// You have to convert it to your desired type
// If you have used a custom EvaluatorFunction to generate the value
// simply cast the interface{} to your desired type
func (c *Config) Get(key string, def interface{}) interface{} {
	return get(c, key, def)
}

// GetString checks if the value of the key can be converted to string or not
// if not or if the key does not exist returns the def value
func (c *Config) GetString(key string, def string) string {
	strVal, ok := c.Get(key, def).(string)
	if ok {
		return strVal
	}
	return def
}

// GetInt checks if the value of the key can be converted to int or not
// if not or if the key does not exist returns the def value
func (c *Config) GetInt(key string, def int) int {
	floatVal, ok := c.Get(key, def).(float64)
	if ok {
		return int(floatVal)
	}
	return def
}

// GetInt64 checks if the value of the key can be converted to int64 or not
// if not or if the key does not exist returns the def value
func (c *Config) GetInt64(key string, def int64) int64 {
	floatVal, ok := c.Get(key, def).(float64)
	if ok {
		return int64(floatVal)
	}
	return def
}

// GetFloat checks if the value of the key can be converted to float64 or not
// if not or if the key does not exist returns the def value
func (c *Config) GetFloat(key string, def float64) float64 {
	floatVal, ok := c.Get(key, def).(float64)
	if ok {
		return floatVal
	}
	return def
}

// GetBoolean checks if the value of the key can be converted to boolean or not
// if not or if the key does not exist returns the def value
// valid values are true,false,1,0
func (c *Config) GetBoolean(key string, def bool) bool {
	val, ok := c.Get(key, def).(bool)
	if !ok {
		val, ok := c.Get(key, def).(float64)
		if ok {
			return val == 1
		}

		return def
	}

	return val
}

// GetStringArray checks if the value of the key can be converted to []string or not
// if not or if the key does not exist returns the def value
func (c *Config) GetStringArray(key string, def []string) []string {
	arr, ok := c.Get(key, def).([]string)
	if ok {
		return arr
	}

	arrS := c.Get(key, def).([]interface{})
	var foundStrings = make([]string, len(arrS))
	for index, item := range arrS {
		foundStrings[index] = item.(string)
	}
	return foundStrings
}

// GetIntArray checks if the value of the key can be converted to []int or not
// if not or if the key does not exist returns the def value
func (c *Config) GetIntArray(key string, def []int) []int {
	arr, ok := c.Get(key, def).([]int)
	if ok {
		return arr
	}

	arrI := c.Get(key, def).([]interface{})
	var foundArray = make([]int, len(arrI))
	for index, item := range arrI {
		foundArray[index] = int(item.(float64))
	}
	return foundArray
}

// GetFloatArray checks if the value of the key can be converted to []float64 or not
// if not or if the key does not exist returns the def value
func (c *Config) GetFloatArray(key string, def []float64) []float64 {
	arr, ok := c.Get(key, def).([]float64)
	if ok {
		return arr
	}

	arrF := c.Get(key, def).([]interface{})
	var foundArray = make([]float64, len(arrF))
	for index, item := range arrF {
		foundArray[index] = item.(float64)
	}
	return foundArray
}

// GetMap returns the raw config object as map of strings
func (c *Config) GetMap(key string, def map[string]interface{}) map[string]interface{} {
	mapVal, ok := c.Get(key, def).(map[string]interface{})
	if ok {
		return mapVal
	}
	return def
}

// GetAsString converts the value of the key to string and returns it,
// if the key does not exist returns the def value
func (c *Config) GetAsString(key string, def string) string {
	val := c.Get(key, def)
	switch val.(type) {
	case string:
		return val.(string)
	case int:
		return strconv.Itoa(val.(int))
	case float64:
		return strconv.FormatFloat(val.(float64), 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", val)
	}
}
