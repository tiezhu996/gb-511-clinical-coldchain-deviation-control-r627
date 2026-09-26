package model

import "time"

// SensorEvidence is immutable metadata for a raw logger artifact stored in MinIO.
type SensorEvidence struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Code          string    `json:"code" gorm:"size:64;uniqueIndex;not null"`
	ExcursionCode string    `json:"excursionCode" gorm:"size:64;index;not null"`
	ContainerCode string    `json:"containerCode" gorm:"size:64;index;not null"`
	ObjectKey     string    `json:"objectKey" gorm:"size:500;uniqueIndex;not null"`
	SHA256        string    `json:"sha256" gorm:"size:64;not null"`
	MediaType     string    `json:"mediaType" gorm:"size:100;not null"`
	SizeBytes     int64     `json:"sizeBytes" gorm:"not null"`
	CapturedAt    time.Time `json:"capturedAt" gorm:"index;not null"`
	CapturedBy    string    `json:"capturedBy" gorm:"size:80;not null"`
	Source        string    `json:"source" gorm:"size:80;not null"`
	CreatedAt     time.Time `json:"createdAt" gorm:"index"`
	UploadURL     string    `json:"uploadUrl,omitempty" gorm:"-"`
}

func (SensorEvidence) TableName() string { return "sensor_evidence" }
