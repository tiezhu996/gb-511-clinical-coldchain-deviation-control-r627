package service

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/config"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/dto"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/model"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var cumulativeDBCounter atomic.Int64

func newCumulativeTestStore(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:mem-%d?mode=memory&cache=shared", cumulativeDBCounter.Add(1))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Role{}, &model.User{}, &model.AuditLog{}, &model.SensorEvidence{},
		&model.TransportContainer{}, &model.TemperatureWindow{},
		&model.ExcursionEvent{}, &model.DispositionDecision{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newCumulativeFixture(t *testing.T) (ExcursionEventService, *gorm.DB, context.Context) {
	db := newCumulativeTestStore(t)
	containers := []model.TransportContainer{
		{BaseModel: model.BaseModel{Code: "TC-A", Name: "在途箱甲", Status: "in_transit", Version: 1}, SensorID: "SN-A"},
		{BaseModel: model.BaseModel{Code: "TC-B", Name: "在途箱乙", Status: "ready", Version: 1}, SensorID: "SN-B"},
		{BaseModel: model.BaseModel{Code: "TC-C", Name: "隔离箱丙", Status: "quarantine", Version: 1}, SensorID: "SN-C"},
	}
	if err := db.Create(&containers).Error; err != nil {
		t.Fatalf("seed containers: %v", err)
	}
	windows := []model.TemperatureWindow{
		{BaseModel: model.BaseModel{Code: "TW-15", Name: "15 分钟规则", Status: "active", Version: 1}, MaxExcursionMinutes: 15, MinimumCelsius: 2, MaximumCelsius: 8},
		{BaseModel: model.BaseModel{Code: "TW-5", Name: "5 分钟规则", Status: "active", Version: 1}, MaxExcursionMinutes: 5, MinimumCelsius: -80, MaximumCelsius: -60},
	}
	if err := db.Create(&windows).Error; err != nil {
		t.Fatalf("seed windows: %v", err)
	}
	security := NewSecurityService(repository.NewSecurityRepository(db), config.Config{})
	svc := NewExcursionEventService(
		repository.NewExcursionEventRepository(db),
		repository.NewTransportContainerRepository(db),
		repository.NewTemperatureWindowRepository(db),
		repository.NewDispositionDecisionRepository(db),
		repository.NewSensorEvidenceRepository(db),
		security,
	)
	return svc, db, context.Background()
}

func createInput(code, container, window string, duration int, at time.Time, risk string) dto.CreateExcursionEvent {
	return dto.CreateExcursionEvent{
		Code: code, Name: code + " 偏差", Facility: "沪杭运输线", Owner: "测试组", Category: "高温偏差",
		RiskLevel: risk, EffectiveAt: at, ContainerCode: container, WindowCode: window,
		ObservedTempC: 9.5, DurationMinutes: duration, DetectedAt: at, SensorEvidence: "minio://sensor/test.csv",
	}
}

func createExistingExcursion(t *testing.T, db *gorm.DB, code, container, window, status string, duration int, at time.Time, cumulative int) {
	t.Helper()
	item := model.ExcursionEvent{
		BaseModel:     model.BaseModel{Code: code, Name: code + " 历史偏差", Status: status, Version: 1},
		ContainerCode: container, WindowCode: window, DurationMinutes: duration,
		CumulativeMinutes: cumulative, DetectedAt: at, EffectiveAt: at,
		SensorEvidence: "minio://sensor/test.csv",
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed excursion %s: %v", code, err)
	}
}

func createExcursion(t *testing.T, svc ExcursionEventService, input dto.CreateExcursionEvent) model.ExcursionEvent {
	t.Helper()
	created, err := svc.Create(context.Background(), input, "operator", "req-"+input.Code)
	if err != nil {
		t.Fatalf("create %s: %v", input.Code, err)
	}
	return created
}

// Several individually-tolerable short trips must add up: once the rolling 24h
// total crosses the rule allowance the new event becomes critical and the
// container is moved into quarantine.
func TestCreateEscalatesAndQuarantinesOnCumulativeBreach(t *testing.T) {
	svc, db, ctx := newCumulativeFixture(t)
	now := time.Now().UTC()
	createExistingExcursion(t, db, "EE-1", "TC-A", "TW-15", "open", 8, now.Add(-6*time.Hour), 8)
	createExistingExcursion(t, db, "EE-2", "TC-A", "TW-15", "in_review", 5, now.Add(-2*time.Hour), 13)

	created, err := svc.Create(ctx, createInput("EE-3", "TC-A", "TW-15", 4, now, "high"), "operator", "req-3")
	if err != nil {
		t.Fatalf("create cumulative breach excursion: %v", err)
	}
	if created.RiskLevel != "critical" {
		t.Fatalf("expected critical risk after cumulative breach, got %q", created.RiskLevel)
	}
	if created.CumulativeMinutes != 17 {
		t.Fatalf("expected cumulative 17 minutes, got %d", created.CumulativeMinutes)
	}
	var container model.TransportContainer
	if err := db.Where("code = ?", "TC-A").First(&container).Error; err != nil {
		t.Fatalf("reload container: %v", err)
	}
	if container.Status != "quarantine" {
		t.Fatalf("expected container quarantined automatically, got %q", container.Status)
	}
	if container.Version != 2 {
		t.Fatalf("expected container version bumped to 2, got %d", container.Version)
	}
	var audits []model.AuditLog
	if err := db.Where("entity_type = ? AND entity_id = ?", "TransportContainer", container.ID).Find(&audits).Error; err != nil {
		t.Fatalf("load container audits: %v", err)
	}
	found := false
	for _, audit := range audits {
		if audit.Action == "transition" && audit.BeforeState == "in_transit" && audit.AfterState == "quarantine" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected quarantine transition audit, got %+v", audits)
	}
}

// Each event below the allowance stays low risk and the container keeps moving.
func TestCreateLeavesRiskAloneWhenWithinAllowance(t *testing.T) {
	svc, db, ctx := newCumulativeFixture(t)
	now := time.Now().UTC()
	createExistingExcursion(t, db, "EE-1", "TC-B", "TW-15", "open", 5, now.Add(-3*time.Hour), 5)
	created, err := svc.Create(ctx, createInput("EE-2", "TC-B", "TW-15", 10, now, "medium"), "operator", "req-2")
	if err != nil {
		t.Fatalf("create within allowance: %v", err)
	}
	if created.CumulativeMinutes != 15 {
		t.Fatalf("expected cumulative 15, got %d", created.CumulativeMinutes)
	}
	if created.RiskLevel != "medium" {
		t.Fatalf("risk must stay medium when cumulative equals allowance, got %q", created.RiskLevel)
	}
	var container model.TransportContainer
	if err := db.Where("code = ?", "TC-B").First(&container).Error; err != nil {
		t.Fatalf("reload container: %v", err)
	}
	if container.Status != "ready" {
		t.Fatalf("container must not be quarantined, got %q", container.Status)
	}
}

// Closed-loop deviations never enter the rolling total, even within 24h.
func TestCreateExcludesClosedEvents(t *testing.T) {
	svc, db, _ := newCumulativeFixture(t)
	now := time.Now().UTC()
	createExistingExcursion(t, db, "EE-1", "TC-B", "TW-15", "closed", 14, now.Add(-time.Hour), 14)
	created := createExcursion(t, svc, createInput("EE-2", "TC-B", "TW-15", 2, now, "low"))
	if created.CumulativeMinutes != 2 {
		t.Fatalf("closed excursion must not count, expected cumulative 2, got %d", created.CumulativeMinutes)
	}
}

// A release disposition only finalizes the disposition; while the excursion is
// not closed it is still an open deviation and stays in the cumulative total.
func TestCreateIncludesDecidedReleasedEvents(t *testing.T) {
	svc, db, ctx := newCumulativeFixture(t)
	now := time.Now().UTC()
	createExistingExcursion(t, db, "EE-1", "TC-B", "TW-15", "decided", 12, now.Add(-30*time.Minute), 12)
	created, err := svc.Create(ctx, createInput("EE-2", "TC-B", "TW-15", 4, now, "high"), "operator", "req-2")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.CumulativeMinutes != 16 {
		t.Fatalf("decided (released, unclosed) excursion must count, got %d", created.CumulativeMinutes)
	}
	if created.RiskLevel != "critical" {
		t.Fatalf("expected critical after breach, got %q", created.RiskLevel)
	}
}

// Events older than 24h relative to the new detection time drop out of the window.
func TestCreateUsesRolling24hWindow(t *testing.T) {
	svc, db, _ := newCumulativeFixture(t)
	now := time.Now().UTC()
	createExistingExcursion(t, db, "EE-OLD", "TC-B", "TW-15", "open", 14, now.Add(-25*time.Hour), 14)
	createExistingExcursion(t, db, "EE-RECENT", "TC-B", "TW-15", "closed", 14, now.Add(-2*time.Hour), 14)
	created := createExcursion(t, svc, createInput("EE-NEW", "TC-B", "TW-15", 10, now, "low"))
	if created.CumulativeMinutes != 10 {
		t.Fatalf("only the new event should count, got %d", created.CumulativeMinutes)
	}
}

// The cumulative rollup is scoped per container and per referenced rule.
func TestCreateScopesByContainerAndWindow(t *testing.T) {
	svc, db, _ := newCumulativeFixture(t)
	now := time.Now().UTC()
	createExistingExcursion(t, db, "EE-OTHER-CONTAINER", "TC-A", "TW-15", "open", 14, now.Add(-time.Hour), 14)
	createExistingExcursion(t, db, "EE-OTHER-WINDOW", "TC-B", "TW-5", "open", 4, now.Add(-time.Hour), 4)
	created := createExcursion(t, svc, createInput("EE-NEW", "TC-B", "TW-15", 10, now, "low"))
	if created.CumulativeMinutes != 10 {
		t.Fatalf("other container/window events must not count, got %d", created.CumulativeMinutes)
	}
}

// An already quarantined container is not migrated again; the event still
// records the breach and the critical risk.
func TestCreateDoesNotRequarantineQuarantinedContainer(t *testing.T) {
	svc, db, ctx := newCumulativeFixture(t)
	now := time.Now().UTC()
	createExistingExcursion(t, db, "EE-1", "TC-C", "TW-5", "open", 3, now.Add(-time.Hour), 3)
	created, err := svc.Create(ctx, createInput("EE-2", "TC-C", "TW-5", 3, now, "high"), "operator", "req-2")
	if err != nil {
		t.Fatalf("create against quarantined container: %v", err)
	}
	if created.CumulativeMinutes != 6 || created.RiskLevel != "critical" {
		t.Fatalf("expected critical cumulative 6, got %d/%s", created.CumulativeMinutes, created.RiskLevel)
	}
	var container model.TransportContainer
	if err := db.Where("code = ?", "TC-C").First(&container).Error; err != nil {
		t.Fatalf("reload container: %v", err)
	}
	if container.Status != "quarantine" || container.Version != 1 {
		t.Fatalf("quarantined container must not be migrated, got %s v%d", container.Status, container.Version)
	}
	var containerAudits int64
	if err := db.Model(&model.AuditLog{}).Where("entity_type = ?", "TransportContainer").Count(&containerAudits).Error; err != nil {
		t.Fatalf("count audits: %v", err)
	}
	if containerAudits != 0 {
		t.Fatalf("expected no container quarantine audit, found %d", containerAudits)
	}
}

// Missing referenced entities are rejected as business errors instead of
// silently persisting an unbound deviation.
func TestCreateRejectsUnknownReferences(t *testing.T) {
	svc, _, ctx := newCumulativeFixture(t)
	now := time.Now().UTC()
	if _, err := svc.Create(ctx, createInput("EE-X", "TC-UNKNOWN", "TW-15", 5, now, "low"), "operator", "req-x"); err == nil || !strings.Contains(err.Error(), "container") {
		t.Fatalf("expected invalid container error, got %v", err)
	}
	if _, err := svc.Create(ctx, createInput("EE-X", "TC-A", "TW-UNKNOWN", 5, now, "low"), "operator", "req-x"); err == nil || !strings.Contains(err.Error(), "temperature window") {
		t.Fatalf("expected invalid window error, got %v", err)
	}
}

// Reviewer edits must not erase the registration-time cumulative snapshot.
func TestUpdatePreservesCumulativeSnapshot(t *testing.T) {
	svc, _, ctx := newCumulativeFixture(t)
	now := time.Now().UTC()
	created := createExcursion(t, svc, createInput("EE-KEEP", "TC-B", "TW-15", 10, now, "low"))
	update := dto.UpdateExcursionEvent{
		ExpectedVersion: created.Version, Name: "EE-KEEP 偏差", Facility: "沪杭运输线", Owner: "测试组",
		Category: "高温偏差", RiskLevel: "medium", EffectiveAt: now, ContainerCode: "TC-B", WindowCode: "TW-15",
		ObservedTempC: 9.2, DurationMinutes: 10, DetectedAt: now, SensorEvidence: "minio://sensor/test.csv",
	}
	updated, err := svc.Update(ctx, created.ID, update, "reviewer", "req-upd")
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.CumulativeMinutes != created.CumulativeMinutes {
		t.Fatalf("cumulative snapshot must survive edits, got %d want %d", updated.CumulativeMinutes, created.CumulativeMinutes)
	}
}

// The read model behind the deviation page shows the live rolling total and the
// rule comparison.
func TestCumulativeReadModel(t *testing.T) {
	svc, db, ctx := newCumulativeFixture(t)
	now := time.Now().UTC()
	createExistingExcursion(t, db, "EE-1", "TC-A", "TW-15", "open", 8, now.Add(-2*time.Hour), 8)
	createExistingExcursion(t, db, "EE-2", "TC-A", "TW-15", "decided", 8, now.Add(-30*time.Minute), 16)
	createExistingExcursion(t, db, "EE-3", "TC-A", "TW-15", "closed", 8, now.Add(-10*time.Minute), 24)

	view, err := svc.Cumulative(ctx, dto.ExcursionCumulativeQuery{ContainerCode: "tc-a", WindowCode: "tw-15"})
	if err != nil {
		t.Fatalf("cumulative read: %v", err)
	}
	if view.CumulativeMinutes != 16 {
		t.Fatalf("expected 16 open minutes, got %d", view.CumulativeMinutes)
	}
	if view.OpenEventCount != 2 {
		t.Fatalf("expected 2 open events, got %d", view.OpenEventCount)
	}
	if view.MaxAllowedMinutes != 15 || !view.Exceeded {
		t.Fatalf("expected exceeded against 15 minute rule, got max=%d exceeded=%v", view.MaxAllowedMinutes, view.Exceeded)
	}
	if view.WindowHours != 24 || view.ContainerCode != "TC-A" {
		t.Fatalf("unexpected view envelope: %+v", view)
	}
}
