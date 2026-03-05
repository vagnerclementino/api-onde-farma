#!/usr/bin/env sh
set -eu

AWS_ENDPOINT="http://localhost:4566"
FUNCTION_NAME="ondefarma-api"
ROLE_ARN="arn:aws:iam::000000000000:role/lambda-role"

zip -j build/function.zip build/bootstrap >/dev/null

awslocal --endpoint-url="$AWS_ENDPOINT" iam create-role \
  --role-name lambda-role \
  --assume-role-policy-document '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"lambda.amazonaws.com"},"Action":"sts:AssumeRole"}]}' \
  >/dev/null 2>&1 || true

awslocal --endpoint-url="$AWS_ENDPOINT" lambda get-function --function-name "$FUNCTION_NAME" >/dev/null 2>&1 \
  && awslocal --endpoint-url="$AWS_ENDPOINT" lambda update-function-code --function-name "$FUNCTION_NAME" --zip-file fileb://build/function.zip >/dev/null \
  || awslocal --endpoint-url="$AWS_ENDPOINT" lambda create-function \
    --function-name "$FUNCTION_NAME" \
    --runtime provided.al2023 \
    --handler bootstrap \
    --zip-file fileb://build/function.zip \
    --role "$ROLE_ARN" \
    --environment "Variables={DATABASE_URL=${DATABASE_URL},ALLOWED_ORIGINS=${ALLOWED_ORIGINS:-http://localhost:3000},MAX_BODY_BYTES=${MAX_BODY_BYTES:-65536}}" \
    >/dev/null

awslocal --endpoint-url="$AWS_ENDPOINT" lambda add-permission \
  --function-name "$FUNCTION_NAME" \
  --statement-id apigw-invoke-v2 \
  --action lambda:InvokeFunction \
  --principal apigateway.amazonaws.com \
  --source-arn "arn:aws:execute-api:us-east-1:000000000000:*/*/*/*" >/dev/null 2>&1 || true

API_ID=$(awslocal --endpoint-url="$AWS_ENDPOINT" apigatewayv2 create-api \
  --name ondefarma-http-api \
  --protocol-type HTTP \
  --target "arn:aws:lambda:us-east-1:000000000000:function:${FUNCTION_NAME}" \
  --query "ApiId" \
  --output text 2>/dev/null || true)
if [ -z "$API_ID" ] || [ "$API_ID" = "None" ]; then
  API_ID=$(awslocal --endpoint-url="$AWS_ENDPOINT" apigatewayv2 get-apis --query "Items[?Name=='ondefarma-http-api'].ApiId | [0]" --output text)
fi

awslocal --endpoint-url="$AWS_ENDPOINT" apigatewayv2 update-api \
  --api-id "$API_ID" \
  --cors-configuration "AllowOrigins=[${ALLOWED_ORIGINS:-http://localhost:3000}],AllowMethods=[GET,POST,OPTIONS],AllowHeaders=[Content-Type,Authorization],MaxAge=86400" >/dev/null 2>&1 || true

echo "Local API URL: ${AWS_ENDPOINT}/restapis/${API_ID}/\$default/_user_request_/v1/pharmacies"
