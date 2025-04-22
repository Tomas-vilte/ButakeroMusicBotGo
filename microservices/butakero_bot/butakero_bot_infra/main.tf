data "terraform_remote_state" "shared_resources" {
  backend = "s3"

  config = {
    bucket = "song-download-tf-state"
    key    = "prod/terraform.tfstate"
    region = "us-east-1"
  }
}

data "aws_caller_identity" "current" {}

module "ecs" {
  source = "./modules/ecs"
  cluster_name              = data.terraform_remote_state.shared_resources.outputs.ecs_cluster_name
  music_bot_image           = "${module.ecr.repository_url}:latest"
  cpu                       = "256"
  memory                    = "512"
  desired_count             = 1
  public_subnet_ids         = module.networking.public_subnet_ids
  security_group_id         = aws_security_group.music_bot_sg.id
  ecs_task_execution_role_arn = module.iam.ecs_task_execution_role_arn
  aws_region                = var.aws_region
  aws_secret_name           = module.secret_manager.secret_name
  ecs_task_role_arn = module.iam.ecs_task_role_arn
}

module "networking" {
  source = "./modules/networking"

  remote_state_bucket       = "song-download-tf-state"
  remote_state_key          = "prod/terraform.tfstate"
  aws_region                = "us-east-1"
}

module "iam" {
  source = "./modules/iam"
  }

module "ecr" {
  source = "./modules/ecr"

  repository_name    = "butakero-bot-prod"
  aws_region        = "us-east-1"
  aws_account_id    = data.aws_caller_identity.current.account_id
  image_tag         = "latest"
  docker_context_path = "."
}

module "secret_manager" {
  source = "./modules/secret_manager"

  secret_name     = data.terraform_remote_state.shared_resources.outputs.secret_name
  secret_arn      = data.terraform_remote_state.shared_resources.outputs.secret_arn
  command_prefix  = var.command_prefix
  discord_token   = var.discord_token
}

resource "aws_security_group" "music_bot_sg" {
  name        = "music-bot-sg"
  description = "Security Group para el bot de musica"
  vpc_id      = module.networking.vpc_id

  tags = {
    Name        = "music-bot-sg"
    Environment = "prod"
  }
}

resource "aws_security_group_rule" "music_bot_egress_all" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.music_bot_sg.id
}