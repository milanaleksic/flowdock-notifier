#!/bin/bash

. "$( dirname "${BASH_SOURCE[0]}" )"/_commons.sh

# aws cognito-identity list-identities \
    # --identity-pool-id `readFromSettings GENERATED_COGNITO_POOL_ID` \
    # --max-results 10

aws cognito-identity describe-identity \
    --identity-id 'eu-west-1:ed4e9f7d-380d-4865-8ea2-3bc98dfeaa46' 