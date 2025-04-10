module "networking" {
  source = "./modules/networking"

  project_name = var.project_name
  environment  = var.environment
  tags = var.networking_tags
}

data "aws_caller_identity" "current" {}


module "storage" {
  source = "./modules/storage"

  project_name = var.project_name
  environment  = var.environment
  tags = var.storage_s3_tags
}

module "secret_manager" {
  source = "./modules/secret_manager"

  project_name = var.project_name
  environment  = var.environment
  secret_values = {
    "GIN_MODE": var.gin_mode
    "YOUTUBE_API_KEY": var.youtube_api_key
    "S3_BUCKET_NAME": module.storage.bucket_name
    "DYNAMODB_TABLE_SONGS": module.database.songs_table_name
    "SERVICE_MAX_ATTEMPTS": var.service_max_attempts
    "SERVICE_TIMEOUT": var.service_timeout
    "AUDIO_PROCESSOR_URL": "http://${module.alb.alb_dns_name}"
    SQS_BOT_DOWNLOAD_STATUS_URL: module.messaging.status_queue_url
    SQS_BOT_DOWNLOAD_REQUESTS_URL: module.messaging.requests_queue_url
    NUM_WORKERS: var.workers_count
  }
  tags = var.sm_tags
  secret_name = var.secret_name
}

module "database" {
  source = "./modules/database"
  dynamodb_table_songs_tag = var.dynamodb_table_songs_tag
  project_name = var.project_name
  environment  = var.environment
}

module "messaging" {
  source = "./modules/messaging"

  project_name = var.project_name
  environment  = var.environment
  tags_sqs_queue = var.sqs_queue_tag
}

module "ecr" {
  source = "./modules/ecr"

  project_name = var.project_name
  environment  = var.environment
  tags         = var.ecr_tags
  aws_account_id = data.aws_caller_identity.current.account_id
  docker_context_path = "."
  image_tag = "latest"
  aws_region = var.aws_region
}

module "cloudwatch_logs" {
  source = "./modules/cloudwatch_logs"

  project_name      = var.project_name
  environment       = var.environment
  retention_in_days = 30
  tags              = var.cloudwatch_tags
}

module "iam" {
  source = "./modules/iam"

  project_name = var.project_name
  environment  = var.environment

  s3_bucket_arns      = [module.storage.bucket_arn]
  sqs_queue_arns      = module.messaging.queues_arn
  secrets_manager_arns = [ module.secret_manager.secret_arn]
  dynamodb_table_arns = module.database.table_arns
  tags                = var.iam_tags

  cloudwatch_log_group_arn = module.cloudwatch_logs.cloudwatch_log_group_arn
}

module "security_groups" {
  source = "./modules/security_groups"

  project_name = var.project_name
  environment = var.environment
  vpc_id = module.networking.vpc_id
  container_port = var.container_port
  
  tags = var.security_group_tags
}

module "alb" {
  source = "./modules/alb"

  project_name = var.project_name
  environment  = var.environment

  vpc_id             = module.networking.vpc_id
  subnet_ids         = module.networking.public_subnet_ids
  tags               = var.alb_tags
  security_group_alb = module.security_groups.security_group_alb_id
  private_subnet_ids = [module.networking.private_subnet_ids[0], module.networking.private_subnet_ids[1]]

  logs_bucket = module.storage.bucket_name
}

module "ecs" {
  source = "./modules/ecs"

  project_name = var.project_name
  environment  = var.environment

  aws_region            = var.aws_region
  ecs_security_group_id = module.security_groups.security_group_ecs_id
  subnet_ids            = module.networking.public_subnet_ids

  target_group_arn      = module.alb.target_group_arn
  container_port        = var.container_port
  service_desired_count = var.ecs_service_desired_count
  task_cpu              = var.ecs_task_cpu
  task_memory           = var.ecs_task_memory
  execution_role_arn    = module.iam.execution_role_arn
  task_role_arn         = module.iam.task_role_arn
  secret_name = module.secret_manager.secret_name
  min_capacity     = var.ecs_min_capacity
  max_capacity     = var.ecs_max_capacity
  cpu_threshold    = var.ecs_cpu_threshold
  memory_threshold = var.ecs_memory_threshold
  cloudwatch_log_group  = module.cloudwatch_logs.cloudwatch_log_group_name
  ecr_repository_url    = module.ecr.repository_url

  tags = var.ecs_tags
}