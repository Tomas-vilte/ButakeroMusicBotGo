variable "aws_region" {
  description = "region de aws"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Nombre del proyecto"
  type        = string
  default     = "music-downloader"
}

variable "environment" {
  description = "Ambiente de despliegue"
  type        = string
}

variable "youtube_api_key" {
  description = "API KEY para Youtube"
  type        = string
}

variable "oauth2_enabled" {
  description = "Habilitar OAuth2"
  type        = string
  default     = "false"
}


variable "container_port" {
  description = "Puerto del contenedor donde corre la aplicacion"
  type        = number
}

variable "alb_tags" {
  description = "Tags especificos para el Application Load Balancer"
  type        = map(string)
  default     = {}
}

variable "ecs_tags" {
  description = "Tags especificos para ECS"
  type        = map(string)
  default     = {}
}

variable "storage_s3_tags" {
  description = "Tags especificos para el storage en S3"
  type = map(string)
}

variable "dynamodb_table_operations_tags" {
  description = "Tags especificos para la tabla de Operaciones de Dynamodb"
  type = map(string)
}

variable "dynamodb_table_songs_tag" {
  description = "Tags especificos para la tabla de Canciones de Dynamodb"
  type = map(string)
}

variable "sqs_queue_tag" {
  description = "Tags especificos para la queue de SQS"
  type = map(string)
}

variable "ecr_tags" {
  description = "Tags especificos para ECR"
  type        = map(string)
  default     = {}
}

variable "networking_tags" {
  description = "Tags especificos para el networking"
  type = map(string)
}

variable "cloudwatch_tags" {
  description = "Tags especificos para CloudWatch"
  type        = map(string)
  default     = {}
}

variable "iam_tags" {
  description = "Tags especificos para roles y políticas IAM"
  type        = map(string)
  default     = {}
}

variable "security_group_tags" {
  description = "Tags especificos para Security Groups"
  type        = map(string)
  default     = {}
}

variable "ecs_service_desired_count" {
  description = "Numero deseado de tareas para el servicio ecs"
  type        = number
  default     = 2
}

variable "ecs_task_cpu" {
  description = "Unidades de CPU para la tarea de ECS"
  type        = string
  default     = "512"
}

variable "ecs_task_memory" {
  description = "Memoria en MB para la tarea de ECS"
  type        = string
  default     = "1024"
}

variable "ecs_min_capacity" {
  description = "Capacidad minima para el auto scaling de ecs"
  type        = number
  default     = 1
}

variable "ecs_max_capacity" {
  description = "Capacidad maxima para el auto scaling de ECS"
  type        = number
  default     = 10
}

variable "ecs_cpu_threshold" {
  description = "Umbral de CPU para el auto scaling"
  type        = number
  default     = 75
}

variable "ecs_memory_threshold" {
  description = "Umbral de memoria para el auto scaling"
  type        = number
  default     = 75
}


variable "gin_mode" {
  description = "Modo para el servidor web Gin"
  type        = string
}

variable "secret_name" {
  description = "Nombre del secreto"
  type = string
}

variable "service_max_attempts" {
  type = number
}

variable "service_timeout" {
  type = number
}

variable "sm_tags" {
  description = "Tags especificos para el Secret Manager"
  type = map(string)
}