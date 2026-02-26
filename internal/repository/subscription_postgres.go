package repository

import (
	"fmt"

	"github.com/BountyM/effectiveMobileTestTask/internal/models"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Subscription interface {
	Create(subscription models.Subscription) (uuid.UUID, error)
	Get(params models.SubscriptionParams) ([]models.Subscription, error)
	Delete(uuid.UUID) error
	Update(uuid uuid.UUID, subscription models.Subscription) error
	GetCost(params models.SubscriptionParams) (int64, error)
}

type SubscriptionPostgres struct {
	db *sqlx.DB
}

func NewSubscriptionPostgres(db *sqlx.DB) *SubscriptionPostgres {
	return &SubscriptionPostgres{
		db: db,
	}
}

func (r *SubscriptionPostgres) Create(subscription models.Subscription) (uuid.UUID, error) {
	builder := squirrel.Insert(models.SubscriptionTable).
		Columns(
			"id",
			"service_name",
			"price",
			"user_id",
			"start_date",
			"end_date",
		)
	id := uuid.New()
	// Подготавливаем значения
	values := []any{
		id,                       // id
		subscription.ServiceName, // service_name
		subscription.Price,       // price
		subscription.UserID,      // user_id
		subscription.StartDate,   // start_date (корректный формат)
	}

	// Добавляем end_date, только если он есть
	if subscription.EndDate != nil {
		values = append(values, *subscription.EndDate)
	} else {
		values = append(values, nil)
	}

	query, args, err := builder.Values(values...).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("SubscriptionPostgres Create() ошибка построения SQL-запроса: %w", err)
	}

	// Выполнение запроса
	_, err = r.db.Exec(query, args...)
	if err != nil {
		return uuid.Nil, fmt.Errorf("SubscriptionPostgres Create() ошибка выполнения SQL-запроса: %w", err)
	}
	return id, nil
}

func (r *SubscriptionPostgres) Get(params models.SubscriptionParams) ([]models.Subscription, error) {
	query := squirrel.Select(
		"id", "service_name", "price", "user_id", "start_date", "end_date").
		From(models.SubscriptionTable)

	// Фильтрация по пользователю
	if params.UserID != nil {
		query = query.Where(squirrel.Eq{"user_id": *params.UserID})
	}

	// Пагинация
	if params.Limit > 0 {
		query = query.Limit(uint64(params.Limit))
	}
	if params.Page > 0 && params.Limit > 0 {
		offset := (params.Page - 1) * params.Limit
		query = query.Offset(uint64(offset))
	}

	query = query.PlaceholderFormat(squirrel.Dollar)
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("SubscriptionPostgres Get() ошибка построения SQL-запроса: %w", err)
	}

	rows, err := r.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("SubscriptionPostgres Get() ошибка выполнения запроса: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var subscriptions []models.Subscription
	for rows.Next() {
		var sub models.Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&sub.EndDate,
		)
		if err != nil {
			return nil, fmt.Errorf("SubscriptionPostgres Get() ошибка сканирования строки: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("SubscriptionPostgres Get() ошибка итерации по строкам: %w", err)
	}

	return subscriptions, nil
}

func (r *SubscriptionPostgres) Delete(id uuid.UUID) error {
	query := squirrel.Delete(models.SubscriptionTable).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("SubscriptionPostgres Delete() ошибка построения SQL-запроса: %w", err)
	}

	result, err := r.db.Exec(sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("SubscriptionPostgres Delete() ошибка выполнения запроса: %w", err)
	}

	// Проверяем, была ли удалена запись
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("SubscriptionPostgres Delete() ошибка получения количества изменённых строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("SubscriptionPostgres Delete() запись с ID %s не найдена", id)
	}

	return nil
}

func (r *SubscriptionPostgres) Update(id uuid.UUID, subscription models.Subscription) error {
	builder := squirrel.Update(models.SubscriptionTable).
		Set("service_name", subscription.ServiceName).
		Set("price", subscription.Price).
		Set("user_id", subscription.UserID).
		Set("start_date", subscription.StartDate).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar)

	// Добавляем end_date, только если он есть
	if subscription.EndDate != nil {
		builder = builder.Set("end_date", *subscription.EndDate)
	} else {
		builder = builder.Set("end_date", nil)
	}

	sqlQuery, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("SubscriptionPostgres Update() ошибка построения SQL-запроса: %w", err)
	}

	result, err := r.db.Exec(sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("SubscriptionPostgres Update() ошибка выполнения запроса: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("SubscriptionPostgres Update() ошибка получения количества изменённых строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("SubscriptionPostgres Update() запись с ID %s не найдена", id)
	}

	return nil
}

func (r *SubscriptionPostgres) GetCost(params models.SubscriptionParams) (int64, error) {
	query := squirrel.Select("COALESCE(SUM(price), 0)").
		From(models.SubscriptionTable).PlaceholderFormat(squirrel.Dollar)

	if !params.StartDate.IsZero() {
		query = query.Where(squirrel.GtOrEq{"start_date": params.StartDate})
	}
	if !params.EndDate.IsZero() {
		query = query.Where(squirrel.LtOrEq{"end_date": params.EndDate})
	}

	if params.UserID != nil {
		query = query.Where(squirrel.Eq{"user_id": *params.UserID})
	}

	if params.ServiceName != "" {
		query = query.Where(squirrel.Eq{"service_name": params.ServiceName})
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("SubscriptionPostgres GetCost() ошибка построения SQL-запроса: %w", err)
	}

	var cost int64
	err = r.db.QueryRow(sqlQuery, args...).Scan(&cost)
	if err != nil {
		return 0, fmt.Errorf("SubscriptionPostgres GetCost() ошибка выполнения запроса: %w", err)
	}

	return cost, nil
}
