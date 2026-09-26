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

type TransportContainerService interface {
	List(context.Context, dto.PageQuery) (repository.Page[model.TransportContainer], error)
	Get(context.Context, uint) (model.TransportContainer, error)
	Create(context.Context, dto.CreateTransportContainer, string, string) (model.TransportContainer, error)
	Update(context.Context, uint, dto.UpdateTransportContainer, string, string) (model.TransportContainer, error)
	Transition(context.Context, uint, dto.TransitionRequest, string, string) (model.TransportContainer, error)
	Delete(context.Context, uint, string, string) error
	StatusCounts(context.Context) (map[string]int64, error)
}

type transportContainerService struct {
	repository repository.TransportContainerRepository
	security   SecurityService
}

func NewTransportContainerService(repo repository.TransportContainerRepository, security SecurityService) TransportContainerService {
	return &transportContainerService{repository: repo, security: security}
}

func (s *transportContainerService) List(ctx context.Context, query dto.PageQuery) (repository.Page[model.TransportContainer], error) {
	return s.repository.List(ctx, query)
}

func (s *transportContainerService) Get(ctx context.Context, id uint) (model.TransportContainer, error) {
	return s.repository.Get(ctx, id)
}

func (s *transportContainerService) Create(ctx context.Context, input dto.CreateTransportContainer, actor, requestID string) (model.TransportContainer, error) {
	if err := validateTransportContainerBusinessFields(input.Code, input.Name, input.Facility, input.Owner); err != nil {
		return model.TransportContainer{}, err
	}
	item := model.TransportContainer{
		BaseModel: model.BaseModel{
			Code: strings.ToUpper(strings.TrimSpace(input.Code)), Name: strings.TrimSpace(input.Name),
			Status: model.TransportContainerInitialStatus, Version: 1, Description: strings.TrimSpace(input.Description),
		},
		Facility: strings.TrimSpace(input.Facility), Owner: strings.TrimSpace(input.Owner),
		Category: strings.TrimSpace(input.Category), RiskLevel: input.RiskLevel,
		MetricValue: input.MetricValue, MetricUnit: strings.TrimSpace(input.MetricUnit),
		EffectiveAt: input.EffectiveAt.UTC(), Evidence: strings.TrimSpace(input.Evidence),
		RelatedCode:       strings.ToUpper(strings.TrimSpace(input.RelatedCode)),
		SensorID:          strings.ToUpper(strings.TrimSpace(firstNonEmpty(input.SensorID, input.RelatedCode))),
		ContainerType:     strings.TrimSpace(firstNonEmpty(input.ContainerType, input.Category)),
		CurrentLocation:   strings.TrimSpace(firstNonEmpty(input.CurrentLocation, input.Facility)),
		Custodian:         strings.TrimSpace(firstNonEmpty(input.Custodian, input.Owner)),
		CurrentTempC:      input.CurrentTempC,
		LastSensorReading: fallbackTime(input.LastSensorReading, input.EffectiveAt),
	}
	if input.CurrentTempC == 0 {
		item.CurrentTempC = input.MetricValue
	}
	if item.SensorID == "" {
		return model.TransportContainer{}, fmt.Errorf("%w: sensor id is required", ErrInvalidInput)
	}
	if err := s.repository.Create(ctx, &item); err != nil {
		return model.TransportContainer{}, fmt.Errorf("create 运输容器: %w", err)
	}
	_ = s.security.Audit(ctx, actor, requestID, "create", "TransportContainer", item.ID, "", item.Status, "created 运输容器")
	return item, nil
}

func (s *transportContainerService) Update(ctx context.Context, id uint, input dto.UpdateTransportContainer, actor, requestID string) (model.TransportContainer, error) {
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return model.TransportContainer{}, err
	}
	if err := validateTransportContainerBusinessFields(current.Code, input.Name, input.Facility, input.Owner); err != nil {
		return model.TransportContainer{}, err
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
	current.SensorID = strings.ToUpper(strings.TrimSpace(firstNonEmpty(input.SensorID, input.RelatedCode)))
	current.ContainerType = strings.TrimSpace(firstNonEmpty(input.ContainerType, input.Category))
	current.CurrentLocation = strings.TrimSpace(firstNonEmpty(input.CurrentLocation, input.Facility))
	current.Custodian = strings.TrimSpace(firstNonEmpty(input.Custodian, input.Owner))
	current.CurrentTempC = input.CurrentTempC
	if input.CurrentTempC == 0 {
		current.CurrentTempC = input.MetricValue
	}
	current.LastSensorReading = fallbackTime(input.LastSensorReading, input.EffectiveAt)
	current.Version = input.ExpectedVersion + 1
	current.UpdatedAt = time.Now().UTC()
	if err := s.repository.Update(ctx, id, input.ExpectedVersion, &current); err != nil {
		return model.TransportContainer{}, fmt.Errorf("update 运输容器: %w", err)
	}
	_ = s.security.Audit(ctx, actor, requestID, "update", "TransportContainer", id, current.Status, current.Status, "updated business fields")
	return s.repository.Get(ctx, id)
}

func (s *transportContainerService) Transition(ctx context.Context, id uint, input dto.TransitionRequest, actor, requestID string) (model.TransportContainer, error) {
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return model.TransportContainer{}, err
	}
	target := strings.TrimSpace(input.Status)
	if !constants.CanTransition(constants.TransportContainerTransitions, current.Status, target) {
		return model.TransportContainer{}, fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, current.Status, target)
	}
	before := current.Status
	current.Status = target
	current.Version = input.ExpectedVersion + 1
	current.UpdatedAt = time.Now().UTC()
	if err := s.repository.Update(ctx, id, input.ExpectedVersion, &current); err != nil {
		return model.TransportContainer{}, fmt.Errorf("transition 运输容器: %w", err)
	}
	if err := s.security.Audit(ctx, actor, requestID, "transition", "TransportContainer", id, before, target, input.Reason); err != nil {
		return model.TransportContainer{}, fmt.Errorf("persist transition audit: %w", err)
	}
	return s.repository.Get(ctx, id)
}

func (s *transportContainerService) Delete(ctx context.Context, id uint, actor, requestID string) error {
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repository.Delete(ctx, id); err != nil {
		return err
	}
	return s.security.Audit(ctx, actor, requestID, "delete", "TransportContainer", id, current.Status, "deleted", "soft deleted 运输容器")
}

func (s *transportContainerService) StatusCounts(ctx context.Context) (map[string]int64, error) {
	return s.repository.CountByStatus(ctx)
}

func validateTransportContainerBusinessFields(code, name, facility, owner string) error {
	if strings.TrimSpace(code) == "" || strings.TrimSpace(name) == "" || strings.TrimSpace(facility) == "" || strings.TrimSpace(owner) == "" {
		return ErrInvalidInput
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func fallbackTime(primary, fallback time.Time) time.Time {
	if primary.IsZero() {
		return fallback.UTC()
	}
	return primary.UTC()
}
