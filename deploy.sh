#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

gcloud run deploy moneysaver \
  --project $PROJECT_ID \
  --image ${LOCATION}-docker.pkg.dev/${PROJECT_ID}/containers/moneysaver:latest \
  --cpu 1 \
  --max-instances 1 \
  --memory 100Mi \
  --platform managed \
  --port 8080 \
  --service-account $SERVICE_ACCOUNT \
  --timeout 10s \
  --set-env-vars SLACK_BOT_TOKEN=$SLACK_BOT_TOKEN \
  --allow-unauthenticated \
  --region ${LOCATION}

