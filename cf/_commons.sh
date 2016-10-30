#!/bin/bash

export SCRIPTS_DIR=$( dirname "${BASH_SOURCE[0]}" )
export ROOT_APP_DIR=$(realpath $SCRIPTS_DIR/../)
. $ROOT_APP_DIR/personal.env

function aws() {
    docker run -i --rm \
    -v $ROOT_APP_DIR:/data \
    --env AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
    --env AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
    -u $UID \
    garland/aws-cli-docker \
    aws --region $AWS_REGION $*
}

function readFromSettings() {
    fgrep $1 personal.env | awk 'BEGIN { FS = "=" } ; { print $2 }'
}