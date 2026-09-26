package repository

import (
	"context"
	"strings"

	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/dto"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/model"
	"gorm.io/gorm"
)

type SensorEvidenceRepository interface {
	List(context.Context, dto.PageQuery) (Page[model.SensorEvidence], error)
	Get(context.Context, uint) (model.SensorEvidence, error)
	Create(context.Context, *model.SensorEvidence, *model.AuditLog) error
	CountForExcursion(context.Context, string) (int64, error)
	ReferencesExist(context.Context, string, string) (bool, error)
}

type sensorEvidenceRepository struct{ db *gorm.DB }

func NewSensorEvidenceRepository(db *gorm.DB) SensorEvidenceRepository {
	return &sensorEvidenceRepository{db: db}
}

func (r *sensorEvidenceRepository) List(ctx context.Context, query dto.PageQuery) (Page[model.SensorEvidence], error) {
	page, pageSize := normalizePage(query.Page, query.PageSize)
	db := r.db.WithContext(ctx).Model(&model.SensorEvidence{})
	if search := strings.TrimSpace(query.Search); search != "" {
		like := "%" + search + "%"
		db = db.Where("code ILIKE ? OR excursion_code ILIKE ? OR container_code ILIKE ?", like, like, like)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return Page[model.SensorEvidence]{}, err
	}
	items := make([]model.SensorEvidence, 0)
	err := db.Order("captured_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return Page[model.SensorEvidence]{Items: items, Total: total, Page: page, PageSize: pageSize}, err
}

func (r *sensorEvidenceRepository) Get(ctx context.Context, id uint) (model.SensorEvidence, error) {
	var item model.SensorEvidence
	err := r.db.WithContext(ctx).First(&item, id).Error
	return item, err
}

func (r *sensorEvidenceRepository) Create(ctx context.Context, item *model.SensorEvidence, audit *model.AuditLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(item).Error; err != nil {
			return err
		}
		audit.EntityID = item.ID
		return tx.Create(audit).Error
	})
}

func (r *sensorEvidenceRepository) CountForExcursion(ctx context.Context, code string) (int64, error) {
	var total int64
	return total, r.db.WithContext(ctx).Model(&model.SensorEvidence{}).Where("excursion_code = ?", code).Count(&total).Error
}

func (r *sensorEvidenceRepository) ReferencesExist(ctx context.Context, excursion, container string) (bool, error) {
	var excursions, containers int64
	if err := r.db.WithContext(ctx).Model(&model.ExcursionEvent{}).Where("code = ? AND container_code = ?", excursion, container).Count(&excursions).Error; err != nil {
		return false, err
	}
	if err := r.db.WithContext(ctx).Model(&model.TransportContainer{}).Where("code = ?", container).Count(&containers).Error; err != nil {
		return false, err
	}
	return excursions == 1 && containers == 1, nil
}
