#!/bin/bash

. "$( dirname "${BASH_SOURCE[0]}" )"/../_commons.sh

SCRIPT_DIR=`dirname "${BASH_SOURCE[0]}"`
THIS_DIR_RELATIVE_TO_ROOT=`realpath --relative-to=$ROOT_APP_DIR $(realpath $SCRIPT_DIR)`

rm -rf $SCRIPT_DIR/temp/
aws s3 sync s3://$BUCKET_FOR_LOGS/igor-export-task-output/ /data/$THIS_DIR_RELATIVE_TO_ROOT/temp/

find $SCRIPT_DIR -name '*.gz' | xargs gunzip