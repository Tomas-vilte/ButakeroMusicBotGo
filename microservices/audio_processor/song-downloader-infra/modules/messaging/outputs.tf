output "status_queue_url" {
  description = "URL de la cola SQS para estados de descarga"
  value       = aws_sqs_queue.download_status.url
}

output "requests_queue_url" {
  description = "URL de la cola SQS para solicitudes de descarga"
  value       = aws_sqs_queue.download_requests.url
}

output "queues_arn" {
  description = "ARNs de las colas SQS"
  value       = [
    aws_sqs_queue.download_status.arn,
    aws_sqs_queue.download_requests.arn,
  ]
}