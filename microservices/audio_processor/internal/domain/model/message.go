package model

import "time"

type (
	// Message seria la estructura de un mensaje
	Message struct {
		ID            string
		Content       string
		Status        Status
		ReceiptHandle string
	}

	Status struct {
		ID             string    `json:"id"`
		SK             string    `json:"sk"`
		Status         string    `json:"status"`
		Message        string    `json:"message"`
		Metadata       *Metadata `json:"metadata"`
		FileData       *FileData `json:"file_data"`
		ProcessingDate time.Time `json:"processing_date"`
		Success        bool      `json:"success"`
		Attempts       int       `json:"attempts"`
		Failures       int       `json:"failures"`
	}
)
