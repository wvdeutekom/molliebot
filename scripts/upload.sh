#!/bin/bash

set -e

if [ $# -ne 1 ]; then
    echo 'Usage: ./upload.sh [IMAGE_TAG]'
    exit 1
fi

IMAGE_TAG=${1:?"You must provide a IMAGE_TAG as parameter to this script"}

docker push $IMAGE_TAG
