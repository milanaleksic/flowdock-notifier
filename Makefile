PACKAGE := $(shell go list -e)
APP_NAME = $(lastword $(subst /, ,$(PACKAGE)))

include gomakefiles/common.mk
include gomakefiles/metalinter.mk
include gomakefiles/upx.mk

SOURCES := $(shell find $(SOURCEDIR) -name '*.go' -or -name '*.js' \
	-not -path './vendor/*')

$(MAIN_APP_DIR)/$(APP_NAME): $(SOURCES)
	cd $(MAIN_APP_DIR)/ && go build -ldflags '-X main.Version=${VERSION}' -o ${APP_NAME}

${RELEASE_SOURCES}: $(SOURCES)

include gomakefiles/semaphore.mk

.PHONY: package
package: $(APP_NAME) 
	zip archive.zip flowdock-notifier main.js

.PHONY: prepare
prepare: prepare_metalinter prepare_upx prepare_github_release

.PHONY: clean
clean: clean_common clean_bindata
	rm -rf $(MAIN_APP_DIR)/${APP_NAME}
	rm -rf $(MAIN_APP_DIR)/${APP_NAME}.exe
