#!/bin/bash

set -e
cd ../$( dirname "$0" )

echo "Usage: ./build.sh [optional: 'upload']"

GITREF=$(git rev-parse --short HEAD)
IMAGE_ID=$(docker build . 2>/dev/null | awk '/Successfully built/{print $NF}')
IMAGE_NAME='wvdeutekom/molliebot'

docker tag $IMAGE_ID $IMAGE_NAME:$GITREF
docker tag $IMAGE_ID $IMAGE_NAME:latest

if [ "$1" == "upload" ]; then
    scripts/upload.sh $IMAGE_NAME:$GITREF
    scripts/upload.sh $IMAGE_NAME:latest
fi
