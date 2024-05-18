variable "lambda_execution_role_arn" {
  type = string
}

resource "aws_lambda_function" "message_processor_lambda" {
  function_name    = "MessageProcessorLambda"
  runtime          = "provided.al2023"
  handler          = "main"
  filename         = "${path.module}/lambda.zip"
  source_code_hash = filebase64sha256("${path.module}/lambda.zip")
  role             = var.lambda_execution_role_arn
}
