package postgres

import (
	"context"
	"errors"
	"fmt"

	"dapter-kz/internal/models"
	"dapter-kz/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type agreementRepo struct {
	db *pgxpool.Pool
}

// NewAgreementRepository создает новый экземпляр репозитория договоров
func NewAgreementRepository(db *pgxpool.Pool) repository.AgreementRepository {
	return &agreementRepo{db: db}
}

func (r *agreementRepo) Create(ctx context.Context, agreement *models.Agreement) error {
	query := `
		INSERT INTO agreements (shop_id, customer_id, credit_limit, balance, due_date, status, signature_sms_id, signed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		agreement.ShopID,
		agreement.CustomerID,
		agreement.CreditLimit,
		agreement.Balance,
		agreement.DueDate,
		agreement.Status,
		agreement.SignatureSmsID,
		agreement.SignedAt,
	).Scan(&agreement.ID, &agreement.CreatedAt, &agreement.UpdatedAt)
	if err != nil {
		return fmt.Errorf("ошибка создания договора в БД: %w", err)
	}
	return nil
}

func (r *agreementRepo) GetByID(ctx context.Context, id string) (*models.Agreement, error) {
	query := `
		SELECT a.id, a.shop_id, a.customer_id, a.credit_limit, a.balance, a.due_date, a.status, 
		       a.signature_sms_id, a.signed_at, a.created_at, a.updated_at,
		       s.name as shop_name, u.phone as customer_phone, u.full_name_encrypted as customer_name
		FROM agreements a
		JOIN shops s ON a.shop_id = s.id
		JOIN users u ON a.customer_id = u.id
		WHERE a.id = $1
	`
	var agreement models.Agreement
	err := r.db.QueryRow(ctx, query, id).Scan(
		&agreement.ID,
		&agreement.ShopID,
		&agreement.CustomerID,
		&agreement.CreditLimit,
		&agreement.Balance,
		&agreement.DueDate,
		&agreement.Status,
		&agreement.SignatureSmsID,
		&agreement.SignedAt,
		&agreement.CreatedAt,
		&agreement.UpdatedAt,
		&agreement.ShopName,
		&agreement.CustomerPhone,
		&agreement.CustomerName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("договор с ID %s не найден", id)
		}
		return nil, fmt.Errorf("ошибка получения договора по ID: %w", err)
	}
	return &agreement, nil
}

func (r *agreementRepo) GetActiveByShopAndCustomer(ctx context.Context, shopID, customerID string) (*models.Agreement, error) {
	query := `
		SELECT a.id, a.shop_id, a.customer_id, a.credit_limit, a.balance, a.due_date, a.status, 
		       a.signature_sms_id, a.signed_at, a.created_at, a.updated_at,
		       s.name as shop_name, u.phone as customer_phone, u.full_name_encrypted as customer_name
		FROM agreements a
		JOIN shops s ON a.shop_id = s.id
		JOIN users u ON a.customer_id = u.id
		WHERE a.shop_id = $1 AND a.customer_id = $2 
		  AND a.status IN ('pending_confirmation', 'active', 'suspended')
		LIMIT 1
	`
	var agreement models.Agreement
	err := r.db.QueryRow(ctx, query, shopID, customerID).Scan(
		&agreement.ID,
		&agreement.ShopID,
		&agreement.CustomerID,
		&agreement.CreditLimit,
		&agreement.Balance,
		&agreement.DueDate,
		&agreement.Status,
		&agreement.SignatureSmsID,
		&agreement.SignedAt,
		&agreement.CreatedAt,
		&agreement.UpdatedAt,
		&agreement.ShopName,
		&agreement.CustomerPhone,
		&agreement.CustomerName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Возвращаем nil, nil если активный договор отсутствует
		}
		return nil, fmt.Errorf("ошибка получения активного договора по паре магазин-покупатель: %w", err)
	}
	return &agreement, nil
}

func (r *agreementRepo) GetByCustomerID(ctx context.Context, customerID string) ([]*models.Agreement, error) {
	query := `
		SELECT a.id, a.shop_id, a.customer_id, a.credit_limit, a.balance, a.due_date, a.status, 
		       a.signature_sms_id, a.signed_at, a.created_at, a.updated_at,
		       s.name as shop_name, u.phone as customer_phone, u.full_name_encrypted as customer_name
		FROM agreements a
		JOIN shops s ON a.shop_id = s.id
		JOIN users u ON a.customer_id = u.id
		WHERE a.customer_id = $1
		ORDER BY a.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения договоров покупателя: %w", err)
	}
	defer rows.Close()

	agreements := make([]*models.Agreement, 0)
	for rows.Next() {
		var agreement models.Agreement
		err := rows.Scan(
			&agreement.ID,
			&agreement.ShopID,
			&agreement.CustomerID,
			&agreement.CreditLimit,
			&agreement.Balance,
			&agreement.DueDate,
			&agreement.Status,
			&agreement.SignatureSmsID,
			&agreement.SignedAt,
			&agreement.CreatedAt,
			&agreement.UpdatedAt,
			&agreement.ShopName,
			&agreement.CustomerPhone,
			&agreement.CustomerName,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования записи договора: %w", err)
		}
		agreements = append(agreements, &agreement)
	}
	return agreements, nil
}

func (r *agreementRepo) GetByShopID(ctx context.Context, shopID string) ([]*models.Agreement, error) {
	query := `
		SELECT a.id, a.shop_id, a.customer_id, a.credit_limit, a.balance, a.due_date, a.status, 
		       a.signature_sms_id, a.signed_at, a.created_at, a.updated_at,
		       s.name as shop_name, u.phone as customer_phone, u.full_name_encrypted as customer_name
		FROM agreements a
		JOIN shops s ON a.shop_id = s.id
		JOIN users u ON a.customer_id = u.id
		WHERE a.shop_id = $1
		ORDER BY a.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, shopID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения договоров магазина: %w", err)
	}
	defer rows.Close()

	agreements := make([]*models.Agreement, 0)
	for rows.Next() {
		var agreement models.Agreement
		err := rows.Scan(
			&agreement.ID,
			&agreement.ShopID,
			&agreement.CustomerID,
			&agreement.CreditLimit,
			&agreement.Balance,
			&agreement.DueDate,
			&agreement.Status,
			&agreement.SignatureSmsID,
			&agreement.SignedAt,
			&agreement.CreatedAt,
			&agreement.UpdatedAt,
			&agreement.ShopName,
			&agreement.CustomerPhone,
			&agreement.CustomerName,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования записи договора: %w", err)
		}
		agreements = append(agreements, &agreement)
	}
	return agreements, nil
}

func (r *agreementRepo) Update(ctx context.Context, agreement *models.Agreement) error {
	query := `
		UPDATE agreements
		SET shop_id = $2, customer_id = $3, credit_limit = $4, balance = $5, 
		    due_date = $6, status = $7, signature_sms_id = $8, signed_at = $9, updated_at = NOW()
		WHERE id = $1
	`
	result, err := r.db.Exec(ctx, query,
		agreement.ID,
		agreement.ShopID,
		agreement.CustomerID,
		agreement.CreditLimit,
		agreement.Balance,
		agreement.DueDate,
		agreement.Status,
		agreement.SignatureSmsID,
		agreement.SignedAt,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления договора: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("договор для обновления не найден")
	}
	return nil
}
