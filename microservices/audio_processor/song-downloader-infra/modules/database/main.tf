resource "aws_dynamodb_table" "songs" {
  name = "${var.project_name}-songs-${var.environment}"
  billing_mode = "PAY_PER_REQUEST"
  hash_key       = "PK"
  range_key      = "SK"
  

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
    name = "url_youtube"
    type = "S"
  }

  global_secondary_index {
    name = "GSI2-title-index"
    hash_key = "title"
    projection_type = "ALL"
    read_capacity = 5
    write_capacity = 5
  }

  global_secondary_index {
    name = "GSI1-url-youtube-index"
    hash_key = "url_youtube"
    projection_type = "ALL"
    read_capacity = 5
    write_capacity = 5
  }

  tags = var.dynamodb_table_songs_tag
}

resource "aws_dynamodb_table" "operations" {
  name = "${var.project_name}-operations-${var.environment}"
  billing_mode = "PAY_PER_REQUEST"
  hash_key       = "PK"
  range_key      = "SK"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  attribute {
    name = "SongID"
    type = "S"
  }

  global_secondary_index {
    name = "SongIDIndex"
    hash_key = "SongID"
    projection_type = "ALL"
    read_capacity = 5
    write_capacity = 5
  }

  tags = var.dynamodb_table_operations_tags
}