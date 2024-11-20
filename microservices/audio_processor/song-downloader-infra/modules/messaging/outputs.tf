output "queue_url" {
  description = "URL de la cola SQS"
  value       = aws_sqs_queue.main.url
}

output "queue_arn" {
  description = "ARN de la cola SQS"
  value       = aws_sqs_queue.main.arn
}