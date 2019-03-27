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
 golang.org/x/tools/cmd/cover \
 github.com/mattn/goveralls

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

#### SCRIPTS
all: go_get test

go_get:
	@($(foreach dep, $(DEPENDENCIES), $(GOGET) $(dep);))
test:
	$(GOTEST) -c -o $(BINARY_PATH)/config_test -v -covermode=count ./ && $(BINARY_PATH)/config_test -test.coverprofile coverage.out
clean:
	$(GOCLEAN)
	rm -f $(BINARY_PATH)/*

