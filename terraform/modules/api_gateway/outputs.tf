output "api_gateway_rest_api_id" {
  description = "ID de la API gateway REST API"
  value       = aws_api_gateway_rest_api.github_webhook_api.id
}

output "api_execution_arn" {
  value = aws_api_gateway_rest_api.github_webhook_api.execution_arn
}