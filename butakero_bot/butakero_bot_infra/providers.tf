terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "song-download-tf-state"
    key            = "music-bot/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform_lock_table_music_bot"
    encrypt        = true
  }
}

provider "aws" {
  region = var.aws_region
}