#!/bin/bash

. "$( dirname "${BASH_SOURCE[0]}" )"/_commons.sh

rootDir=$SCRIPTS_DIR/../../..

identityPoolId=`aws cognito-identity create-identity-pool \
      --identity-pool-name igorCognitoPool \
      --no-allow-unauthenticated-identities \
      --supported-login-providers \
          accounts.google.com=$GOOGLE_OAUTH2_CLIENT_ID \
  | jq '.IdentityPoolId' -r`

sed -i "s/GENERATED_COGNITO_POOL_ID=.*\$/GENERATED_COGNITO_POOL_ID=$identityPoolId/g" \
  $rootDir/personal.env