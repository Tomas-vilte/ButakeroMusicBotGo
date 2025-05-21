output "security_group_ecs_id" {
  value = aws_security_group.ecs.id
}

output "security_group_alb_id" {
  description = "ID del Security Group de ECS"
  value       = aws_security_group.alb.id
}
