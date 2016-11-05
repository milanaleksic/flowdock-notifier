#!/bin/bash

. "$( dirname "${BASH_SOURCE[0]}" )"/_commons.sh

echo "Creating Cognito Identity Pool (not supported cuurently by CloudFormation)"
identityPoolId=`aws cognito-identity create-identity-pool \
      --identity-pool-name igorCognitoPool \
      --no-allow-unauthenticated-identities \
      --supported-login-providers \
          accounts.google.com=$GOOGLE_OAUTH2_CLIENT_ID \
  | $ROOT_APP_DIR/jq '.IdentityPoolId' -r`

echo "Setting Cognito Identity Pool id into personal.env"
sed -i "s/GENERATED_COGNITO_POOL_ID=.*\$/GENERATED_COGNITO_POOL_ID=$identityPoolId/g" \
  $ROOT_APP_DIR/personal.env

cat <<END
Requesting SSL certificate from AWS ACM. Note: although certificates are supported in CloudFormation, 
we can't use any other region than us-east-1 because we utilize CloudFront so CloudFormation
free region selection can't work (more info: https://docs.aws.amazon.com/acm/latest/userguide/acm-regions.html)
END
idempotencyToken=$(date +%s)
certificateArn=`aws acm request-certificate --domain-name $CLOUDFRONT_DOMAIN \
	--region us-east-1 \
	--idempotency-token "$idempotencyToken" \
  | $ROOT_APP_DIR/jq '.CertificateArn' -r \
  | sed 's/\\//\\\\\\//g'`

echo "Setting certificate ARN into personal.env. Make sure you approve this certificate request in your email!"
sed -i "s/GENERATED_CERTIFICATE_ARN=.*\$/GENERATED_CERTIFICATE_ARN=$certificateArn/g" \
   $ROOT_APP_DIR/personal.env
