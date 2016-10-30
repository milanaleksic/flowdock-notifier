#!/bin/bash

. "$( dirname "${BASH_SOURCE[0]}" )"/../_commons.sh

if [[ "$1" == "" ]];
then
    echo "Please provide number of days of logs you want to schedule for downloading"
    exit 1
fi

SCRIPT_DIR=`dirname "${BASH_SOURCE[0]}"`
THIS_DIR_RELATIVE_TO_ROOT=`realpath --relative-to=$ROOT_APP_DIR $(realpath $SCRIPT_DIR)`

cat $SCRIPT_DIR/cloudwatch_exporting.policy.template \
    | sed -e "s/BUCKET_FOR_LOGS/$BUCKET_FOR_LOGS/g" > $SCRIPT_DIR/cloudwatch_exporting.policy

aws s3api put-bucket-policy --bucket $BUCKET_FOR_LOGS --policy file:///data/$THIS_DIR_RELATIVE_TO_ROOT/cloudwatch_exporting.policy

aws logs create-export-task \
    --task-name "igor-export-$RANDOM" \
    --log-group-name "/aws/lambda/igor" \
    --from `date --date="$1 days ago" +%s000` \
    --to `date +%s000` \
    --destination $BUCKET_FOR_LOGS \
    --destination-prefix igor-export-task-output