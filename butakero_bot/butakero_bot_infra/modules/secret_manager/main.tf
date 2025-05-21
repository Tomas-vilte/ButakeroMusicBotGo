data "aws_secretsmanager_secret" "existing_secret" {
  arn = var.secret_arn
}

data "aws_secretsmanager_secret_version" "existing_secret_version" {
  secret_id = data.aws_secretsmanager_secret.existing_secret.id
}

locals {
  existing_secret = jsondecode(data.aws_secretsmanager_secret_version.existing_secret_version.secret_string)
  new_secret = {
    COMMAND_PREFIX = var.command_prefix
    DISCORD_TOKEN  = var.discord_token
  }
  combined_secret = merge(local.existing_secret, local.new_secret)
}

resource "aws_secretsmanager_secret_version" "bot_secrets" {
  secret_id = var.secret_arn

  secret_string = jsonencode(local.combined_secret)
}