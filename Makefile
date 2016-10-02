PACKAGE := $(shell go list -e)
APP_NAME = $(lastword $(subst /, ,$(PACKAGE)))

include gomakefiles/common.mk
include gomakefiles/metalinter.mk

SOURCES := $(shell find $(SOURCEDIR) -name '*.go' -or -name '*.js' \
	-not -path './vendor/*')

$(MAIN_APP_DIR)/$(APP_NAME): $(SOURCES)
	cd $(MAIN_APP_DIR)/ && go build -ldflags '-X main.Version=${VERSION}' -o ${APP_NAME}

${RELEASE_SOURCES}: $(SOURCES)

include gomakefiles/semaphore.mk

archive.zip: $(APP_NAME) 
	zip archive.zip app main.js

.PHONY: deployaws
deployaws: archive.zip
ifndef AWS_ACCESS_KEY_ID
	$(error AWS_ACCESS_KEY_ID parameter must be set)
endif
ifndef AWS_SECRET_ACCESS_KEY
	$(error AWS_SECRET_ACCESS_KEY parameter must be set)
endif
ifndef AWS_REGION
	$(error AWS_REGION parameter must be set)
endif
	docker run -i --rm \
		-v $(abspath $(MAIN_APP_DIR)):/data \
	    --env AWS_ACCESS_KEY_ID=$$AWS_ACCESS_KEY_ID \
	    --env AWS_SECRET_ACCESS_KEY=$$AWS_SECRET_ACCESS_KEY \
	    garland/aws-cli-docker \
	    aws lambda update-function-code \
		  --region $$AWS_REGION \
		  --function-name "flowdock-notifier" \
		  --zip-file fileb:///data/archive.zip 

.PHONY: invoke
invoke:
ifndef AWS_ACCESS_KEY_ID
	$(error AWS_ACCESS_KEY_ID parameter must be set)
endif
ifndef AWS_SECRET_ACCESS_KEY
	$(error AWS_SECRET_ACCESS_KEY parameter must be set)
endif
ifndef AWS_REGION
	$(error AWS_REGION parameter must be set)
endif
	docker run -i --rm \
		-v $(abspath $(MAIN_APP_DIR)):/data \
	    --env AWS_ACCESS_KEY_ID=$$AWS_ACCESS_KEY_ID \
	    --env AWS_SECRET_ACCESS_KEY=$$AWS_SECRET_ACCESS_KEY \
	    garland/aws-cli-docker \
	    aws lambda invoke \
		  --region $$AWS_REGION \
		  --log-type Tail \
		  --function-name "flowdock-notifier" \
		  /tmp/invoke_output | jq '.LogResult' -r | base64 --decode

.PHONY: prepare
prepare: prepare_metalinter prepare_github_release

.PHONY: clean
clean: clean_common clean_bindata
	rm -rf $(MAIN_APP_DIR)/${APP_NAME}
	rm -rf $(MAIN_APP_DIR)/${APP_NAME}.exe
