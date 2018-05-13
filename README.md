# Conf
[![Build Status](https://travis-ci.org/peyman-abdi/config.svg?branch=master)](https://travis-ci.org/peyman-abdi/config)

**Conf** provides easy and powerfull methods to access configurations stored in **hjson**, **json**, **env** files. You pass root configurations directory to the package and it will parse all config files recursively. Provide your own **Evaluators** and make a dynamic config files by calling your applications functions.

# Features

 - Recursively access config files stored in directories with a dot
 - Get entire config objects as `map[string]interface{}`
 - Built in methods for accessing `string` `int` `float64` `int64` `[]string` `[]int` `[]float` `[]interface` `map[string]interface{}`
 - Use **HJSON** or **JSON** syntax for configuration files
 - Use **.env** file to override environment variables
 - USE **.env.test** file to override environment variables in test mode
 - Use `env("environment variable name", "default value")` in configuration files to access environment varaiables in config files
 - Provide your own **Evaluators** to access custom app variables while parsing configurations

## Documentation
[![GoDoc](https://godoc.org/github.com/peyman-abdi/conf?status.svg)](https://godoc.org/github.com/peyman-abdi/conf)

## Dependencies

- [godotenv](https://github.com/joho/godotenv)
- [hjson-go](https://github.com/hjson/hjson-go)

## Installation

    go get github.com/peyman-abdi/conf

## Usage

### Basic usage

```go
import "peyman-abdi/conf"
...

config, err := conf.New("/path/to/configs/dir", "/path/to/envs/dir", nil)
if err != nil {
    fmt.Println(err)
}

config.GetString("filename.object.innerobject.value", "default")
config.GetString("dir.another_dir.filename.object.array[3].value", "default")
...
```

### Access Environment Variables in config files

use `env()` function in json/hjson files to access environment variables

```go
// app.hjson
{
    server: {
        port: env(PORT, 8080),
        host: env(HOST, "localhost"),
    },
    database: {
        table: env(TABLE),
        password: env(PASSWORD),
        username: env(USERNAME, "root"),
    },
    debug: true,
    logger: {
        other: "vairables",
    }
}

// .env
HOST=github.com
PASSWORD=secret

// main.go
config.GetInt("app.server.port", 0) 				// returns 8080
config.GetString("app.server.host", "domain.com") 	// returns "github.com"
config.GetString("app.database.table", "my_table") 	// returns "my_table"
config.GetString("app.database.password", "") 		// returns "secret"
config.GetString("app.database.username", "user") 	// returns "root"
```

### Custom Evaluators

Use custom evaluators to build your own functions to be used inside json/hjson files.

```go
// my_evals.go
type MyJoinEvaluatorFunction struct {
}
var _ conf.EvaluatorFunction = (*MyJoinEvaluatorFunction)(nil)

func (_ *MyJoinEvaluatorFunction) GetFunctionName() string {
   return "myJoinFunction" // the function name used in hjson/json files
}
func (_ *MyJoinEvaluatorFunction) Eval(params []string, def interface{}) interface{} {
    if len(params) > 0 {
        return strings.Join(params, "::")
    }
    return def
}


// my.hjson
{
    joined: myJoinFunction(1,2,3,4,5)
}

// main.go
config, err := conf.New("/path/to/configs/dir", "/path/to/envs/dir", []conf.EvaluatorFunction {
   new(MyJoinEvaluatorFunction),
})

config.GetString("my.joined", "") // returns "1::2::3::4::5"
```

you can use this functionallity and add more power to your config files, like:
- relative pathes
- time functions
- os dependant evaluations
- ...

