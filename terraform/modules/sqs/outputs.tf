output "queue_url" {
  value = aws_sqs_queue.event_queue.id
}