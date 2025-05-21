output "execution_role_arn" {
  description = "ARN del rol de ejecución"
  value       = aws_iam_role.ecs_execution_role.arn
}

output "execution_role_name" {
  description = "Nombre del rol de ejecución"
  value       = aws_iam_role.ecs_execution_role.name
}

output "task_role_arn" {
  description = "ARN del rol de tarea"
  value       = aws_iam_role.ecs_task_role.arn
}

output "task_role_name" {
  description = "Nombre del rol de tarea"
  value       = aws_iam_role.ecs_task_role.name
}