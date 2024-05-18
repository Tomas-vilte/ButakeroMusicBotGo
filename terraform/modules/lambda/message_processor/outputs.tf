output "message_processor_lambda_arn" {
  value = aws_lambda_function.message_processor_lambda.invoke_arn
}

output "message_processor_lambda_name" {
  value = aws_lambda_function.message_processor_lambda.function_name
}
