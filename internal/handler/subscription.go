package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/BountyM/effectiveMobileTestTask/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateSubscriptionRequest model
type reqCreate struct {
	ServiceName string    `json:"service_name"`
	Price       int64     `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date,omitempty"`
}

func reqToSubscription(r reqCreate) (models.Subscription, error) {
	start, err := time.Parse("01-2006", r.StartDate)
	if err != nil {
		return models.Subscription{}, errors.New("invalid start_date format, expected MM-YYYY")
	}

	var end *time.Time
	if r.EndDate != "" {
		parsedEnd, err := time.Parse("01-2006", r.EndDate)
		if err != nil {
			return models.Subscription{}, errors.New("invalid end_date format, expected MM-YYYY")
		}
		// Дополнительная проверка: end_date должна быть после start_date
		if parsedEnd.Before(start) {
			return models.Subscription{}, errors.New("end_date must be after start_date")
		}
		end = &parsedEnd
	}

	return models.Subscription{
		ServiceName: r.ServiceName,
		Price:       r.Price,
		UserID:      r.UserID,
		StartDate:   start,
		EndDate:     end,
	}, nil
}

// validateCreate проверяет обязательные поля
func validateCreate(r reqCreate) error {
	if r.ServiceName == "" {
		return errors.New("service_name is required")
	}
	if r.Price <= 0 {
		return errors.New("price must be positive")
	}
	if r.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if r.StartDate == "" {
		return errors.New("start_date is required")
	}
	return nil
}

// @Summary Создать подписку
// @Description Создаёт новую подписку для пользователя
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body reqCreate true "Данные подписки"
// @Success 200 {object} object{res=string,uuid=string} "Успешное создание, возвращает ID подписки"
// @Failure 400 {object} object{error=string} "Некорректные данные: invalid input body"
// @Failure 500 {object} object{error=string} "Внутренняя ошибка сервера: internal error"
// @Router /subscription [post]
func (h *Handler) createSubscription(c *gin.Context) {
	logger := h.getRequestLogger(c)

	var r reqCreate
	if err := c.BindJSON(&r); err != nil {
		logger.Warn("invalid JSON body", "error", err)
		newErrorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// Валидация обязательных полей
	if err := validateCreate(r); err != nil {
		logger.Warn("validation failed", "error", err)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	subscription, err := reqToSubscription(r)
	if err != nil {
		logger.Warn("invalid date format or logic", "error", err)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.services.Create(subscription)
	if err != nil {
		logger.Error("failed to create subscription", "error", err)
		newErrorResponse(c, http.StatusInternalServerError, "internal server error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"res":  "ok",
		"uuid": id,
	})
}

// @Summary Получить подписки пользователя
// @Description Возвращает список подписок пользователя с пагинацией. Если page или limit не указаны, используются значения по умолчанию: page=1, limit=10.
// @Tags subscriptions
// @Produce json
// @Param user_id path string true "ID пользователя" format:"uuid"
// @Param page path int false "Номер страницы" minimum:"1" default:"1"
// @Param limit path int false "Количество записей на страницу" minimum:"1" maximum:"100" default:"10"
// @Success 200 {object} object{res=string,subscriptions=[]models.Subscription} "Список подписок с пагинацией"
// @Failure 400 {object} object{error=string} "Некорректный ID пользователя: invalid input body"
// @Failure 500 {object} object{error=string} "Внутренняя ошибка сервера: internal error"
// @Router /subscription/{user_id}/{page}/{limit} [get]
// @Router /subscription/{user_id} [get]
func (h *Handler) getSubscriptions(c *gin.Context) {
	logger := h.getRequestLogger(c)

	// Проверяем обязательный параметр user_id
	userIDStr := c.Param("user_id")
	if userIDStr == "" {
		newErrorResponse(c, http.StatusBadRequest, "user_id is required")
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Warn("invalid user_id format", "error", err)
		newErrorResponse(c, http.StatusBadRequest, "invalid user_id format")
		return
	}

	// Парсим page и limit со значениями по умолчанию
	page := 1
	limit := 10

	if pageStr := c.Param("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Param("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	params := models.SubscriptionParams{
		UserID: &userID,
		Page:   page,
		Limit:  limit,
	}

	subscriptions, err := h.services.Get(params)
	if err != nil {
		logger.Error("failed to get subscriptions", "error", err)
		newErrorResponse(c, http.StatusInternalServerError, "internal server error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"res":           "ok",
		"subscriptions": subscriptions,
	})
}

// @Summary Удалить подписку
// @Description Удаляет подписку по её ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки" format:"uuid"
// @Success 200 {object} object{res=string} "Успешное удаление"
// @Failure 400 {object} object{error=string} "Некорректный ID подписки: invalid input body"
// @Failure 500 {object} object{error=string} "Внутренняя ошибка сервера: internal error"
// @Router /subscription/{id} [delete]
func (h *Handler) deleteSubscription(c *gin.Context) {
	logger := h.getRequestLogger(c)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Warn("invalid subscription id format", "error", err)
		newErrorResponse(c, http.StatusBadRequest, "invalid subscription id")
		return
	}

	if err := h.services.Delete(id); err != nil {
		logger.Error("failed to delete subscription", "error", err)
		newErrorResponse(c, http.StatusInternalServerError, "internal server error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"res": "ok"})
}

// @Summary Обновить подписку
// @Description Обновляет данные подписки по ID. Принимает JSON с данными подписки.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки" format:"uuid"
// @Param request body reqCreate true "Обновляемые данные подписки"
// @Success 200 {object} object{res=string} "Успешное обновление"
// @Failure 400 {object} object{error=string} "Некорректные данные: invalid input body"
// @Failure 500 {object} object{error=string} "Внутренняя ошибка сервера: internal error"
// @Router /subscription/{id} [put]
func (h *Handler) updateSubscription(c *gin.Context) {
	logger := h.getRequestLogger(c)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Warn("invalid subscription id format", "error", err)
		newErrorResponse(c, http.StatusBadRequest, "invalid subscription id")
		return
	}

	var r reqCreate
	if err := c.BindJSON(&r); err != nil {
		logger.Warn("invalid JSON body", "error", err)
		newErrorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// При обновлении можно разрешить частичное обновление, но для простоты проверим обязательные поля
	if err := validateCreate(r); err != nil {
		logger.Warn("validation failed", "error", err)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	subscription, err := reqToSubscription(r)
	if err != nil {
		logger.Warn("invalid date format or logic", "error", err)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.services.Update(id, subscription); err != nil {
		logger.Error("failed to update subscription", "error", err)
		newErrorResponse(c, http.StatusInternalServerError, "internal server error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"res": "ok"})
}

// CostRequest model
type reqCost struct {
	UserID      *uuid.UUID `json:"user_id"`
	ServiceName string     `json:"service_name"`
	StartDate   string     `json:"start_date"`
	EndDate     string     `json:"end_date"`
}

func reqToSubscriptionParams(r reqCost) (params models.SubscriptionParams, err error) {
	var start time.Time
	if r.StartDate != "" {
		start, err = time.Parse("01-2006", r.StartDate)
		if err != nil {
			return models.SubscriptionParams{}, errors.New("invalid start_date format, expected MM-YYYY")
		}
		params.StartDate = start
	}

	var end time.Time
	if r.EndDate != "" {
		end, err = time.Parse("01-2006", r.EndDate)
		if err != nil {
			return models.SubscriptionParams{}, errors.New("invalid end_date format, expected MM-YYYY")
		}
		params.EndDate = end
	}

	params.UserID = r.UserID
	params.ServiceName = r.ServiceName
	return
}

func validateCost(params models.SubscriptionParams) error {

	if !params.StartDate.IsZero() && !params.EndDate.IsZero() && params.EndDate.Before(params.StartDate) {
		return errors.New("end_date must be after start_date")
	}
	return nil
}

// @Summary Рассчитать стоимость подписок
// @Description Рассчитывает общую стоимость подписок пользователя за указанный период
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body reqCost true "Параметры расчёта стоимости"
// @Success 200 {object} object{res=string,cost=number} "Успешный расчёт, возвращает стоимость"
// @Failure 400 {object} object{error=string} "Некорректные данные: invalid input body"
// @Failure 500 {object} object{error=string} "Внутренняя ошибка сервера: internal error"
// @Router /subscription/cost [post]
func (h *Handler) getCost(c *gin.Context) {
	logger := h.getRequestLogger(c)

	var r reqCost
	if err := c.BindJSON(&r); err != nil {
		logger.Warn("invalid JSON body", "error", err)
		newErrorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}

	params, err := reqToSubscriptionParams(r)
	if err != nil {
		logger.Warn("invalid date format", "error", err)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateCost(params); err != nil {
		logger.Warn("validation failed", "error", err)
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	cost, err := h.services.GetCost(params)
	if err != nil {
		logger.Error("failed to calculate cost", "error", err)
		newErrorResponse(c, http.StatusInternalServerError, "internal server error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"res":  "ok",
		"cost": cost,
	})
}
