# API Gateway REST API
resource "aws_api_gateway_rest_api" "github_webhook_api" {
  name = "GithubWebhookAPI"
  description = "API Gateway para manejar webhooks de Github"
}

# API Gateway Resource for Event Processor
resource "aws_api_gateway_resource" "event_resource" {
  rest_api_id = aws_api_gateway_rest_api.github_webhook_api.id
  parent_id   = aws_api_gateway_rest_api.github_webhook_api.root_resource_id
  path_part   = "github-webhook"
}

# API Gateway Resource for Message Processor
resource "aws_api_gateway_resource" "message_resource" {
  rest_api_id = aws_api_gateway_rest_api.github_webhook_api.id
  parent_id   = aws_api_gateway_rest_api.github_webhook_api.root_resource_id
  path_part   = "message"
}

# API Gateway Method for Event Processor (POST)
resource "aws_api_gateway_method" "event_method" {
  rest_api_id   = aws_api_gateway_rest_api.github_webhook_api.id
  resource_id   = aws_api_gateway_resource.event_resource.id
  http_method   = "POST"
  authorization = "NONE"
}

# API Gateway Method for Message Processor (POST)
resource "aws_api_gateway_method" "message_method" {
  rest_api_id   = aws_api_gateway_rest_api.github_webhook_api.id
  resource_id   = aws_api_gateway_resource.message_resource.id
  http_method   = "POST"
  authorization = "NONE"
}

# Integration with Event Processor Lambda
resource "aws_api_gateway_integration" "event_integration" {
  rest_api_id = aws_api_gateway_rest_api.github_webhook_api.id
  resource_id = aws_api_gateway_resource.event_resource.id
  http_method = aws_api_gateway_method.event_method.http_method
  integration_http_method = "POST"
  type        = "AWS_PROXY"
  uri         = var.lambda_execution_event_role_arn_invoke
}

# Integration with Message Processor Lambda
resource "aws_api_gateway_integration" "message_integration" {
  rest_api_id = aws_api_gateway_rest_api.github_webhook_api.id
  resource_id = aws_api_gateway_resource.message_resource.id
  http_method = aws_api_gateway_method.message_method.http_method
  integration_http_method = "POST"
  type        = "AWS_PROXY"
  uri         = var.lambda_execution_message_role_arn_invoke
}


# Deploy the API
resource "aws_api_gateway_deployment" "api_deployment" {
  depends_on = [aws_api_gateway_integration.event_integration, aws_api_gateway_integration.message_integration]

  rest_api_id = aws_api_gateway_rest_api.github_webhook_api.id
  stage_name  = "prod"
}

