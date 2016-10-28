#!/bin/bash

. "$( dirname "${BASH_SOURCE[0]}" )"/../_commons.sh

SCRIPT_DIR=`dirname "${BASH_SOURCE[0]}"`
THIS_DIR_RELATIVE_TO_ROOT=`realpath --relative-to=$ROOT_APP_DIR $(realpath $SCRIPT_DIR)`

aws s3api put-bucket-policy --bucket $BUCKET_FOR_LOGS --policy file:///data/$THIS_DIR_RELATIVE_TO_ROOT/cloudwatch_exporting.policy

aws logs create-export-task \
    --task-name "igor-export-$RANDOM" \
    --log-group-name "/aws/lambda/igor" \
    --from `date --date="2 days ago" +%s000` \
    --to `date +%s000` \
    --destination $BUCKET_FOR_LOGS \
    --destination-prefix igor-export-task-output