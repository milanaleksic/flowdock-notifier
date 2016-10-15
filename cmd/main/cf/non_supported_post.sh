#!/bin/bash

. "$( dirname "${BASH_SOURCE[0]}" )"/_commons.sh

roleArn=`aws cloudformation describe-stacks --stack-name igor | \
    jq '.Stacks[0].Outputs[] | select(.OutputKey=="cognitoRuleArn") | .OutputValue' -r`

aws cognito-identity set-identity-pool-roles \
    --identity-pool-id `readIdentityPoolIdFromSettings` \
    --roles authenticated=$roleArn