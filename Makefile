PACKAGE := $(shell go list -e)
APP_NAME = $(lastword $(subst /, ,$(PACKAGE)))
MAIN_APP_DIR = cmd/main

include gomakefiles/common.mk
include gomakefiles/metalinter.mk

SOURCES := $(shell find $(SOURCEDIR) -name '*.go' -or -name '*.js' \
	-not -path './vendor/*')

$(MAIN_APP_DIR)/$(APP_NAME): $(SOURCES) config.toml
	cd $(MAIN_APP_DIR)/ && go build -ldflags '-X main.Version=${VERSION}' -o ${APP_NAME}

${RELEASE_SOURCES}: $(SOURCES)

include gomakefiles/semaphore.mk

config.toml: $(MAIN_APP_DIR)/config.toml.template
ifndef FLOWDOCK_API_TOKEN
	$(error FLOWDOCK_API_TOKEN parameter must be set)
endif
ifndef FLOWDOCK_NICK
	$(error FLOWDOCK_NICK parameter must be set)
endif
	cat $(MAIN_APP_DIR)/config.toml.template | sed \
		-e "s/FLOWDOCK_API_TOKEN/$$FLOWDOCK_API_TOKEN/" \
		-e "s/FLOWDOCK_NICK/$$FLOWDOCK_NICK/" \
		> $(MAIN_APP_DIR)/config.toml

$(MAIN_APP_DIR)/archive.zip: $(MAIN_APP_DIR)/$(APP_NAME) $(MAIN_APP_DIR)/config.toml
	cd $(MAIN_APP_DIR) && zip archive.zip $(APP_NAME) main.js config.toml
	printf "@ $(APP_NAME)\n@=app\n" | zipnote -w $(MAIN_APP_DIR)/archive.zip

.PHONY: _check_aws_params
_check_aws_params: 
ifndef AWS_ACCESS_KEY_ID
	$(error AWS_ACCESS_KEY_ID parameter must be set)
endif
ifndef AWS_SECRET_ACCESS_KEY
	$(error AWS_SECRET_ACCESS_KEY parameter must be set)
endif
ifndef AWS_REGION
	$(error AWS_REGION parameter must be set)
endif

.PHONY: deployaws
deployaws: $(MAIN_APP_DIR)/archive.zip _check_aws_params
	docker run -i --rm \
		-v $(abspath $(MAIN_APP_DIR)):/data \
	    --env AWS_ACCESS_KEY_ID=$$AWS_ACCESS_KEY_ID \
	    --env AWS_SECRET_ACCESS_KEY=$$AWS_SECRET_ACCESS_KEY \
	    garland/aws-cli-docker \
	    aws lambda update-function-code \
		  --region $$AWS_REGION \
		  --function-name $(APP_NAME) \
		  --zip-file fileb:///data/archive.zip 

.PHONY: invoke
invoke: _check_aws_params
	docker run -i --rm \
		-v $(abspath $(MAIN_APP_DIR)):/data \
	    --env AWS_ACCESS_KEY_ID=$$AWS_ACCESS_KEY_ID \
	    --env AWS_SECRET_ACCESS_KEY=$$AWS_SECRET_ACCESS_KEY \
	    garland/aws-cli-docker \
	    aws lambda invoke \
		  --region $$AWS_REGION \
		  --log-type Tail \
		  --function-name $(APP_NAME) \
		  /tmp/invoke_output | jq '.LogResult' -r | base64 --decode

.PHONY: prepare
prepare: prepare_metalinter prepare_github_release

.PHONY: clean
clean: clean_common clean_bindata
	rm -rf $(MAIN_APP_DIR)/${APP_NAME}
	rm -rf $(MAIN_APP_DIR)/${APP_NAME}.exe
