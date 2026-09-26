package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/constants"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/dto"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/model"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/repository"
)

type TemperatureWindowService interface {
	List(context.Context, dto.PageQuery) (repository.Page[model.TemperatureWindow], error)
	Get(context.Context, uint) (model.TemperatureWindow, error)
	Create(context.Context, dto.CreateTemperatureWindow, string, string) (model.TemperatureWindow, error)
	Update(context.Context, uint, dto.UpdateTemperatureWindow, string, string) (model.TemperatureWindow, error)
	Transition(context.Context, uint, dto.TransitionRequest, string, string) (model.TemperatureWindow, error)
	Delete(context.Context, uint, string, string) error
	StatusCounts(context.Context) (map[string]int64, error)
}

type temperatureWindowService struct {
	repository repository.TemperatureWindowRepository
	security   SecurityService
}

func NewTemperatureWindowService(repo repository.TemperatureWindowRepository, security SecurityService) TemperatureWindowService {
	return &temperatureWindowService{repository: repo, security: security}
}

func (s *temperatureWindowService) List(ctx context.Context, query dto.PageQuery) (repository.Page[model.TemperatureWindow], error) {
	return s.repository.List(ctx, query)
}

func (s *temperatureWindowService) Get(ctx context.Context, id uint) (model.TemperatureWindow, error) {
	return s.repository.Get(ctx, id)
}

func (s *temperatureWindowService) Create(ctx context.Context, input dto.CreateTemperatureWindow, actor, requestID string) (model.TemperatureWindow, error) {
	if err := validateTemperatureWindowBusinessFields(input.Code, input.Name, input.Facility, input.Owner); err != nil {
		return model.TemperatureWindow{}, err
	}
	item := model.TemperatureWindow{
		BaseModel: model.BaseModel{
			Code: strings.ToUpper(strings.TrimSpace(input.Code)), Name: strings.TrimSpace(input.Name),
			Status: model.TemperatureWindowInitialStatus, Version: 1, Description: strings.TrimSpace(input.Description),
		},
		Facility: strings.TrimSpace(input.Facility), Owner: strings.TrimSpace(input.Owner),
		Category: strings.TrimSpace(input.Category), RiskLevel: input.RiskLevel,
		MetricValue: input.MetricValue, MetricUnit: strings.TrimSpace(input.MetricUnit),
		EffectiveAt: input.EffectiveAt.UTC(), Evidence: strings.TrimSpace(input.Evidence),
		RelatedCode:    strings.ToUpper(strings.TrimSpace(input.RelatedCode)),
		ProductClass:   strings.TrimSpace(firstNonEmpty(input.ProductClass, input.Category)),
		MinimumCelsius: input.MinimumCelsius, MaximumCelsius: input.MaximumCelsius,
		MaxExcursionMinutes: input.MaxExcursionMinutes,
		QualityOwner:        strings.TrimSpace(firstNonEmpty(input.QualityOwner, input.Owner)),
	}
	if item.MaximumCelsius == 0 {
		item.MaximumCelsius = input.MetricValue
	}
	if item.MaximumCelsius <= item.MinimumCelsius {
		return model.TemperatureWindow{}, fmt.Errorf("%w: maximum temperature must exceed minimum temperature", ErrInvalidInput)
	}
	if err := s.repository.Create(ctx, &item); err != nil {
		return model.TemperatureWindow{}, fmt.Errorf("create 温控规则: %w", err)
	}
	_ = s.security.Audit(ctx, actor, requestID, "create", "TemperatureWindow", item.ID, "", item.Status, "created 温控规则")
	return item, nil
}

func (s *temperatureWindowService) Update(ctx context.Context, id uint, input dto.UpdateTemperatureWindow, actor, requestID string) (model.TemperatureWindow, error) {
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return model.TemperatureWindow{}, err
	}
	if err := validateTemperatureWindowBusinessFields(current.Code, input.Name, input.Facility, input.Owner); err != nil {
		return model.TemperatureWindow{}, err
	}
	current.Name = strings.TrimSpace(input.Name)
	current.Description = strings.TrimSpace(input.Description)
	current.Facility = strings.TrimSpace(input.Facility)
	current.Owner = strings.TrimSpace(input.Owner)
	current.Category = strings.TrimSpace(input.Category)
	current.RiskLevel = input.RiskLevel
	current.MetricValue = input.MetricValue
	current.MetricUnit = strings.TrimSpace(input.MetricUnit)
	current.EffectiveAt = input.EffectiveAt.UTC()
	current.Evidence = strings.TrimSpace(input.Evidence)
	current.RelatedCode = strings.ToUpper(strings.TrimSpace(input.RelatedCode))
	current.ProductClass = strings.TrimSpace(firstNonEmpty(input.ProductClass, input.Category))
	current.MinimumCelsius = input.MinimumCelsius
	current.MaximumCelsius = input.MaximumCelsius
	if current.MaximumCelsius == 0 {
		current.MaximumCelsius = input.MetricValue
	}
	current.MaxExcursionMinutes = input.MaxExcursionMinutes
	current.QualityOwner = strings.TrimSpace(firstNonEmpty(input.QualityOwner, input.Owner))
	current.Version = input.ExpectedVersion + 1
	current.UpdatedAt = time.Now().UTC()
	if err := s.repository.Update(ctx, id, input.ExpectedVersion, &current); err != nil {
		return model.TemperatureWindow{}, fmt.Errorf("update 温控规则: %w", err)
	}
	_ = s.security.Audit(ctx, actor, requestID, "update", "TemperatureWindow", id, current.Status, current.Status, "updated business fields")
	return s.repository.Get(ctx, id)
}

func (s *temperatureWindowService) Transition(ctx context.Context, id uint, input dto.TransitionRequest, actor, requestID string) (model.TemperatureWindow, error) {
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return model.TemperatureWindow{}, err
	}
	target := strings.TrimSpace(input.Status)
	if !constants.CanTransition(constants.TemperatureWindowTransitions, current.Status, target) {
		return model.TemperatureWindow{}, fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, current.Status, target)
	}
	before := current.Status
	current.Status = target
	current.Version = input.ExpectedVersion + 1
	current.UpdatedAt = time.Now().UTC()
	if err := s.repository.Update(ctx, id, input.ExpectedVersion, &current); err != nil {
		return model.TemperatureWindow{}, fmt.Errorf("transition 温控规则: %w", err)
	}
	if err := s.security.Audit(ctx, actor, requestID, "transition", "TemperatureWindow", id, before, target, input.Reason); err != nil {
		return model.TemperatureWindow{}, fmt.Errorf("persist transition audit: %w", err)
	}
	return s.repository.Get(ctx, id)
}

func (s *temperatureWindowService) Delete(ctx context.Context, id uint, actor, requestID string) error {
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repository.Delete(ctx, id); err != nil {
		return err
	}
	return s.security.Audit(ctx, actor, requestID, "delete", "TemperatureWindow", id, current.Status, "deleted", "soft deleted 温控规则")
}

func (s *temperatureWindowService) StatusCounts(ctx context.Context) (map[string]int64, error) {
	return s.repository.CountByStatus(ctx)
}

func validateTemperatureWindowBusinessFields(code, name, facility, owner string) error {
	if strings.TrimSpace(code) == "" || strings.TrimSpace(name) == "" || strings.TrimSpace(facility) == "" || strings.TrimSpace(owner) == "" {
		return ErrInvalidInput
	}
	return nil
}
