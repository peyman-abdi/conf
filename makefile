##### CONFIG BUILD
DEBUG=1
ENTRY=config.go
VERSION=1.0.0-Alpha1

#### EXPORT PATHES
BINARY_PATH=build

##### DEPENDENCIES
DEPENDENCIES=\
 github.com/joho/godotenv \
 github.com/hjson/hjson-go \

##### BUILD COMMANDS
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

##### METHODS
define getVariant
$(if ($(DEBUG),1),DEBUG,PRODUCTION)
endef
define getPlatform
$(if ($(OS),Windows_NT),$(if ($(shell uname -s),Linux),OSX,LINUX),WIN32)
endef

#### BUILD FLAGS
PLATFORM=$(call getPlatform)
VARIANT=$(call getVariant)
BUILD_TIME=$(shell date +%FT%T%Z)
BUILD_CODE=$(shell git rev-parse HEAD)
PACKAGE=avalanche/app/core/app

LDFLAGS=-ldflags "-X $(PACKAGE).Version=$(VERSION) -X $(PACKAGE).Code=$(BUILD_CODE) -X $(PACKAGE).Variant=$(VARIANT) -X $(PACKAGE).Platform=$(PLATFORM) -X $(PACKAGE).BuildTime=$(BUILD_TIME)"

#### SCRIPTS
all: go_get test

go_get:
	@($(foreach dep, $(DEPENDENCIES), $(GOGET) $(dep);))
test:
	$(GOTEST) -o $(BINARY_PATH)/$(basename $(notdir $(ENTRY)))  config_test.go
clean:
	$(GOCLEAN)
	rm -f $(BINARY_PATH)/*

