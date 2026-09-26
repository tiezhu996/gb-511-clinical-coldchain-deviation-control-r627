package repository

import (
	"context"

	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/dto"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/model"
	"gorm.io/gorm"
)

// TemperatureWindowRepository owns all persistence operations for 温控规则.
type TemperatureWindowRepository interface {
	List(context.Context, dto.PageQuery) (Page[model.TemperatureWindow], error)
	Get(context.Context, uint) (model.TemperatureWindow, error)
	Create(context.Context, *model.TemperatureWindow) error
	Update(context.Context, uint, uint, *model.TemperatureWindow) error
	Delete(context.Context, uint) error
	CountByStatus(context.Context) (map[string]int64, error)
}

type temperatureWindowRepository struct {
	store *Store[model.TemperatureWindow]
}

func NewTemperatureWindowRepository(db *gorm.DB) TemperatureWindowRepository {
	return &temperatureWindowRepository{store: NewStore[model.TemperatureWindow](db)}
}

func (r *temperatureWindowRepository) List(ctx context.Context, q dto.PageQuery) (Page[model.TemperatureWindow], error) {
	return r.store.List(ctx, q)
}
func (r *temperatureWindowRepository) Get(ctx context.Context, id uint) (model.TemperatureWindow, error) {
	return r.store.Get(ctx, id)
}
func (r *temperatureWindowRepository) Create(ctx context.Context, item *model.TemperatureWindow) error {
	return r.store.Create(ctx, item)
}
func (r *temperatureWindowRepository) Update(ctx context.Context, id, version uint, item *model.TemperatureWindow) error {
	return r.store.Update(ctx, id, version, item)
}
func (r *temperatureWindowRepository) Delete(ctx context.Context, id uint) error {
	return r.store.Delete(ctx, id)
}
func (r *temperatureWindowRepository) CountByStatus(ctx context.Context) (map[string]int64, error) {
	return r.store.CountByStatus(ctx)
}
