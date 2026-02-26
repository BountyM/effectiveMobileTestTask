package handler

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/BountyM/effectiveMobileTestTask/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	services *service.Service
	logger   *slog.Logger
}

func New(services *service.Service, logger *slog.Logger) *Handler {
	return &Handler{
		services: services,
		logger:   logger,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	// Добавляем middleware
	router.Use(gin.Recovery())      // Стандартный recovery middleware
	router.Use(h.loggingMiddleware) // Ваш кастомный logging middleware

	// Swagger UI: доступен по /swagger/index.html
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	subscription := router.Group("/subscription")
	subscription.POST("/", h.createSubscription)
	subscription.GET("/:user_id/:page/:limit", h.getSubscriptions)
	subscription.GET("/:user_id", h.getSubscriptions)
	subscription.DELETE("/:id", h.deleteSubscription)
	subscription.PUT("/:id", h.updateSubscription)
	subscription.GET("/cost", h.getCost)

	return router
}

// Middleware для логирования
func (h *Handler) loggingMiddleware(c *gin.Context) {
	// Генерируем уникальный ID для запроса
	requestID := generateRequestID()

	// Добавляем request_id в заголовки ответа (полезно для клиента)
	c.Header("X-Request-ID", requestID)

	// Создаём логгер с request_id для этого запроса
	reqLogger := h.logger.With(slog.String("request_id", requestID))

	// Сохраняем логгер в контексте Gin
	c.Set("logger", reqLogger)

	start := time.Now()
	// Продолжаем обработку запроса
	c.Next()

	// Логируем информацию о запросе
	duration := time.Since(start)
	reqLogger.Info(
		"HTTP Request",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"status", c.Writer.Status(),
		"duration", duration,
		"client_ip", c.ClientIP(),
		"request_id", requestID,
	)
}

// Функция генерации request_id
func generateRequestID() string {
	id, err := uuid.NewRandom()
	if err != nil {
		// В случае ошибки — используем временную метку
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return id.String()
}

// getRequestLogger получает логгер из контекста запроса
func (h *Handler) getRequestLogger(c *gin.Context) *slog.Logger {
	if logger, exists := c.Get("logger"); exists {
		if reqLogger, ok := logger.(*slog.Logger); ok {
			return reqLogger
		}
	}
	// Если логгер не найден — возвращаем базовый
	return h.logger
}
