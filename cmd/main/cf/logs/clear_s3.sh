#!/bin/bash

. "$( dirname "${BASH_SOURCE[0]}" )"/../_commons.sh

aws s3 rm --recursive s3://$BUCKET_FOR_LOGS/igor-export-task-output/
