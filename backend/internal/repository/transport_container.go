package repository

import (
	"context"
	"time"

	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/dto"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/model"
	"gorm.io/gorm"
)

// TransportContainerRepository owns all persistence operations for 运输容器.
type TransportContainerRepository interface {
	List(context.Context, dto.PageQuery) (Page[model.TransportContainer], error)
	Get(context.Context, uint) (model.TransportContainer, error)
	GetByCode(context.Context, string) (model.TransportContainer, error)
	Create(context.Context, *model.TransportContainer) error
	Update(context.Context, uint, uint, *model.TransportContainer) error
	Delete(context.Context, uint) error
	CountByStatus(context.Context) (map[string]int64, error)
	// MoveToQuarantine atomically transitions a non-quarantined container into
	// quarantine and appends the audit log. It returns moved=false when the
	// container is already in quarantine, so callers must not duplicate the move.
	MoveToQuarantine(ctx context.Context, id uint, actor, requestID, reason string) (moved bool, beforeStatus string, err error)
}

type transportContainerRepository struct {
	store *Store[model.TransportContainer]
}

func NewTransportContainerRepository(db *gorm.DB) TransportContainerRepository {
	return &transportContainerRepository{store: NewStore[model.TransportContainer](db)}
}

func (r *transportContainerRepository) List(ctx context.Context, q dto.PageQuery) (Page[model.TransportContainer], error) {
	return r.store.List(ctx, q)
}
func (r *transportContainerRepository) Get(ctx context.Context, id uint) (model.TransportContainer, error) {
	return r.store.Get(ctx, id)
}
func (r *transportContainerRepository) GetByCode(ctx context.Context, code string) (model.TransportContainer, error) {
	var item model.TransportContainer
	err := r.store.db.WithContext(ctx).Where("code = ?", code).First(&item).Error
	return item, err
}
func (r *transportContainerRepository) Create(ctx context.Context, item *model.TransportContainer) error {
	return r.store.Create(ctx, item)
}
func (r *transportContainerRepository) Update(ctx context.Context, id, version uint, item *model.TransportContainer) error {
	return r.store.Update(ctx, id, version, item)
}
func (r *transportContainerRepository) Delete(ctx context.Context, id uint) error {
	return r.store.Delete(ctx, id)
}
func (r *transportContainerRepository) CountByStatus(ctx context.Context) (map[string]int64, error) {
	return r.store.CountByStatus(ctx)
}

func (r *transportContainerRepository) MoveToQuarantine(ctx context.Context, id uint, actor, requestID, reason string) (bool, string, error) {
	moved := false
	var beforeStatus string
	err := r.store.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current model.TransportContainer
		if err := tx.First(&current, id).Error; err != nil {
			return err
		}
		beforeStatus = current.Status
		if beforeStatus == "quarantine" {
			return nil
		}
		result := tx.Model(&model.TransportContainer{}).
			Where("id = ? AND version = ? AND status <> ?", id, current.Version, "quarantine").
			Updates(map[string]any{"status": "quarantine", "version": current.Version + 1, "updated_at": time.Now().UTC()})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrVersionConflict
		}
		audit := &model.AuditLog{
			Actor: actor, RequestID: requestID, Action: "transition",
			EntityType: "TransportContainer", EntityID: id,
			BeforeState: beforeStatus, AfterState: "quarantine", Detail: reason,
			CreatedAt: time.Now().UTC(),
		}
		if err := tx.Create(audit).Error; err != nil {
			return err
		}
		moved = true
		return nil
	})
	return moved, beforeStatus, err
}
