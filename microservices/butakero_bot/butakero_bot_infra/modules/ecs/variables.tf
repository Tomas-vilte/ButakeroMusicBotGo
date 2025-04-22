variable "cluster_name" {
  description = "Nombre del cluster ECS"
  type        = string
}

variable "music_bot_image" {
  description = "Imagen de Docker para el bot de música"
  type        = string
}

variable "cpu" {
  description = "Cantidad de CPU para el bot de música"
  type        = string
  default     = "256"
}

variable "memory" {
  description = "Cantidad de memoria para el bot de música"
  type        = string
  default     = "512"
}

variable "desired_count" {
  description = "Número de tareas deseadas para el servicio ECS"
  type        = number
  default     = 1
}

variable "public_subnet_ids" {
  description = "Lista de subnets donde se desplegará el bot de música"
  type        = list(string)
}

variable "security_group_id" {
  description = "ID del security group para el bot de música"
  type        = string
}

variable "ecs_task_execution_role_arn" {
  description = "ARN del rol de ejecución de ECS"
  type        = string
}

variable "aws_region" {
  description = "Región de AWS"
  type        = string
  default     = "us-east-1"
}

variable "aws_secret_name" {
  description = "Nombre del secreto de AWS"
  type        = string
}

variable "ecs_task_role_arn" {
  description = "ARN del rol de tarea de ECS"
  type        = string
}