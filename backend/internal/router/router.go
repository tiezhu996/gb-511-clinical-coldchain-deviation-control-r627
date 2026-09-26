package router

import (
	"log/slog"
	"net/http"

	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/config"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/handler"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/middleware"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/repository"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func New(cfg config.Config, db *gorm.DB, redisClient *redis.Client, logger *slog.Logger) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(middleware.RequestContext(logger))
	if len(cfg.TrustedProxies) > 0 {
		_ = engine.SetTrustedProxies(cfg.TrustedProxies)
	} else {
		_ = engine.SetTrustedProxies(nil)
	}

	securityRepository := repository.NewSecurityRepository(db)
	securityService := service.NewSecurityService(securityRepository, cfg)
	transportContainerRepository := repository.NewTransportContainerRepository(db)
	temperatureWindowRepository := repository.NewTemperatureWindowRepository(db)
	excursionEventRepository := repository.NewExcursionEventRepository(db)
	dispositionDecisionRepository := repository.NewDispositionDecisionRepository(db)
	sensorEvidenceRepository := repository.NewSensorEvidenceRepository(db)
	minioClient, minioErr := minio.New(cfg.MinIOEndpoint, &minio.Options{Creds: credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""), Secure: cfg.MinIOUseSSL})
	if minioErr != nil {
		logger.Error("object storage client unavailable", "error", minioErr)
		minioClient = nil
	}
	transportContainerService := service.NewTransportContainerService(transportContainerRepository, securityService)
	temperatureWindowService := service.NewTemperatureWindowService(temperatureWindowRepository, securityService)
	excursionEventService := service.NewExcursionEventService(excursionEventRepository, dispositionDecisionRepository, sensorEvidenceRepository, securityService)
	dispositionDecisionService := service.NewDispositionDecisionService(dispositionDecisionRepository, sensorEvidenceRepository, securityService)
	sensorEvidenceService := service.NewSensorEvidenceService(sensorEvidenceRepository, minioClient, cfg.MinIOBucket)
	transportContainerHandler := handler.NewTransportContainerHandler(transportContainerService)
	temperatureWindowHandler := handler.NewTemperatureWindowHandler(temperatureWindowService)
	excursionEventHandler := handler.NewExcursionEventHandler(excursionEventService)
	dispositionDecisionHandler := handler.NewDispositionDecisionHandler(dispositionDecisionService)
	sensorEvidenceHandler := handler.NewSensorEvidenceHandler(sensorEvidenceService)
	systemHandler := handler.NewSystemHandler(securityService, transportContainerService, temperatureWindowService, excursionEventService, dispositionDecisionService, db, redisClient)

	engine.GET("/healthz", systemHandler.Health)
	engine.POST("/api/auth/login", systemHandler.Login)

	limiter := middleware.NewLimiter(redisClient, cfg.RequestLimit)
	api := engine.Group("/api")
	api.Use(limiter.Middleware(), middleware.Authenticate(cfg, db))
	api.GET("/overview", systemHandler.Overview)
	api.GET("/audits", middleware.RequireMinimumRole("reviewer"), systemHandler.Audits)
	api.GET("/session", systemHandler.Session)
	api.GET("/runtime", systemHandler.Runtime)
	api.GET("/audit-summary", middleware.RequireMinimumRole("reviewer"), systemHandler.AuditSummary)
	api.GET("/audits/:entityType/:id", middleware.RequireMinimumRole("reviewer"), systemHandler.EntityHistory)
	transportContainerHandler.Register(api)
	temperatureWindowHandler.Register(api)
	excursionEventHandler.Register(api)
	dispositionDecisionHandler.Register(api)
	sensorEvidenceHandler.Register(api)

	engine.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "route_not_found"})
	})
	return engine
}
