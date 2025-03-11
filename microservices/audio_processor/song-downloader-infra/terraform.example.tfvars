aws_region = "us-east-1"
project_name = "butakero-music-download"
environment = "prod"
gin_mode = "release"
service_max_attempts = 5
service_timeout = 2
youtube_api_key = ""
oauth2_enabled = "true"
container_port = 8080
secret_name = "butakero-audio-service-secrets"

alb_tags = {
  Project     = "music-downloader"
  Environment = "production"
}

ecs_tags = {
  Project     = "music-downloader"
  Environment = "production"
}

storage_s3_tags = {
  Project     = "music-downloader"
  Environment = "production"
}

dynamodb_table_songs_tag = {
  Project     = "music-downloader"
  Environment = "production"
}

sqs_queue_tag = {
  Project     = "music-downloader"
  Environment = "production"
}

ecr_tags = {
  Project     = "music-downloader"
  Environment = "production"
}

networking_tags = {
  Project     = "music-downloader"
  Environment = "production"
}

cloudwatch_tags = {
  Project     = "music-downloader"
  Environment = "production"
}

iam_tags = {
  Project     = "music-downloader"
  Environment = "production"
}

security_group_tags = {
  Project     = "music-downloader"
  Environment = "production"
}

sm_tags = {
  Project     = "music-downloader"
  Environment = "production"
}