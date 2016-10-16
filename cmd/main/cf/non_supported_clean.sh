#!/bin/bash

. "$( dirname "${BASH_SOURCE[0]}" )"/_commons.sh

aws cognito-identity delete-identity-pool --identity-pool-id `readFromSettings GENERATED_COGNITO_POOL_ID`
