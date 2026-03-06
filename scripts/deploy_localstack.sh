#!/usr/bin/env sh
set -eu

if ! command -v aws >/dev/null 2>&1; then
  echo "Error: aws CLI not found. Install AWS CLI and try again." >&2
  exit 127
fi

AWS_ENDPOINT="http://localhost:4566"
AWS_REGION="${AWS_REGION:-us-east-1}"
FUNCTION_NAME="ondefarma-api"
ROLE_ARN="arn:aws:iam::000000000000:role/lambda-role"

# LocalStack accepts any credentials, but AWS CLI requires them to be set.
export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-test}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-test}"
export AWS_DEFAULT_REGION="$AWS_REGION"
export AWS_PAGER=""

aws_local() {
  aws --endpoint-url="$AWS_ENDPOINT" --region "$AWS_REGION" "$@"
}

zip -j build/function.zip build/bootstrap >/dev/null

aws_local iam create-role \
  --role-name lambda-role \
  --assume-role-policy-document '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"lambda.amazonaws.com"},"Action":"sts:AssumeRole"}]}' \
  >/dev/null 2>&1 || true

aws_local lambda get-function --function-name "$FUNCTION_NAME" >/dev/null 2>&1 \
  && aws_local lambda update-function-code --function-name "$FUNCTION_NAME" --zip-file fileb://build/function.zip >/dev/null \
  || aws_local lambda create-function \
    --function-name "$FUNCTION_NAME" \
    --runtime provided.al2023 \
    --handler bootstrap \
    --zip-file fileb://build/function.zip \
    --role "$ROLE_ARN" \
    --environment "Variables={DATABASE_URL=${DATABASE_URL},ALLOWED_ORIGINS=${ALLOWED_ORIGINS:-http://localhost:3000},MAX_BODY_BYTES=${MAX_BODY_BYTES:-65536}}" \
    >/dev/null

aws_local lambda add-permission \
  --function-name "$FUNCTION_NAME" \
  --statement-id apigw-invoke-v2 \
  --action lambda:InvokeFunction \
  --principal apigateway.amazonaws.com \
  --source-arn "arn:aws:execute-api:us-east-1:000000000000:*/*/*/*" >/dev/null 2>&1 || true

API_ID=$(aws_local apigatewayv2 create-api \
  --name ondefarma-http-api \
  --protocol-type HTTP \
  --target "arn:aws:lambda:us-east-1:000000000000:function:${FUNCTION_NAME}" \
  --query "ApiId" \
  --output text 2>/dev/null || true)
if [ -z "$API_ID" ] || [ "$API_ID" = "None" ]; then
  API_ID=$(aws_local apigatewayv2 get-apis --query "Items[?Name=='ondefarma-http-api'].ApiId | [0]" --output text)
fi

aws_local apigatewayv2 update-api \
  --api-id "$API_ID" \
  --cors-configuration "AllowOrigins=[${ALLOWED_ORIGINS:-http://localhost:3000}],AllowMethods=[GET,POST,OPTIONS],AllowHeaders=[Content-Type,Authorization],MaxAge=86400" >/dev/null 2>&1 || true

echo "Local API URL: ${AWS_ENDPOINT}/restapis/${API_ID}/\$default/_user_request_/v1/pharmacies"
