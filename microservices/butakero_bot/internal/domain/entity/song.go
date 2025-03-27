package entity

import "time"

type (
	SongEntity struct {
		ID             string    `bson:"_id" dynamodbav:"-"`
		PK             string    `bson:"-" dynamodbav:"PK"`
		SK             string    `bson:"-" dynamodbav:"SK"`
		TitleLower     string    `bson:"title_lower" dynamodbav:"title_lower"`
		Status         string    `bson:"status" dynamodbav:"status"`
		Message        string    `bson:"message" dynamodbav:"message"`
		Metadata       Metadata  `bson:"metadata" dynamodbav:"metadata"`
		FileData       FileData  `bson:"file_data" dynamodbav:"file_data"`
		ProcessingDate time.Time `bson:"processing_date" dynamodbav:"processing_date"`
		Success        bool      `bson:"success" dynamodbav:"success"`
		Attempts       int       `bson:"attempts" dynamodbav:"attempts"`
		Failures       int       `bson:"failures" dynamodbav:"failures"`
		CreatedAt      time.Time `bson:"created_at" dynamodbav:"created_at"`
		UpdatedAt      time.Time `bson:"updated_at" dynamodbav:"updated_at"`
		PlayCount      int       `bson:"play_count" dynamodbav:"play_count"`
		GSI1PK         string    `bson:"-" dynamodbav:"GSI1PK"`
		GSI1SK         string    `bson:"-" dynamodbav:"GSI1SK"`
	}
)
