#!/bin/bash

. "$( dirname "${BASH_SOURCE[0]}" )"/_commons.sh

echo "Setting CloudFormation-generated Role as Cognito's Identity Pool Role"
roleArn=`aws cloudformation describe-stacks --stack-name igor | \
    $ROOT_APP_DIR/jq '.Stacks[0].Outputs[] | select(.OutputKey=="cognitoRuleArn") | .OutputValue' -r`

aws cognito-identity set-identity-pool-roles \
    --identity-pool-id `readFromSettings GENERATED_COGNITO_POOL_ID` \
    --roles authenticated=$roleArn

echo "Setting CloudFront Distribution's id into personal.env"
distributionId=`aws cloudformation describe-stacks --stack-name igor | \
    $ROOT_APP_DIR/jq '.Stacks[0].Outputs[] | select(.OutputKey=="distributionId") | .OutputValue' -r`

sed -i "s/DISTRIBUTION_ID=.*\$/DISTRIBUTION_ID=$distributionId/g" \
   $ROOT_APP_DIR/personal.env

echo "Reading CloudFront distribution domain name for information purposes only"
distributionDomainName=`aws cloudformation describe-stacks --stack-name igor | \
    $ROOT_APP_DIR/jq '.Stacks[0].Outputs[] | select(.OutputKey=="distributionDomainName") | .OutputValue' -r`

echo "Information: Please setup your DNS to point your $CLOUDFRONT_DOMAIN to $distributionDomainName, since this script will not use Route53"

cd $ROOT_APP_DIR

echo "Executing site deployment and CloudFront invalidation"
make site-deploy

echo "Requesting CloudFront invalidation (will take awhile, will not wait for ending of it)"
make site-invalidate