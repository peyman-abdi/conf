package nemo

import (
	"strings"
	"os"
	"fmt"
	"path/filepath"
	"strconv"
	"github.com/joho/godotenv"
	"reflect"
	"io/ioutil"
	"github.com/hjson/hjson-go"
	"flag"
)

func New(configDir string, envDir string, evalFunctions []EvaluatorFunction) *Config {
	config := new(Config)

	var configFiles []string

	configFiles = iterateForConfig(configDir, configFiles)

	config.configs = make(map[string]interface{})
	for _, file := range configFiles {
		if !strings.HasSuffix(file, ".hjson") {
			continue
		}

		content, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Println("Failed reading file at path: " + file)
			panic(err)
		}

		var conf map[string]interface{}
		err = hjson.Unmarshal(content, &conf)
		if err != nil {
			fmt.Println("Failed deserializing lib file at path: " + file)
			panic(err)
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

		config.configs[filename] = conf
	}

	if envDir != "" {
		err := godotenv.Load(filepath.Join(envDir, ".env"))
		if err != nil {
			panic(err)
		}

		if flag.Lookup("test.v") != nil {
			err = godotenv.Overload(filepath.Join(envDir, ".env.test"))
			if err != nil {
				panic(err)
			}
		}
	}

	envEval := new(envEvaluator)
	config.evaluatorsFunctions = map[string]EvaluatorFunction {
		envEval.GetFunctionName(): envEval,
	}
	if evalFunctions != nil {
		for _, evalFunc := range evalFunctions {
			config.evaluatorsFunctions[evalFunc.GetFunctionName()] = evalFunc
		}
	}

	return config
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

	return iterateForKey(config, &keys, config.configs, def)
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
		arrayIndex, err = strconv.Atoi(key[indexStart+1:indexEnd])
		if err != nil {
			panic(err)
		}

		key = keyName
	}

	if len(*keys) == 1 {
		if conf[key] != nil {
			if isArray {
				return conf[key].([]interface{})[arrayIndex]
			} else {
				if reflect.TypeOf(conf[key]).Kind() == reflect.String {
					strKey := conf[key].(string)
					return evalStringValue(config, strKey, def)
				}
				return conf[key]
			}
		} else {
			return def
		}
	} else {
		if conf[key] != nil {
			if isArray {
				newKeys := (*keys)[1:]
				arr := conf[key].([]interface{})
				return iterateForKey(config, &newKeys, arr[arrayIndex].(map[string]interface{}), def)
			} else {
				newKeys := (*keys)[1:]
				return iterateForKey(config, &newKeys, conf[key].(map[string]interface{}), def)
			}
		} else {
			return def
		}
	}
}
func evalStringValue(config *Config, content string, def interface{}) interface{} {
	evalStartIndex := strings.Index(content, "(")
	evalEndIndex := strings.Index(content, ")")
	if evalStartIndex > 0 && evalEndIndex > 0 {
		methodName := strings.Trim(content[:evalStartIndex], "\"\t' ")
		if config.evaluatorsFunctions[methodName] != nil {
			evalParamsString := content[evalStartIndex+1:evalEndIndex]
			evalParams := strings.Split(evalParamsString, ",")
			var evalParamsSanitized []string
			for _,param := range evalParams {
				evalParamsSanitized = append(evalParamsSanitized, strings.Trim(param, "\"\t' "))
			}

			return config.evaluatorsFunctions[methodName].Eval(evalParamsSanitized, def)
		}
	}
	return content
}

type EvaluatorFunction interface {
	Eval(params []string, def interface{}) interface{}
	GetFunctionName() string
}

type Config struct {
	configs map[string]interface{}
	evaluatorsFunctions map[string]EvaluatorFunction
}
func (c *Config) Get(key string, def interface{}) interface{} {
	return get(c, key, def)
}
func (c *Config) GetString(key string, def string) string {
	return c.Get(key, def).(string)
}
func (c *Config) GetInt(key string, def int) int {
	return int(c.Get(key, def).(float64))
}
func (c *Config) GetFloat(key string, def float64) float64 {
	return c.Get(key, def).(float64)
}
func (c *Config) GetBoolean(key string, def bool) bool {
	val, ok := c.Get(key, def).(bool)
	if !ok {
		val, ok := c.Get(key, def).(float64)
		if ok {
			return val == 1
		} else {
			return def
		}
	} else {
		return val
	}
}
func (c *Config) GetStringArray(key string, def []string) []string {
	arr := c.Get(key, def).([]interface{})
	var foundStrings []string
	for _, item := range arr {
		foundStrings = append(foundStrings, item.(string))
	}
	return foundStrings
}
func (c *Config) GetIntArray(key string, def []int) []int {
	return c.Get(key, def).([]int)
}
func (c *Config) GetFloatArray(key string, def []float64) []float64 {
	return c.Get(key, def).([]float64)
}
func (c *Config) GetMap(key string, def map[string]interface{}) map[string]interface{} {
	return c.Get(key, def).(map[string]interface{})
}
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

