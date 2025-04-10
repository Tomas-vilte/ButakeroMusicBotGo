resource "aws_sqs_queue" "download_status" {
  name = "${var.project_name}-download-status-${var.environment}"

  delay_seconds = 0
  max_message_size = 262144
  message_retention_seconds = 345600
  visibility_timeout_seconds = 60

  tags = var.tags_sqs_queue
}

resource "aws_sqs_queue" "download_requests" {
  name = "${var.project_name}-download-requests-${var.environment}"

  delay_seconds = 0
  max_message_size = 262144
  message_retention_seconds = 345600
  visibility_timeout_seconds = 60

  tags = var.tags_sqs_queue
}