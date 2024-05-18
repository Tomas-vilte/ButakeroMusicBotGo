variable "lambda_execution_event_role_arn_invoke" {
  type = string
}

variable "lambda_execution_message_role_arn_invoke" {
  type = string
}


variable "region" {
  description = "La región de AWS"
  type        = string
  default     = "us-east-1"
}
