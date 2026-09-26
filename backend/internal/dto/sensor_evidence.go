package dto

import "time"

type CreateSensorEvidence struct {
	Code          string    `json:"code" binding:"required,min=3,max=64"`
	ExcursionCode string    `json:"excursionCode" binding:"required,min=3,max=64"`
	ContainerCode string    `json:"containerCode" binding:"required,min=3,max=64"`
	ObjectKey     string    `json:"objectKey" binding:"required,min=5,max=500"`
	SHA256        string    `json:"sha256" binding:"required,len=64,hexadecimal"`
	MediaType     string    `json:"mediaType" binding:"required,max=100"`
	SizeBytes     int64     `json:"sizeBytes" binding:"required,min=1"`
	CapturedAt    time.Time `json:"capturedAt" binding:"required"`
	Source        string    `json:"source" binding:"required,max=80"`
}
