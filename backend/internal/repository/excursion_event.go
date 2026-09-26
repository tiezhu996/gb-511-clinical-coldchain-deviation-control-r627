package repository

import (
	"context"
	"time"

	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/dto"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/model"
	"gorm.io/gorm"
)

// ExcursionEventRepository owns all persistence operations for 偏差事件.
type ExcursionEventRepository interface {
	List(context.Context, dto.PageQuery) (Page[model.ExcursionEvent], error)
	Get(context.Context, uint) (model.ExcursionEvent, error)
	GetByCode(context.Context, string) (model.ExcursionEvent, error)
	Create(context.Context, *model.ExcursionEvent) error
	Update(context.Context, uint, uint, *model.ExcursionEvent, ...*model.AuditLog) error
	Delete(context.Context, uint) error
	CountByStatus(context.Context) (map[string]int64, error)
	// SumOpenMinutes aggregates non-closed excursion durations for one container
	// and temperature window inside [since, until]. Closed-loop events are excluded.
	SumOpenMinutes(ctx context.Context, containerCode, windowCode string, since, until time.Time) (int, error)
	CountOpen(ctx context.Context, containerCode, windowCode string, since, until time.Time) (int64, error)
}

type excursionEventRepository struct {
	store *Store[model.ExcursionEvent]
}

func NewExcursionEventRepository(db *gorm.DB) ExcursionEventRepository {
	return &excursionEventRepository{store: NewStore[model.ExcursionEvent](db)}
}

func (r *excursionEventRepository) List(ctx context.Context, q dto.PageQuery) (Page[model.ExcursionEvent], error) {
	return r.store.List(ctx, q)
}
func (r *excursionEventRepository) Get(ctx context.Context, id uint) (model.ExcursionEvent, error) {
	return r.store.Get(ctx, id)
}
func (r *excursionEventRepository) GetByCode(ctx context.Context, code string) (model.ExcursionEvent, error) {
	var item model.ExcursionEvent
	err := r.store.db.WithContext(ctx).Where("code = ?", code).First(&item).Error
	return item, err
}
func (r *excursionEventRepository) SumOpenMinutes(ctx context.Context, containerCode, windowCode string, since, until time.Time) (int, error) {
	var total int64
	db := r.store.db.WithContext(ctx).Model(&model.ExcursionEvent{}).
		Where("container_code = ? AND status <> ? AND detected_at >= ? AND detected_at <= ?",
			containerCode, "closed", since.UTC(), until.UTC())
	if windowCode != "" {
		db = db.Where("window_code = ?", windowCode)
	}
	if err := db.Select("COALESCE(SUM(duration_minutes), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return int(total), nil
}
func (r *excursionEventRepository) CountOpen(ctx context.Context, containerCode, windowCode string, since, until time.Time) (int64, error) {
	var total int64
	db := r.store.db.WithContext(ctx).Model(&model.ExcursionEvent{}).
		Where("container_code = ? AND status <> ? AND detected_at >= ? AND detected_at <= ?",
			containerCode, "closed", since.UTC(), until.UTC())
	if windowCode != "" {
		db = db.Where("window_code = ?", windowCode)
	}
	return total, db.Count(&total).Error
}
func (r *excursionEventRepository) Create(ctx context.Context, item *model.ExcursionEvent) error {
	return r.store.Create(ctx, item)
}
func (r *excursionEventRepository) Update(ctx context.Context, id, version uint, item *model.ExcursionEvent, audits ...*model.AuditLog) error {
	return r.store.Update(ctx, id, version, item, audits...)
}
func (r *excursionEventRepository) Delete(ctx context.Context, id uint) error {
	return r.store.Delete(ctx, id)
}
func (r *excursionEventRepository) CountByStatus(ctx context.Context) (map[string]int64, error) {
	return r.store.CountByStatus(ctx)
}
