resource "aws_dynamodb_table" "songs" {
  name         = "${var.project_name}-songs-${var.environment}"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "PK"
  range_key    = "SK"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  attribute {
    name = "title"
    type = "S"
  }

  attribute {
    name = "url"
    type = "S"
  }

  attribute {
    name = "created_at"
    type = "S"
  }

  attribute {
    name = "updated_at"
    type = "S"
  }

  attribute {
    name = "status"
    type = "S"
  }

  attribute {
    name = "message"
    type = "S"
  }

  attribute {
    name = "play_count"
    type = "N"
  }

  attribute {
    name = "attempts"
    type = "N"
  }

  attribute {
    name = "failures"
    type = "N"
  }

  attribute {
    name = "processing_date"
    type = "S"
  }

  attribute {
    name = "success"
    type = "BOOL"
  }

  global_secondary_index {
    name            = "TitleIndex"
    hash_key        = "title"
    projection_type = "ALL"
  }

  tags = var.dynamodb_table_songs_tag
}