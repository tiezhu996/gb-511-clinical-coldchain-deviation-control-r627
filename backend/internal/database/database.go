package database

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/config"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/model"
	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(ctx context.Context, cfg config.Config, log *slog.Logger) (*gorm.DB, *redis.Client, error) {
	var dialector gorm.Dialector
	switch cfg.DatabaseDriver {
	case "postgres":
		dialector = postgres.Open(cfg.DatabaseDSN)
	case "mysql":
		dialector = mysql.Open(cfg.DatabaseDSN)
	case "sqlite":
		dialector = sqlite.Open(cfg.DatabaseDSN)
	default:
		return nil, nil, fmt.Errorf("unsupported database driver %q", cfg.DatabaseDriver)
	}
	logLevel := logger.Warn
	if cfg.Environment == "development" {
		logLevel = logger.Info
	}
	var db *gorm.DB
	var err error
	for attempt := 1; attempt <= 20; attempt++ {
		db, err = gorm.Open(dialector, &gorm.Config{Logger: logger.Default.LogMode(logLevel)})
		if err == nil {
			sqlDB, dbErr := db.DB()
			if dbErr == nil && sqlDB.PingContext(ctx) == nil {
				break
			}
			if dbErr != nil {
				err = dbErr
			} else {
				err = sqlDB.PingContext(ctx)
			}
		}
		log.Warn("database not ready", "attempt", attempt, "error", err)
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("connect database: %w", err)
	}
	if err := migrate(db); err != nil {
		return nil, nil, err
	}
	if err := Seed(ctx, db); err != nil {
		return nil, nil, err
	}
	var redisClient *redis.Client
	if cfg.RedisAddr != "" {
		redisClient = redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword})
		if err := redisClient.Ping(ctx).Err(); err != nil {
			return nil, nil, fmt.Errorf("connect redis: %w", err)
		}
	}
	return db, redisClient, nil
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Role{}, &model.User{}, &model.AuditLog{}, &model.SensorEvidence{},
		&model.TransportContainer{},
		&model.TemperatureWindow{},
		&model.ExcursionEvent{},
		&model.DispositionDecision{},
	)
}

func Seed(ctx context.Context, db *gorm.DB) error {
	roles := []model.Role{
		{Code: model.RoleViewer, Name: "只读观察员", Permissions: "read", Active: true},
		{Code: model.RoleOperator, Name: "冷链操作员", Permissions: "read,operate,propose", Active: true},
		{Code: model.RoleReviewer, Name: "质量复核员", Permissions: "read,operate,review,audit", Active: true},
		{Code: model.RoleAdmin, Name: "系统管理员", Permissions: "*", Active: true},
	}
	for i := range roles {
		if err := db.WithContext(ctx).Where("code = ?", roles[i].Code).FirstOrCreate(&roles[i]).Error; err != nil {
			return err
		}
	}
	var users int64
	if err := db.WithContext(ctx).Model(&model.User{}).Count(&users).Error; err != nil {
		return err
	}
	if users == 0 {
		password, err := bcrypt.GenerateFromPassword([]byte("Admin123!"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		seedUsers := []model.User{
			{Username: "admin", DisplayName: "系统管理员", PasswordHash: string(password), Role: model.RoleAdmin, Active: true},
			{Username: "reviewer", DisplayName: "质量复核员", PasswordHash: string(password), Role: model.RoleReviewer, Active: true},
			{Username: "operator", DisplayName: "现场操作员", PasswordHash: string(password), Role: model.RoleOperator, Active: true},
			{Username: "viewer", DisplayName: "只读观察员", PasswordHash: string(password), Role: model.RoleViewer, Active: true},
		}
		if err := db.WithContext(ctx).Create(&seedUsers).Error; err != nil {
			return err
		}
	}

	if err := seedTransportContainer(ctx, db); err != nil {
		return err
	}

	if err := seedTemperatureWindow(ctx, db); err != nil {
		return err
	}

	if err := seedExcursionEvent(ctx, db); err != nil {
		return err
	}
	if err := seedSensorEvidence(ctx, db); err != nil {
		return err
	}

	if err := seedDispositionDecision(ctx, db); err != nil {
		return err
	}

	return nil
}

func seedTransportContainer(ctx context.Context, db *gorm.DB) error {
	var count int64
	if err := db.WithContext(ctx).Model(&model.TransportContainer{}).Count(&count).Error; err != nil || count > 0 {
		return err
	}
	now := time.Now().UTC()
	items := []model.TransportContainer{
		{BaseModel: model.BaseModel{Code: "TC-001", Name: "细胞治疗样本箱", Status: "ready", Version: 1,
			Description: "2-8C 临床样本运输容器"}, SensorID: "SN-TEMP-4101", ContainerType: "主动制冷箱", CurrentLocation: "上海配送中心", Custodian: "李运输", CurrentTempC: 4.2, LastSensorReading: now.Add(-8 * time.Minute),
			Facility: "上海配送中心", Owner: "李运输", Category: "主动制冷箱", RiskLevel: "low", MetricValue: 4.2, MetricUnit: "C", EffectiveAt: now.Add(-8 * time.Minute), Evidence: "minio://sensor/tc-001/latest.json", RelatedCode: "SN-TEMP-4101"},
		{BaseModel: model.BaseModel{Code: "TC-002", Name: "疫苗临床用箱", Status: "in_transit", Version: 1,
			Description: "机场至研究中心在途容器"}, SensorID: "SN-TEMP-4102", ContainerType: "被动保温箱", CurrentLocation: "沪杭高速 G60", Custodian: "王押运", CurrentTempC: 7.6, LastSensorReading: now.Add(-3 * time.Minute),
			Facility: "沪杭运输线", Owner: "王押运", Category: "被动保温箱", RiskLevel: "medium", MetricValue: 7.6, MetricUnit: "C", EffectiveAt: now.Add(-3 * time.Minute), Evidence: "minio://sensor/tc-002/latest.json", RelatedCode: "SN-TEMP-4102"},
		{BaseModel: model.BaseModel{Code: "TC-003", Name: "冻存试剂运输罐", Status: "quarantine", Version: 1,
			Description: "检测到持续高温，已物理隔离"}, SensorID: "SN-TEMP-4103", ContainerType: "干冰运输罐", CurrentLocation: "杭州研究中心隔离区", Custodian: "赵收货", CurrentTempC: -41.8, LastSensorReading: now.Add(-2 * time.Minute),
			Facility: "杭州研究中心", Owner: "赵收货", Category: "干冰运输罐", RiskLevel: "critical", MetricValue: -41.8, MetricUnit: "C", EffectiveAt: now.Add(-2 * time.Minute), Evidence: "minio://sensor/tc-003/excursion.csv", RelatedCode: "SN-TEMP-4103"},
	}
	return db.WithContext(ctx).Create(&items).Error
}

func seedTemperatureWindow(ctx context.Context, db *gorm.DB) error {
	var count int64
	if err := db.WithContext(ctx).Model(&model.TemperatureWindow{}).Count(&count).Error; err != nil || count > 0 {
		return err
	}
	now := time.Now().UTC()
	items := []model.TemperatureWindow{
		{BaseModel: model.BaseModel{Code: "TW-001", Name: "冷藏临床样本 2-8C", Status: "active", Version: 1, Description: "单次越界最长允许 15 分钟"}, ProductClass: "临床样本", MinimumCelsius: 2, MaximumCelsius: 8, MaxExcursionMinutes: 15, QualityOwner: "reviewer", Facility: "质量体系 QMS-CC-02", Owner: "reviewer", Category: "临床样本", RiskLevel: "high", MetricValue: 8, MetricUnit: "C", EffectiveAt: now.Add(-30 * 24 * time.Hour), Evidence: "SOP-CC-02 v4", RelatedCode: "SOP-CC-02"},
		{BaseModel: model.BaseModel{Code: "TW-002", Name: "深低温试剂 -80~-60C", Status: "active", Version: 1, Description: "深低温试剂运输温控窗口"}, ProductClass: "冻存试剂", MinimumCelsius: -80, MaximumCelsius: -60, MaxExcursionMinutes: 5, QualityOwner: "reviewer", Facility: "质量体系 QMS-CC-08", Owner: "reviewer", Category: "冻存试剂", RiskLevel: "critical", MetricValue: -60, MetricUnit: "C", EffectiveAt: now.Add(-20 * 24 * time.Hour), Evidence: "SOP-CC-08 v2", RelatedCode: "SOP-CC-08"},
		{BaseModel: model.BaseModel{Code: "TW-003", Name: "常温辅材 15-25C", Status: "draft", Version: 1, Description: "待质量负责人批准启用"}, ProductClass: "临床辅材", MinimumCelsius: 15, MaximumCelsius: 25, MaxExcursionMinutes: 60, QualityOwner: "reviewer", Facility: "质量体系草案", Owner: "reviewer", Category: "临床辅材", RiskLevel: "medium", MetricValue: 25, MetricUnit: "C", EffectiveAt: now.Add(24 * time.Hour), Evidence: "DRAFT-SOP-CC-15", RelatedCode: "DRAFT-SOP-CC-15"},
	}
	return db.WithContext(ctx).Create(&items).Error
}

func seedExcursionEvent(ctx context.Context, db *gorm.DB) error {
	var count int64
	if err := db.WithContext(ctx).Model(&model.ExcursionEvent{}).Count(&count).Error; err != nil || count > 0 {
		return err
	}
	now := time.Now().UTC()
	items := []model.ExcursionEvent{
		{BaseModel: model.BaseModel{Code: "EE-001", Name: "TC-003 深低温回升", Status: "open", Version: 1, Description: "温度连续 12 分钟高于 -60C"}, ContainerCode: "TC-003", WindowCode: "TW-002", ObservedTempC: -41.8, DurationMinutes: 12, DetectedAt: now.Add(-35 * time.Minute), SensorEvidence: "minio://sensor/tc-003/excursion.csv", Facility: "杭州研究中心", Owner: "未分配", Category: "高温偏差", RiskLevel: "critical", MetricValue: -41.8, MetricUnit: "C", EffectiveAt: now.Add(-35 * time.Minute), Evidence: "minio://sensor/tc-003/excursion.csv", RelatedCode: "TC-003"},
		{BaseModel: model.BaseModel{Code: "EE-002", Name: "TC-002 短时接近上限", Status: "in_review", Version: 1, Description: "7.6C 持续 8 分钟，尚未越过 8C"}, ContainerCode: "TC-002", WindowCode: "TW-001", ObservedTempC: 7.6, DurationMinutes: 8, DetectedAt: now.Add(-2 * time.Hour), SensorEvidence: "minio://sensor/tc-002/trace.csv", Reviewer: "reviewer", Facility: "沪杭运输线", Owner: "reviewer", Category: "趋势预警", RiskLevel: "medium", MetricValue: 7.6, MetricUnit: "C", EffectiveAt: now.Add(-2 * time.Hour), Evidence: "minio://sensor/tc-002/trace.csv", RelatedCode: "TC-002"},
		{BaseModel: model.BaseModel{Code: "EE-003", Name: "TC-001 门开短时波动", Status: "decided", Version: 2, Description: "峰值 8.4C 持续 3 分钟，稳定性评估可接受"}, ContainerCode: "TC-001", WindowCode: "TW-001", ObservedTempC: 8.4, DurationMinutes: 3, DetectedAt: now.Add(-6 * time.Hour), SensorEvidence: "minio://sensor/tc-001/door-open.csv", Reviewer: "reviewer", Facility: "上海配送中心", Owner: "reviewer", Category: "短时高温", RiskLevel: "low", MetricValue: 8.4, MetricUnit: "C", EffectiveAt: now.Add(-6 * time.Hour), Evidence: "minio://sensor/tc-001/door-open.csv", RelatedCode: "TC-001"},
	}
	return db.WithContext(ctx).Create(&items).Error
}

func seedDispositionDecision(ctx context.Context, db *gorm.DB) error {
	var count int64
	if err := db.WithContext(ctx).Model(&model.DispositionDecision{}).Count(&count).Error; err != nil || count > 0 {
		return err
	}
	now := time.Now().UTC()
	items := []model.DispositionDecision{
		{BaseModel: model.BaseModel{Code: "DD-001", Name: "EE-001 处置提议", Status: "draft", Version: 1, Description: "建议隔离并进行稳定性评估"}, ExcursionCode: "EE-001", DecisionBasis: "深低温窗口超限 18.2C", SensorEvidence: "minio://sensor/tc-003/excursion.csv", ProposedBy: "operator", Facility: "质量放行组", Owner: "operator", Category: "隔离提议", RiskLevel: "critical", MetricValue: -41.8, MetricUnit: "C", EffectiveAt: now.Add(-20 * time.Minute), Evidence: "minio://sensor/tc-003/excursion.csv", RelatedCode: "EE-001"},
		{BaseModel: model.BaseModel{Code: "DD-002", Name: "EE-003 放行决定", Status: "release", Version: 2, Description: "3 分钟短时波动低于允许时长"}, ExcursionCode: "EE-003", DecisionBasis: "稳定性报告 SR-1021 支持继续使用", SensorEvidence: "minio://sensor/tc-001/door-open.csv", ProposedBy: "operator", ApprovedBy: "reviewer", DecidedAt: timePointer(now.Add(-5 * time.Hour)), Facility: "质量放行组", Owner: "operator", Category: "放行", RiskLevel: "low", MetricValue: 8.4, MetricUnit: "C", EffectiveAt: now.Add(-5 * time.Hour), Evidence: "minio://sensor/tc-001/door-open.csv", RelatedCode: "EE-003"},
		{BaseModel: model.BaseModel{Code: "DD-003", Name: "历史隔离决定", Status: "quarantine", Version: 2, Description: "等待第三方稳定性复核"}, ExcursionCode: "EE-HISTORY", DecisionBasis: "证据不足，维持隔离", SensorEvidence: "minio://sensor/history-511.csv", ProposedBy: "operator", ApprovedBy: "admin", DecidedAt: timePointer(now.Add(-24 * time.Hour)), Facility: "质量放行组", Owner: "operator", Category: "隔离", RiskLevel: "high", MetricValue: 11.2, MetricUnit: "C", EffectiveAt: now.Add(-24 * time.Hour), Evidence: "minio://sensor/history-511.csv", RelatedCode: "EE-HISTORY"},
	}
	return db.WithContext(ctx).Create(&items).Error
}

func seedSensorEvidence(ctx context.Context, db *gorm.DB) error {
	var count int64
	if err := db.WithContext(ctx).Model(&model.SensorEvidence{}).Count(&count).Error; err != nil || count > 0 {
		return err
	}
	now := time.Now().UTC()
	items := []model.SensorEvidence{
		{Code: "SE-001", ExcursionCode: "EE-001", ContainerCode: "TC-003", ObjectKey: "sensor/tc-003/excursion.csv", SHA256: strings.Repeat("a", 64), MediaType: "text/csv", SizeBytes: 18342, CapturedAt: now.Add(-35 * time.Minute), CapturedBy: "SN-TEMP-4103", Source: "calibrated-logger"},
		{Code: "SE-002", ExcursionCode: "EE-002", ContainerCode: "TC-002", ObjectKey: "sensor/tc-002/trace.csv", SHA256: strings.Repeat("b", 64), MediaType: "text/csv", SizeBytes: 9216, CapturedAt: now.Add(-2 * time.Hour), CapturedBy: "SN-TEMP-4102", Source: "calibrated-logger"},
		{Code: "SE-003", ExcursionCode: "EE-003", ContainerCode: "TC-001", ObjectKey: "sensor/tc-001/door-open.csv", SHA256: strings.Repeat("c", 64), MediaType: "text/csv", SizeBytes: 6740, CapturedAt: now.Add(-6 * time.Hour), CapturedBy: "SN-TEMP-4101", Source: "calibrated-logger"},
	}
	return db.WithContext(ctx).Create(&items).Error
}

func timePointer(value time.Time) *time.Time { return &value }
