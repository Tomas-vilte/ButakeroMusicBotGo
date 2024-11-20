output "secret_arn" {
  value = aws_secretsmanager_secret.audio_service_secrets.arn
}

output "secret_name" {
  value = aws_secretsmanager_secret.audio_service_secrets.name
}