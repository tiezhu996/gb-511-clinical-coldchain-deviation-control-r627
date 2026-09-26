package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/dto"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/model"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/repository"
	"github.com/minio/minio-go/v7"
)

type SensorEvidenceService interface {
	List(context.Context, dto.PageQuery) (repository.Page[model.SensorEvidence], error)
	Get(context.Context, uint) (model.SensorEvidence, error)
	Create(context.Context, dto.CreateSensorEvidence, string, string) (model.SensorEvidence, error)
}

type sensorEvidenceService struct {
	repository repository.SensorEvidenceRepository
	minio      *minio.Client
	bucket     string
}

func NewSensorEvidenceService(repo repository.SensorEvidenceRepository, client *minio.Client, bucket string) SensorEvidenceService {
	return &sensorEvidenceService{repository: repo, minio: client, bucket: bucket}
}

func (s *sensorEvidenceService) List(ctx context.Context, query dto.PageQuery) (repository.Page[model.SensorEvidence], error) {
	return s.repository.List(ctx, query)
}
func (s *sensorEvidenceService) Get(ctx context.Context, id uint) (model.SensorEvidence, error) {
	return s.repository.Get(ctx, id)
}

func (s *sensorEvidenceService) Create(ctx context.Context, input dto.CreateSensorEvidence, actor, requestID string) (model.SensorEvidence, error) {
	if s.minio == nil || strings.TrimSpace(s.bucket) == "" {
		return model.SensorEvidence{}, fmt.Errorf("object storage is unavailable")
	}
	excursion, container := strings.ToUpper(strings.TrimSpace(input.ExcursionCode)), strings.ToUpper(strings.TrimSpace(input.ContainerCode))
	if strings.HasPrefix(input.ObjectKey, "/") || strings.Contains(input.ObjectKey, "..") {
		return model.SensorEvidence{}, fmt.Errorf("%w: unsafe evidence object key", ErrInvalidInput)
	}
	valid, err := s.repository.ReferencesExist(ctx, excursion, container)
	if err != nil || !valid {
		return model.SensorEvidence{}, fmt.Errorf("%w: evidence must reference an existing excursion and its container", ErrInvalidInput)
	}
	exists, err := s.minio.BucketExists(ctx, s.bucket)
	if err != nil {
		return model.SensorEvidence{}, fmt.Errorf("check evidence bucket: %w", err)
	}
	if !exists {
		if err := s.minio.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return model.SensorEvidence{}, fmt.Errorf("create evidence bucket: %w", err)
		}
	}
	item := model.SensorEvidence{Code: strings.ToUpper(strings.TrimSpace(input.Code)), ExcursionCode: excursion, ContainerCode: container, ObjectKey: strings.TrimSpace(input.ObjectKey), SHA256: strings.ToLower(input.SHA256), MediaType: strings.TrimSpace(input.MediaType), SizeBytes: input.SizeBytes, CapturedAt: input.CapturedAt.UTC(), CapturedBy: actor, Source: strings.TrimSpace(input.Source), CreatedAt: time.Now().UTC()}
	upload, err := s.minio.PresignedPutObject(ctx, s.bucket, item.ObjectKey, 15*time.Minute)
	if err != nil {
		return model.SensorEvidence{}, fmt.Errorf("sign evidence upload: %w", err)
	}
	audit := &model.AuditLog{Actor: actor, RequestID: requestID, Action: "evidence_registered", EntityType: "SensorEvidence", BeforeState: "", AfterState: "registered", Detail: fmt.Sprintf("object=%s sha256=%s excursion=%s", item.ObjectKey, item.SHA256, item.ExcursionCode), CreatedAt: time.Now().UTC()}
	if err := s.repository.Create(ctx, &item, audit); err != nil {
		return model.SensorEvidence{}, fmt.Errorf("register sensor evidence: %w", err)
	}
	item.UploadURL = upload.String()
	return item, nil
}
