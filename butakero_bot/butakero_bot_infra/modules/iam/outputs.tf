output "ecs_task_execution_role_arn" {
  description = "ARN del rol de ejecución de tareas"
  value       = aws_iam_role.ecs_task_execution_role.arn
}

output "ecs_task_role_arn" {
  description = "ARN del rol de tarea"
  value       = aws_iam_role.ecs_task_role.arn
}