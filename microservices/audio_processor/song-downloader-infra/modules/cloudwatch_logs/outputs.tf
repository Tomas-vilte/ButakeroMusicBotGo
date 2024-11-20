output "cloudwatch_log_group_name" {
  value = aws_cloudwatch_log_group.ecs.name
}

output "cloudwatch_log_group_arn" {
  value = aws_cloudwatch_log_group.ecs.arn
}
