#!/bin/bash

set -ea

if [ $# -ne 1 ]; then
    echo 'Usage: ./deploy.sh [ENVIRONMENT]'
    exit 1
fi

API_KEY=${API_KEY:?"You must set a API_KEY environment variable"}
PAGERDUTY_API_KEY=${PAGERDUTY_API_KEY:?"You must set a PAGERDUTY_API_KEY environment variable"}

ENVIRONMENT=$1
if [ -z "$ENVIRONMENT" ]; then
    echo "Deploy to kubernetes using './deploy.sh [environment]'"
    echo "For example 'kubernetes/deploy.sh production"
    exit 1
fi


# Set env variables based on environment
if [ "$ENVIRONMENT" == "production" ]; then
    DEBUG="false"
elif [ "$ENVIRONMENT" == "development" ]; then
    DEBUG="true"
else
    echo 'Only deployment to production or development is supported at the moment'
    exit 1;
fi

IMAGE_TAG=$(git rev-parse --short HEAD)
SLACK_API_KEY=$(echo -n $API_KEY | base64)
PAGERDUTY_API_KEY=$(echo -n $PAGERDUTY_API_KEY | base64)
expenv < ../kubernetes/resources.yml | kubectl --namespace=molliebot-${ENVIRONMENT} apply -f -
