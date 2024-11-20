resource "aws_sqs_queue" "main" {
  name = "${var.project_name}-queue-${var.environment}"

  delay_seconds = 0
  max_message_size = 262144
  message_retention_seconds = 345600
  visibility_timeout_seconds = 60

  tags = var.tags_sqs_queue
}
