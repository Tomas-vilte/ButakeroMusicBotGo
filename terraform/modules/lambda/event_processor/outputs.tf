output "event_processor_lambda_invoke_arn" {
  value = aws_lambda_function.event_processor_lambda.invoke_arn
}

output "event_processor_lambda_name" {
  value = aws_lambda_function.event_processor_lambda.function_name
}