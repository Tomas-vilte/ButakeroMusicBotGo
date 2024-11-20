variable "project_name" {
  description = "Nombre del proyecto"
  type        = string
}

variable "environment" {
  description = "Ambiente de despliegue"
  type        = string
}

variable "s3_bucket_arns" {
  description = "ARNs de los buckets S3"
  type        = list(string)
}

variable "secrets_manager_arns" {
  description = "ARNs de los secrets"
  type = list(string)
}

variable "dynamodb_table_arns" {
  description = "ARNs de las tablas DynamoDB"
  type        = list(string)
}

variable "sqs_queue_arns" {
  description = "ARNs de las colas SQS"
  type        = list(string)
}

variable "cloudwatch_log_group_arn" {
  description = "ARN del grupo de logs de CloudWatch"
  type        = string
}

variable "tags" {
  description = "Tags para los recursos"
  type        = map(string)
  default     = {}
}