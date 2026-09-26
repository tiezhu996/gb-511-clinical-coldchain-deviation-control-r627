package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel provides stable identifiers and optimistic locking for each aggregate.
type BaseModel struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Code        string         `json:"code" gorm:"size:64;uniqueIndex;not null"`
	Name        string         `json:"name" gorm:"size:160;not null"`
	Status      string         `json:"status" gorm:"size:40;index;not null"`
	Version     uint           `json:"version" gorm:"not null;default:1"`
	Description string         `json:"description" gorm:"size:1000"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (b *BaseModel) BeforeCreate(_ *gorm.DB) error {
	if b.Version == 0 {
		b.Version = 1
	}
	return nil
}

type DomainRecord interface{ GetBase() *BaseModel }

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"size:80;uniqueIndex;not null"`
	DisplayName  string    `json:"displayName" gorm:"size:120;not null"`
	PasswordHash string    `json:"-" gorm:"size:120;not null"`
	Role         string    `json:"role" gorm:"size:32;index;not null"`
	Active       bool      `json:"active" gorm:"not null;default:true"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Role struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Code        string    `json:"code" gorm:"size:32;uniqueIndex;not null"`
	Name        string    `json:"name" gorm:"size:80;not null"`
	Permissions string    `json:"permissions" gorm:"size:1000;not null"`
	Active      bool      `json:"active" gorm:"not null;default:true"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type AuditLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	RequestID   string    `json:"requestId" gorm:"size:64;index"`
	Actor       string    `json:"actor" gorm:"size:80;index"`
	Action      string    `json:"action" gorm:"size:80;index"`
	EntityType  string    `json:"entityType" gorm:"size:80;index"`
	EntityID    uint      `json:"entityId" gorm:"index"`
	BeforeState string    `json:"beforeState" gorm:"size:40"`
	AfterState  string    `json:"afterState" gorm:"size:40"`
	Detail      string    `json:"detail" gorm:"size:2000"`
	CreatedAt   time.Time `json:"createdAt" gorm:"index"`
}

type AuditActionCount struct {
	Action string `json:"action" gorm:"column:action"`
	Count  int64  `json:"count" gorm:"column:count"`
}

type AuditEntityCount struct {
	EntityType string `json:"entityType" gorm:"column:entity_type"`
	Count      int64  `json:"count" gorm:"column:count"`
}

type AuditSummary struct {
	Since        time.Time          `json:"since"`
	Total        int64              `json:"total"`
	Transitions  int64              `json:"transitions"`
	UniqueActors int64              `json:"uniqueActors"`
	Actions      []AuditActionCount `json:"actions"`
	EntityTypes  []AuditEntityCount `json:"entityTypes"`
}

const (
	RoleViewer   = "viewer"
	RoleOperator = "operator"
	RoleReviewer = "reviewer"
	RoleAdmin    = "admin"
)
