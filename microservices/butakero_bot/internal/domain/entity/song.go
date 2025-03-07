package entity

type (
	Song struct {
		ID             string   `bson:"_id,omitempty" dynamodbav:"id"`
		SK             string   `bson:"sk" dynamodbav:"sk"`
		Status         string   `bson:"status" dynamodbav:"status"`
		Message        string   `bson:"message" dynamodbav:"message"`
		Metadata       Metadata `bson:"metadata" dynamodbav:"metadata"`
		FileData       FileData `bson:"file_data" dynamodbav:"file_data"`
		ProcessingDate string   `bson:"processing_date" dynamodbav:"processing_date"`
		Success        bool     `bson:"success" dynamodbav:"success"`
		Attempts       int      `bson:"attempts" dynamodbav:"attempts"`
		Failures       int      `bson:"failures" dynamodbav:"failures"`
		RequestedBy    string
	}

	PlayedSong struct {
		Song
		Position      int64
		RequestedBy   string
		StartPosition int64
	}
)
