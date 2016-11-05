PACKAGE := $(shell go list -e)
APP_NAME = $(lastword $(subst /, ,$(PACKAGE)))
MAIN_APP_DIR = cmd/main

include gomakefiles/common.mk
include gomakefiles/metalinter.mk

SOURCES := $(shell find $(SOURCEDIR) -name '*.go' -or -name '*.js' \
	-not -path './vendor/*')

$(MAIN_APP_DIR)/$(APP_NAME): $(SOURCES)
	. personal.env && cd $(MAIN_APP_DIR)/ && go build -ldflags \
		"-X main.SiteDeployment=https://$$CLOUDFRONT_DOMAIN" -o ${APP_NAME}

${RELEASE_SOURCES}: $(SOURCES)

include gomakefiles/semaphore.mk

cf/archive.zip: $(MAIN_APP_DIR)/$(APP_NAME)
	@rm cf/archive.zip > /dev/null 2>&1 || true
	@cd $(MAIN_APP_DIR) && zip archive.zip $(APP_NAME) main.js
	@mv $(MAIN_APP_DIR)/archive.zip cf/
	@printf "@ $(APP_NAME)\n@=app\n" | zipnote -w cf/archive.zip

aws = . personal.env && docker run -i --rm \
		-v $(abspath .):/data \
		-v $(abspath .)/cf/.aws:/root/.aws \
	    --env AWS_ACCESS_KEY_ID=$$AWS_ACCESS_KEY_ID \
	    --env AWS_SECRET_ACCESS_KEY=$$AWS_SECRET_ACCESS_KEY \
	    garland/aws-cli-docker \
	    aws --region $$AWS_REGION

.PHONY: unform
unform: cf/archive.zip
	@$(aws) cloudformation delete-stack  \
		  --stack-name igor

.PHONY: form
form: cf/archive.zip lambda-upload
	@$(aws) cloudformation create-stack  \
		  --stack-name igor \
		  --template-body file:///data/cf/stack.template \
		  --capabilities CAPABILITY_IAM \
		  --parameters \
		  	ParameterKey=DeploymentBucket,ParameterValue=$$BUCKET_DEPLOYMENT \
		  	ParameterKey=WebSiteBucket,ParameterValue=$$BUCKET_SITE \
			ParameterKey=CertificateArn,ParameterValue=$$GENERATED_CERTIFICATE_ARN \
			ParameterKey=DomainName,ParameterValue=$$CLOUDFRONT_DOMAIN \
		  	ParameterKey=CognitoPoolArn,ParameterValue=$$GENERATED_COGNITO_POOL_ID
	$(MAKE) --silent wait-for-status EXPECTED=CREATE_COMPLETE FAILURE=CREATE_ROLLBACK_COMPLETE

.PHONY: reform
reform:
	@$(aws) cloudformation update-stack  \
		  --stack-name igor \
		  --template-body file:///data/cf/stack.template \
		  --capabilities CAPABILITY_IAM \
		  --parameters \
		  	ParameterKey=DeploymentBucket,ParameterValue=$$BUCKET_DEPLOYMENT \
		  	ParameterKey=WebSiteBucket,ParameterValue=$$BUCKET_SITE \
			ParameterKey=CertificateArn,ParameterValue=$$GENERATED_CERTIFICATE_ARN \
			ParameterKey=DomainName,ParameterValue=$$CLOUDFRONT_DOMAIN \
		  	ParameterKey=CognitoPoolArn,ParameterValue=$$GENERATED_COGNITO_POOL_ID || true
	$(MAKE) --silent wait-for-status EXPECTED=UPDATE_COMPLETE FAILURE=UPDATE_ROLLBACK_COMPLETE

.PHONY: lambda-upload
lambda-upload: cf/archive.zip
	@$(aws) s3 cp  \
		  /data/cf/archive.zip \
		  s3://$$BUCKET_DEPLOYMENT/deployment/$(APP_NAME).zip
	@$(aws) lambda update-function-code \
		  --function-name $(APP_NAME) \
		  --s3-bucket $$BUCKET_DEPLOYMENT \
		  --s3-key deployment/$(APP_NAME).zip

.PHONY: lambda-update-from-local
lambda-update-from-local: cf/archive.zip
	@$(aws) lambda update-function-code \
		  --function-name $(APP_NAME) \
		  --zip-file fileb:///data/cf/archive.zip 

.PHONY: lambda-invoke
lambda-invoke:
	@$(aws) lambda invoke \
		  --log-type Tail \
		  --function-name $(APP_NAME) \
		  /tmp/invoke_output | ./jq '.LogResult' -r | base64 --decode

.PHONY: site-prepare
site-prepare:
	@. personal.env && cat site/config.template.js \
		| sed \
			-e "s/GOOGLE_OAUTH2_CLIENT_ID/$$GOOGLE_OAUTH2_CLIENT_ID/g" \
			-e "s/GENERATED_COGNITO_POOL_ID/$$GENERATED_COGNITO_POOL_ID/g" \
			-e "s/AWS_REGION/$$AWS_REGION/g" \
		> site/config.js

.PHONY: site-deploy
site-deploy: site-prepare
	@$(aws) s3 sync --acl public-read --exclude '*.template.js' \
		  /data/site/ \
		  s3://$$BUCKET_SITE/

.PHONY: site-invalidate
site-invalidate:
	@$(aws) cloudfront create-invalidation --distribution-id $$GENERATED_DISTRIBUTION_ID --paths '/*'

.PHONY: wait-for-status
wait-for-status:
ifndef EXPECTED
	$(error EXPECTED parameter must be set)
endif
ifndef FAILURE
	$(error FAILURE parameter must be set)
endif
	@while [ 1 ] ;\
	do \
		status=`$(MAKE) --silent get-status`; \
		if [[ "$$EXPECTED" == "$$status" ]] ; then \
			echo "success, status=$$status!"; \
			exit 0; \
		elif [[ "$$FAILURE" == "$$status" ]] ; then \
			echo "failure, status=$$status!"; \
			exit 1; \
		fi; \
		echo "Waiting, current status is: $$status"; \
		sleep 2; \
	done

.PHONY: get-status
get-status:
	@$(aws) cloudformation describe-stacks --stack-name igor \
		| ./jq '.Stacks[0].StackStatus' -r

.PHONY: prepare
prepare: prepare_metalinter
	@curl -Lo jq https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64
	@chmod +x jq
	@$(aws) configure set preview.cloudfront true

.PHONY: clean
clean: clean_common
