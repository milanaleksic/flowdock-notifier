#!/bin/bash

. "$( dirname "${BASH_SOURCE[0]}" )"/_commons.sh

roleArn=`aws cloudformation describe-stacks --stack-name igor | \
    jq '.Stacks[0].Outputs[] | select(.OutputKey=="cognitoRuleArn") | .OutputValue' -r`

aws cognito-identity set-identity-pool-roles \
    --identity-pool-id `readFromSettings GENERATED_COGNITO_POOL_ID` \
    --roles authenticated=$roleArn

cd $ROOT_APP_DIR
make deploy-site