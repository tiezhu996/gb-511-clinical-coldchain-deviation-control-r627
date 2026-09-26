package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/dto"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/model"
	"gorm.io/gorm"
)

var ErrVersionConflict = errors.New("record was changed by another request")

type Page[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type Store[T any] struct{ db *gorm.DB }

func NewStore[T any](db *gorm.DB) *Store[T] { return &Store[T]{db: db} }

func (s *Store[T]) List(ctx context.Context, query dto.PageQuery) (Page[T], error) {
	page, pageSize := normalizePage(query.Page, query.PageSize)
	db := s.db.WithContext(ctx).Model(new(T))
	if search := strings.TrimSpace(strings.ToLower(query.Search)); search != "" {
		wildcard := "%" + search + "%"
		db = db.Where("LOWER(code) LIKE ? OR LOWER(name) LIKE ?", wildcard, wildcard)
	}
	if status := strings.TrimSpace(query.Status); status != "" {
		db = db.Where("status = ?", status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return Page[T]{}, err
	}
	items := make([]T, 0)
	err := db.Order("updated_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return Page[T]{Items: items, Total: total, Page: page, PageSize: pageSize}, err
}

func (s *Store[T]) Get(ctx context.Context, id uint) (T, error) {
	var item T
	err := s.db.WithContext(ctx).First(&item, id).Error
	return item, err
}
func (s *Store[T]) Create(ctx context.Context, item *T) error {
	return s.db.WithContext(ctx).Create(item).Error
}
func (s *Store[T]) Update(ctx context.Context, id, expectedVersion uint, item *T, audits ...*model.AuditLog) error {
	update := func(tx *gorm.DB) error {
		result := tx.Model(new(T)).Where("id = ? AND version = ?", id, expectedVersion).Select("*").Omit("id", "code", "created_at", "deleted_at").Updates(item)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrVersionConflict
		}
		for _, audit := range audits {
			if audit != nil {
				if err := tx.Create(audit).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}
	db := s.db.WithContext(ctx)
	if len(audits) == 0 {
		return update(db)
	}
	return db.Transaction(update)
}
func (s *Store[T]) Delete(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(new(T), id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (s *Store[T]) CountByStatus(ctx context.Context) (map[string]int64, error) {
	rows, err := s.db.WithContext(ctx).Model(new(T)).Select("status, COUNT(*) AS total").Group("status").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := make(map[string]int64)
	for rows.Next() {
		var status string
		var total int64
		if err := rows.Scan(&status, &total); err != nil {
			return nil, err
		}
		counts[status] = total
	}
	return counts, rows.Err()
}
func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

type SecurityRepository interface {
	FindUserByUsername(context.Context, string) (model.User, error)
	CreateUser(context.Context, *model.User) error
	CountUsers(context.Context) (int64, error)
	AppendAudit(context.Context, *model.AuditLog) error
	ListAudits(context.Context, int, int, string) ([]model.AuditLog, int64, error)
	SummarizeAudits(context.Context, time.Time) (model.AuditSummary, error)
	EntityHistory(context.Context, string, uint, int) ([]model.AuditLog, error)
}

type securityRepository struct{ db *gorm.DB }

func NewSecurityRepository(db *gorm.DB) SecurityRepository {
	return &securityRepository{db: db}
}

func (r *securityRepository) FindUserByUsername(ctx context.Context, username string) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Joins("JOIN roles ON roles.code = users.role AND roles.active = ?", true).
		Where("users.username = ? AND users.active = ?", username, true).First(&user).Error
	return user, err
}

func (r *securityRepository) CreateUser(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *securityRepository) CountUsers(ctx context.Context) (int64, error) {
	var total int64
	return total, r.db.WithContext(ctx).Model(&model.User{}).Count(&total).Error
}

func (r *securityRepository) AppendAudit(ctx context.Context, log *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *securityRepository) ListAudits(ctx context.Context, page, pageSize int, search string) ([]model.AuditLog, int64, error) {
	page, pageSize = normalizePage(page, pageSize)
	db := r.db.WithContext(ctx).Model(&model.AuditLog{})
	if search != "" {
		wildcard := "%" + search + "%"
		db = db.Where("actor LIKE ? OR entity_type LIKE ? OR action LIKE ?", wildcard, wildcard, wildcard)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	logs := make([]model.AuditLog, 0)
	err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

func (r *securityRepository) SummarizeAudits(ctx context.Context, since time.Time) (model.AuditSummary, error) {
	summary := model.AuditSummary{
		Since: since.UTC(), Actions: make([]model.AuditActionCount, 0),
		EntityTypes: make([]model.AuditEntityCount, 0),
	}
	base := r.db.WithContext(ctx).Model(&model.AuditLog{}).Where("created_at >= ?", since.UTC())
	if err := base.Count(&summary.Total).Error; err != nil {
		return model.AuditSummary{}, err
	}
	if err := r.db.WithContext(ctx).Model(&model.AuditLog{}).
		Where("created_at >= ? AND action = ?", since.UTC(), "transition").Count(&summary.Transitions).Error; err != nil {
		return model.AuditSummary{}, err
	}
	if err := r.db.WithContext(ctx).Model(&model.AuditLog{}).
		Where("created_at >= ?", since.UTC()).Distinct("actor").Count(&summary.UniqueActors).Error; err != nil {
		return model.AuditSummary{}, err
	}
	if err := r.db.WithContext(ctx).Model(&model.AuditLog{}).
		Select("action, COUNT(*) AS count").Where("created_at >= ?", since.UTC()).
		Group("action").Order("count DESC").Scan(&summary.Actions).Error; err != nil {
		return model.AuditSummary{}, err
	}
	if err := r.db.WithContext(ctx).Model(&model.AuditLog{}).
		Select("entity_type, COUNT(*) AS count").Where("created_at >= ?", since.UTC()).
		Group("entity_type").Order("count DESC").Scan(&summary.EntityTypes).Error; err != nil {
		return model.AuditSummary{}, err
	}
	return summary, nil
}

func (r *securityRepository) EntityHistory(ctx context.Context, entityType string, entityID uint, limit int) ([]model.AuditLog, error) {
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	logs := make([]model.AuditLog, 0, limit)
	err := r.db.WithContext(ctx).Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}
