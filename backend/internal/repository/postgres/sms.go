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

type smsRepo struct {
	db *pgxpool.Pool
}

// NewSmsRepository создает новый экземпляр репозитория SMS-подтверждений
func NewSmsRepository(db *pgxpool.Pool) repository.SmsRepository {
	return &smsRepo{db: db}
}

func (r *smsRepo) Create(ctx context.Context, verification *models.SmsVerification) error {
	query := `
		INSERT INTO sms_verifications (phone, code_hash, purpose, reference_id, status, expires_at, verified_at, created_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), $8, $9)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(ctx, query,
		verification.Phone,
		verification.CodeHash,
		verification.Purpose,
		verification.ReferenceID,
		verification.Status,
		verification.ExpiresAt,
		verification.VerifiedAt,
		verification.IPAddress,
		verification.UserAgent,
	).Scan(&verification.ID, &verification.CreatedAt)
	if err != nil {
		return fmt.Errorf("ошибка сохранения SMS-верификации в БД: %w", err)
	}
	return nil
}

func (r *smsRepo) GetByID(ctx context.Context, id string) (*models.SmsVerification, error) {
	query := `
		SELECT id, phone, code_hash, purpose, reference_id, status, expires_at, verified_at, created_at, ip_address, user_agent
		FROM sms_verifications
		WHERE id = $1
	`
	var v models.SmsVerification
	err := r.db.QueryRow(ctx, query, id).Scan(
		&v.ID,
		&v.Phone,
		&v.CodeHash,
		&v.Purpose,
		&v.ReferenceID,
		&v.Status,
		&v.ExpiresAt,
		&v.VerifiedAt,
		&v.CreatedAt,
		&v.IPAddress,
		&v.UserAgent,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("запись SMS-верификации %s не найдена", id)
		}
		return nil, fmt.Errorf("ошибка получения SMS-верификации по ID: %w", err)
	}
	return &v, nil
}

func (r *smsRepo) GetActive(ctx context.Context, phone string, purpose string, referenceID *string) (*models.SmsVerification, error) {
	// Ищем не истекший и еще не подтвержденный код для этого телефона и назначения
	query := `
		SELECT id, phone, code_hash, purpose, reference_id, status, expires_at, verified_at, created_at, ip_address, user_agent
		FROM sms_verifications
		WHERE phone = $1 AND purpose = $2 AND status = 'pending' AND expires_at > NOW()
		  AND (reference_id = $3 OR ($3 IS NULL AND reference_id IS NULL))
		ORDER BY created_at DESC
		LIMIT 1
	`
	var v models.SmsVerification
	err := r.db.QueryRow(ctx, query, phone, purpose, referenceID).Scan(
		&v.ID,
		&v.Phone,
		&v.CodeHash,
		&v.Purpose,
		&v.ReferenceID,
		&v.Status,
		&v.ExpiresAt,
		&v.VerifiedAt,
		&v.CreatedAt,
		&v.IPAddress,
		&v.UserAgent,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Нет активного кода
		}
		return nil, fmt.Errorf("ошибка поиска активного SMS-кода: %w", err)
	}
	return &v, nil
}

func (r *smsRepo) Update(ctx context.Context, verification *models.SmsVerification) error {
	query := `
		UPDATE sms_verifications
		SET phone = $2, code_hash = $3, purpose = $4, reference_id = $5, 
		    status = $6, expires_at = $7, verified_at = $8, ip_address = $9, user_agent = $10
		WHERE id = $1
	`
	result, err := r.db.Exec(ctx, query,
		verification.ID,
		verification.Phone,
		verification.CodeHash,
		verification.Purpose,
		verification.ReferenceID,
		verification.Status,
		verification.ExpiresAt,
		verification.VerifiedAt,
		verification.IPAddress,
		verification.UserAgent,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления SMS-верификации: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("запись SMS-верификации для обновления не найдена")
	}
	return nil
}
