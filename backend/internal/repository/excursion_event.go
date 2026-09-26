package repository

import (
	"context"
	"time"

	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/constants"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/dto"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/model"
	"gorm.io/gorm"
)

// ExcursionEventRepository owns all persistence operations for 偏差事件.
type ExcursionEventRepository interface {
	List(context.Context, dto.PageQuery) (Page[model.ExcursionEvent], error)
	Get(context.Context, uint) (model.ExcursionEvent, error)
	Create(context.Context, *model.ExcursionEvent) error
	Update(context.Context, uint, uint, *model.ExcursionEvent, ...*model.AuditLog) error
	Delete(context.Context, uint) error
	CountByStatus(context.Context) (map[string]int64, error)
	SumOpenDurationSince(context.Context, string, time.Time) (int64, error)
	CumulativeOpenDurationByContainer(context.Context, time.Time) (map[string]int64, error)
}

type excursionEventRepository struct {
	store *Store[model.ExcursionEvent]
	db    *gorm.DB
}

func NewExcursionEventRepository(db *gorm.DB) ExcursionEventRepository {
	return &excursionEventRepository{store: NewStore[model.ExcursionEvent](db), db: db}
}

func (r *excursionEventRepository) List(ctx context.Context, q dto.PageQuery) (Page[model.ExcursionEvent], error) {
	return r.store.List(ctx, q)
}
func (r *excursionEventRepository) Get(ctx context.Context, id uint) (model.ExcursionEvent, error) {
	return r.store.Get(ctx, id)
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

// SumOpenDurationSince totals the duration of one container's excursions that are
// still not closed (open/in_review/decided) and were detected since the given time.
func (r *excursionEventRepository) SumOpenDurationSince(ctx context.Context, containerCode string, since time.Time) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&model.ExcursionEvent{}).
		Where("container_code = ? AND status <> ? AND detected_at >= ?", containerCode, string(constants.ExcursionStateClosed), since.UTC()).
		Select("COALESCE(SUM(duration_minutes), 0)").Scan(&total).Error
	return total, err
}

// CumulativeOpenDurationByContainer groups the same not-closed duration total for
// every container so the workbench can show the current 24h exposure per container.
func (r *excursionEventRepository) CumulativeOpenDurationByContainer(ctx context.Context, since time.Time) (map[string]int64, error) {
	type containerTotal struct {
		ContainerCode string `gorm:"column:container_code"`
		Total         int64  `gorm:"column:total"`
	}
	rows := make([]containerTotal, 0)
	err := r.db.WithContext(ctx).Model(&model.ExcursionEvent{}).
		Select("container_code, COALESCE(SUM(duration_minutes), 0) AS total").
		Where("status <> ? AND detected_at >= ?", string(constants.ExcursionStateClosed), since.UTC()).
		Group("container_code").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	totals := make(map[string]int64, len(rows))
	for _, row := range rows {
		totals[row.ContainerCode] = row.Total
	}
	return totals, nil
}
