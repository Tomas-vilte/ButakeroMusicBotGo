package entity

import "time"

type Job struct {
	ID                   string
	S3Key                string
	KEY                  string
	Status               string
	BucketName           string
	Region               string
	TaskDefinition       string
	ClusterName          string
	SecurityGroup        string
	Subnets              []string
	TaskRoleArn          string
	TaskExecutionRoleArn string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
