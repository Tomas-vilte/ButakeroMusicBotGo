variable "project_name" {
  description = "Nombre del proyecto"
  type        = string
}

variable "environment" {
  description = "Ambiente de despliegue"
  type        = string
}

variable "aws_region" {
  description = "Region de aws"
  type = string
}

variable "ecr_repository_url" {
  description = "URL del repositorio ECR"
  type = string
}

variable "execution_role_arn" {
  description = "ARN del rol de ejecuccion"
  type = string
}

variable "task_role_arn" {
  description = "ARN del rol de tarea"
  type = string
}

variable "subnet_ids" {
  description = "IDs de las subnets"
  type = list(string)
}

variable "ecs_security_group_id" {
  description = "ID del security group para ECS"
  type = string
}

variable "tags" {
  description = "Tags para recursos"
  type        = map(string)
  default     = {}
}

variable "task_cpu" {
  description = "Cantidad de unidades de CPU que usa una tarea de ecs"
  type = string
  default = "256"
}

variable "task_memory" {
  description = "Cantidad de memoria (en mib) que usa una tarea de ecs"
  type = string
  default = "512"
}

variable "max_capacity" {
  description = "Numero maximo de tareas para escalar en ecs"
  type = number
  default = 10
}

variable "min_capacity" {
  description = "Numero minimo de tareas para escalar en ecs"
  type = number
  default = 1
}

variable "target_group_arn" {
  description = "ARN del target group para el LB"
  type = string
}

variable "container_port" {
  description = "Puerto en el que escucha el contenedor de la aplicacion"
  type = number
  default = 8080
}

variable "service_desired_count" {
  description = "Numero deseado de tareas en ejecuccion para el servicio de ecs"
  type = number
  default = 1
}

variable "cpu_threshold" {
  description = "Porcentaje de uso de CPU para activar el auto scaling"
  type = number
  default = 75
}

variable "memory_threshold" {
  description = "Porcentaje de uso de memoria para activar el auto scaling"
  type = number
  default = 75
}

variable "cloudwatch_log_group" {
  description = "Grupo de registros"
  type = string
}

variable "secret_name" {
  description = "Nombre del secreto"
  type = string
}