package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/constants"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/dto"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/model"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/repository"
	"gorm.io/gorm"
)

// cumulativeWindow defines how far back open excursions are rolled up when a new
// deviation is registered and when the workbench reads the current total.
const cumulativeWindow = 24 * time.Hour

type ExcursionEventService interface {
	List(context.Context, dto.PageQuery) (repository.Page[model.ExcursionEvent], error)
	Get(context.Context, uint) (model.ExcursionEvent, error)
	Create(context.Context, dto.CreateExcursionEvent, string, string) (model.ExcursionEvent, error)
	Update(context.Context, uint, dto.UpdateExcursionEvent, string, string) (model.ExcursionEvent, error)
	Transition(context.Context, uint, dto.TransitionRequest, string, string) (model.ExcursionEvent, error)
	Delete(context.Context, uint, string, string) error
	StatusCounts(context.Context) (map[string]int64, error)
	Cumulative(context.Context, dto.ExcursionCumulativeQuery) (dto.ExcursionCumulativeView, error)
}

type excursionEventService struct {
	repository  repository.ExcursionEventRepository
	containers  repository.TransportContainerRepository
	windows     repository.TemperatureWindowRepository
	disposition repository.DispositionDecisionRepository
	evidence    repository.SensorEvidenceRepository
	security    SecurityService
}

func NewExcursionEventService(repo repository.ExcursionEventRepository, containers repository.TransportContainerRepository, windows repository.TemperatureWindowRepository, disposition repository.DispositionDecisionRepository, evidence repository.SensorEvidenceRepository, security SecurityService) ExcursionEventService {
	return &excursionEventService{repository: repo, containers: containers, windows: windows, disposition: disposition, evidence: evidence, security: security}
}

func (s *excursionEventService) List(ctx context.Context, query dto.PageQuery) (repository.Page[model.ExcursionEvent], error) {
	return s.repository.List(ctx, query)
}

func (s *excursionEventService) Get(ctx context.Context, id uint) (model.ExcursionEvent, error) {
	return s.repository.Get(ctx, id)
}

func (s *excursionEventService) Create(ctx context.Context, input dto.CreateExcursionEvent, actor, requestID string) (model.ExcursionEvent, error) {
	if err := validateExcursionEventBusinessFields(input.Code, input.Name, input.Facility, input.Owner); err != nil {
		return model.ExcursionEvent{}, err
	}
	item := model.ExcursionEvent{
		BaseModel: model.BaseModel{
			Code: strings.ToUpper(strings.TrimSpace(input.Code)), Name: strings.TrimSpace(input.Name),
			Status: model.ExcursionEventInitialStatus, Version: 1, Description: strings.TrimSpace(input.Description),
		},
		Facility: strings.TrimSpace(input.Facility), Owner: strings.TrimSpace(input.Owner),
		Category: strings.TrimSpace(input.Category), RiskLevel: input.RiskLevel,
		MetricValue: input.MetricValue, MetricUnit: strings.TrimSpace(input.MetricUnit),
		EffectiveAt: input.EffectiveAt.UTC(), Evidence: strings.TrimSpace(input.Evidence),
		RelatedCode:   strings.ToUpper(strings.TrimSpace(input.RelatedCode)),
		ContainerCode: strings.ToUpper(strings.TrimSpace(firstNonEmpty(input.ContainerCode, input.RelatedCode))),
		WindowCode:    strings.ToUpper(strings.TrimSpace(input.WindowCode)),
		ObservedTempC: input.ObservedTempC, DurationMinutes: input.DurationMinutes,
		DetectedAt:     fallbackTime(input.DetectedAt, input.EffectiveAt),
		SensorEvidence: strings.TrimSpace(firstNonEmpty(input.SensorEvidence, input.Evidence)),
		Reviewer:       strings.TrimSpace(input.Reviewer),
	}
	if item.ObservedTempC == 0 {
		item.ObservedTempC = input.MetricValue
	}
	if item.ContainerCode == "" || item.WindowCode == "" || item.SensorEvidence == "" || item.DurationMinutes < 1 {
		return model.ExcursionEvent{}, fmt.Errorf("%w: container, temperature window, duration and sensor evidence are required", ErrInvalidInput)
	}
	container, err := s.containers.GetByCode(ctx, item.ContainerCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.ExcursionEvent{}, fmt.Errorf("%w: referenced container %s does not exist", ErrInvalidInput, item.ContainerCode)
		}
		return model.ExcursionEvent{}, fmt.Errorf("load 运输容器: %w", err)
	}
	window, err := s.windows.GetByCode(ctx, item.WindowCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.ExcursionEvent{}, fmt.Errorf("%w: referenced temperature window %s does not exist", ErrInvalidInput, item.WindowCode)
		}
		return model.ExcursionEvent{}, fmt.Errorf("load 温控规则: %w", err)
	}
	// Roll up still-open deviations for the same container and referenced rule in
	// the trailing 24h, then fold this event in. Closed-loop events are excluded;
	// released-but-unclosed events stay in the total.
	since := item.DetectedAt.Add(-cumulativeWindow)
	previousMinutes, err := s.repository.SumOpenMinutes(ctx, item.ContainerCode, item.WindowCode, since, item.DetectedAt)
	if err != nil {
		return model.ExcursionEvent{}, fmt.Errorf("aggregate cumulative excursion minutes: %w", err)
	}
	item.CumulativeMinutes = previousMinutes + item.DurationMinutes
	exceeded := item.CumulativeMinutes > window.MaxExcursionMinutes
	if exceeded {
		item.RiskLevel = "critical"
	}
	createDetail, _ := json.Marshal(map[string]any{
		"containerCode": item.ContainerCode, "windowCode": item.WindowCode,
		"durationMinutes": item.DurationMinutes, "cumulativeMinutes": item.CumulativeMinutes,
		"maxAllowedMinutes": window.MaxExcursionMinutes, "limitExceeded": exceeded,
	})
	if err := s.repository.Create(ctx, &item); err != nil {
		return model.ExcursionEvent{}, fmt.Errorf("create 偏差事件: %w", err)
	}
	_ = s.security.Audit(ctx, actor, requestID, "create", "ExcursionEvent", item.ID, "", item.Status, string(createDetail))
	if exceeded && container.Status != string(constants.ContainerStateQuarantine) {
		quarantineDetail, _ := json.Marshal(map[string]any{
			"reason":        "rolling 24h cumulative excursion exceeded the temperature window allowance",
			"excursionCode": item.Code, "containerCode": item.ContainerCode, "windowCode": item.WindowCode,
			"cumulativeMinutes": item.CumulativeMinutes, "maxAllowedMinutes": window.MaxExcursionMinutes,
		})
		moved, _, moveErr := s.containers.MoveToQuarantine(ctx, container.ID, actor, requestID, string(quarantineDetail))
		if moveErr != nil {
			return model.ExcursionEvent{}, fmt.Errorf("quarantine 运输容器 after cumulative excursion: %w", moveErr)
		}
		if moved {
			_ = s.security.Audit(ctx, actor, requestID, "quarantine", "ExcursionEvent", item.ID, container.Status, "quarantine", string(quarantineDetail))
		}
	}
	return item, nil
}

func (s *excursionEventService) Update(ctx context.Context, id uint, input dto.UpdateExcursionEvent, actor, requestID string) (model.ExcursionEvent, error) {
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return model.ExcursionEvent{}, err
	}
	if err := validateExcursionEventBusinessFields(current.Code, input.Name, input.Facility, input.Owner); err != nil {
		return model.ExcursionEvent{}, err
	}
	// CumulativeMinutes is a registration-time snapshot that edits must never clear.
	cumulativeSnapshot := current.CumulativeMinutes
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
	current.ContainerCode = strings.ToUpper(strings.TrimSpace(firstNonEmpty(input.ContainerCode, input.RelatedCode)))
	current.WindowCode = strings.ToUpper(strings.TrimSpace(input.WindowCode))
	current.ObservedTempC = input.ObservedTempC
	if current.ObservedTempC == 0 {
		current.ObservedTempC = input.MetricValue
	}
	current.DurationMinutes = input.DurationMinutes
	current.DetectedAt = fallbackTime(input.DetectedAt, input.EffectiveAt)
	current.SensorEvidence = strings.TrimSpace(firstNonEmpty(input.SensorEvidence, input.Evidence))
	current.Reviewer = strings.TrimSpace(input.Reviewer)
	current.CumulativeMinutes = cumulativeSnapshot
	current.Version = input.ExpectedVersion + 1
	current.UpdatedAt = time.Now().UTC()
	if err := s.repository.Update(ctx, id, input.ExpectedVersion, &current); err != nil {
		return model.ExcursionEvent{}, fmt.Errorf("update 偏差事件: %w", err)
	}
	_ = s.security.Audit(ctx, actor, requestID, "update", "ExcursionEvent", id, current.Status, current.Status, "updated business fields")
	return s.repository.Get(ctx, id)
}

func (s *excursionEventService) Transition(ctx context.Context, id uint, input dto.TransitionRequest, actor, requestID string) (model.ExcursionEvent, error) {
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return model.ExcursionEvent{}, err
	}
	target := strings.TrimSpace(input.Status)
	if !constants.CanTransition(constants.ExcursionEventTransitions, current.Status, target) {
		return model.ExcursionEvent{}, fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, current.Status, target)
	}
	before := current.Status
	evidence := strings.TrimSpace(input.Evidence)
	if evidence == "" {
		evidence = strings.TrimSpace(firstNonEmpty(current.SensorEvidence, current.Evidence))
	}
	if target == string(constants.ExcursionStateDecided) && evidence == "" {
		return model.ExcursionEvent{}, fmt.Errorf("%w: sensor evidence is required before deciding an excursion", ErrInvalidInput)
	}
	if target == string(constants.ExcursionStateDecided) {
		if count, err := s.evidence.CountForExcursion(ctx, current.Code); err != nil || count == 0 {
			return model.ExcursionEvent{}, fmt.Errorf("%w: registered sensor evidence is required before deciding an excursion", ErrInvalidInput)
		}
	}
	if target == string(constants.ExcursionStateClosed) {
		final, err := s.disposition.HasFinalForExcursion(ctx, current.Code)
		if err != nil || !final {
			return model.ExcursionEvent{}, fmt.Errorf("%w: a final disposition is required before closing an excursion", ErrInvalidInput)
		}
	}
	current.Status = target
	current.SensorEvidence = evidence
	current.Evidence = evidence
	if target == string(constants.ExcursionStateInReview) || target == string(constants.ExcursionStateDecided) {
		current.Reviewer = actor
	}
	current.Version = input.ExpectedVersion + 1
	current.UpdatedAt = time.Now().UTC()
	detail, _ := json.Marshal(map[string]any{"reason": input.Reason, "sensorEvidence": evidence, "containerCode": current.ContainerCode})
	if err := s.repository.Update(ctx, id, input.ExpectedVersion, &current, auditLog(actor, requestID, "transition", "ExcursionEvent", id, before, target, string(detail))); err != nil {
		return model.ExcursionEvent{}, fmt.Errorf("transition 偏差事件: %w", err)
	}
	return s.repository.Get(ctx, id)
}

func (s *excursionEventService) Delete(ctx context.Context, id uint, actor, requestID string) error {
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repository.Delete(ctx, id); err != nil {
		return err
	}
	return s.security.Audit(ctx, actor, requestID, "delete", "ExcursionEvent", id, current.Status, "deleted", "soft deleted 偏差事件")
}

func (s *excursionEventService) StatusCounts(ctx context.Context) (map[string]int64, error) {
	return s.repository.CountByStatus(ctx)
}

// Cumulative reports the container's current rolling 24h open excursion total so
// the deviation workbench can show how much tolerance is already consumed.
func (s *excursionEventService) Cumulative(ctx context.Context, query dto.ExcursionCumulativeQuery) (dto.ExcursionCumulativeView, error) {
	containerCode, windowCode := query.Normalized()
	if containerCode == "" {
		return dto.ExcursionCumulativeView{}, fmt.Errorf("%w: container code is required", ErrInvalidInput)
	}
	if _, err := s.containers.GetByCode(ctx, containerCode); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.ExcursionCumulativeView{}, fmt.Errorf("%w: container %s does not exist", ErrInvalidInput, containerCode)
		}
		return dto.ExcursionCumulativeView{}, err
	}
	now := time.Now().UTC()
	since := now.Add(-cumulativeWindow)
	total, err := s.repository.SumOpenMinutes(ctx, containerCode, windowCode, since, now)
	if err != nil {
		return dto.ExcursionCumulativeView{}, fmt.Errorf("aggregate cumulative excursion minutes: %w", err)
	}
	openCount, err := s.repository.CountOpen(ctx, containerCode, windowCode, since, now)
	if err != nil {
		return dto.ExcursionCumulativeView{}, err
	}
	view := dto.ExcursionCumulativeView{
		ContainerCode: containerCode, WindowCode: windowCode, WindowHours: int(cumulativeWindow.Hours()),
		CumulativeMinutes: total, OpenEventCount: int(openCount),
	}
	if windowCode != "" {
		window, err := s.windows.GetByCode(ctx, windowCode)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return dto.ExcursionCumulativeView{}, fmt.Errorf("%w: temperature window %s does not exist", ErrInvalidInput, windowCode)
			}
			return dto.ExcursionCumulativeView{}, err
		}
		view.MaxAllowedMinutes = window.MaxExcursionMinutes
		view.Exceeded = total > window.MaxExcursionMinutes
	}
	return view, nil
}

func validateExcursionEventBusinessFields(code, name, facility, owner string) error {
	if strings.TrimSpace(code) == "" || strings.TrimSpace(name) == "" || strings.TrimSpace(facility) == "" || strings.TrimSpace(owner) == "" {
		return ErrInvalidInput
	}
	return nil
}
