resource "aws_secretsmanager_secret" "audio_service_secrets" {
  name = "${var.project_name}-${var.environment}-${var.secret_name}"
  description = "Secretos para ${var.project_name} en ${var.environment}"

  tags = merge(
    var.tags,
    {
      Name = "${var.project_name}-${var.environment}-${var.secret_name}"
    }
  )
}

resource "aws_secretsmanager_secret_version" "audio_service_secrets_version" {
  secret_id = aws_secretsmanager_secret.audio_service_secrets.id

  secret_string = jsonencode(var.secret_values)
}