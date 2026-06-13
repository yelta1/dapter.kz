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

type transactionRepo struct {
	db *pgxpool.Pool
}

// NewTransactionRepository создает новый экземпляр репозитория транзакций
func NewTransactionRepository(db *pgxpool.Pool) repository.TransactionRepository {
	return &transactionRepo{db: db}
}

func (r *transactionRepo) Create(ctx context.Context, transaction *models.Transaction) error {
	query := `
		INSERT INTO transactions (agreement_id, type, amount, receipt_image_url, status, signature_sms_id, confirmed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		transaction.AgreementID,
		transaction.Type,
		transaction.Amount,
		transaction.ReceiptImageUrl,
		transaction.Status,
		transaction.SignatureSmsID,
		transaction.ConfirmedAt,
	).Scan(&transaction.ID, &transaction.CreatedAt, &transaction.UpdatedAt)
	if err != nil {
		return fmt.Errorf("ошибка создания транзакции в БД: %w", err)
	}
	return nil
}

func (r *transactionRepo) GetByID(ctx context.Context, id string) (*models.Transaction, error) {
	query := `
		SELECT t.id, t.agreement_id, t.type, t.amount, t.receipt_image_url, t.status, 
		       t.signature_sms_id, t.confirmed_at, t.created_at, t.updated_at,
		       s.name as shop_name, u.phone as customer_phone
		FROM transactions t
		JOIN agreements a ON t.agreement_id = a.id
		JOIN shops s ON a.shop_id = s.id
		JOIN users u ON a.customer_id = u.id
		WHERE t.id = $1
	`
	var transaction models.Transaction
	err := r.db.QueryRow(ctx, query, id).Scan(
		&transaction.ID,
		&transaction.AgreementID,
		&transaction.Type,
		&transaction.Amount,
		&transaction.ReceiptImageUrl,
		&transaction.Status,
		&transaction.SignatureSmsID,
		&transaction.ConfirmedAt,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
		&transaction.ShopName,
		&transaction.CustomerPhone,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("транзакция с ID %s не найдена", id)
		}
		return nil, fmt.Errorf("ошибка получения транзакции по ID: %w", err)
	}
	return &transaction, nil
}

func (r *transactionRepo) GetByAgreementID(ctx context.Context, agreementID string) ([]*models.Transaction, error) {
	query := `
		SELECT t.id, t.agreement_id, t.type, t.amount, t.receipt_image_url, t.status, 
		       t.signature_sms_id, t.confirmed_at, t.created_at, t.updated_at,
		       s.name as shop_name, u.phone as customer_phone
		FROM transactions t
		JOIN agreements a ON t.agreement_id = a.id
		JOIN shops s ON a.shop_id = s.id
		JOIN users u ON a.customer_id = u.id
		WHERE t.agreement_id = $1
		ORDER BY t.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, agreementID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка транзакций договора: %w", err)
	}
	defer rows.Close()

	transactions := make([]*models.Transaction, 0)
	for rows.Next() {
		var transaction models.Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.AgreementID,
			&transaction.Type,
			&transaction.Amount,
			&transaction.ReceiptImageUrl,
			&transaction.Status,
			&transaction.SignatureSmsID,
			&transaction.ConfirmedAt,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
			&transaction.ShopName,
			&transaction.CustomerPhone,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования записи транзакции: %w", err)
		}
		transactions = append(transactions, &transaction)
	}
	return transactions, nil
}

func (r *transactionRepo) Update(ctx context.Context, transaction *models.Transaction) error {
	query := `
		UPDATE transactions
		SET agreement_id = $2, type = $3, amount = $4, receipt_image_url = $5, 
		    status = $6, signature_sms_id = $7, confirmed_at = $8, updated_at = NOW()
		WHERE id = $1
	`
	result, err := r.db.Exec(ctx, query,
		transaction.ID,
		transaction.AgreementID,
		transaction.Type,
		transaction.Amount,
		transaction.ReceiptImageUrl,
		transaction.Status,
		transaction.SignatureSmsID,
		transaction.ConfirmedAt,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления транзакции: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("транзакция для обновления не найдена")
	}
	return nil
}

func (r *transactionRepo) ConfirmTransactionTx(ctx context.Context, transactionID string, signatureSmsID string, isRepayment bool) error {
	// Начинаем транзакцию базы данных
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию БД: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Получаем и блокируем транзакцию для предотвращения гонок
	var agreementID string
	var amount float64
	var status models.ConfirmationStatus
	
	err = tx.QueryRow(ctx, 
		"SELECT agreement_id, amount, status FROM transactions WHERE id = $1 FOR UPDATE", 
		transactionID,
	).Scan(&agreementID, &amount, &status)
	if err != nil {
		return fmt.Errorf("ошибка получения/блокировки транзакции: %w", err)
	}

	if status != models.StatusPending {
		return fmt.Errorf("транзакция уже обработана (текущий статус: %s)", status)
	}

	// 2. Получаем и блокируем договор для проверки баланса и лимита
	var creditLimit float64
	var currentBalance float64
	var agreementStatus models.AgreementStatus

	err = tx.QueryRow(ctx,
		"SELECT credit_limit, balance, status FROM agreements WHERE id = $1 FOR UPDATE",
		agreementID,
	).Scan(&creditLimit, &currentBalance, &agreementStatus)
	if err != nil {
		return fmt.Errorf("ошибка получения/блокировки договора: %w", err)
	}

	if agreementStatus != models.StatusActive {
		return fmt.Errorf("договор не в активном статусе (текущий статус: %s)", agreementStatus)
	}

	var newBalance float64
	if isRepayment {
		newBalance = currentBalance - amount
		if newBalance < 0 {
			// Допускаем переплату (баланс уходит в минус - переплата) или запрещаем? 
			// Обычно в таких системах переплата просто уменьшает долг до 0. Для строгости разрешим баланс >= 0.
			// Если баланс уходит в минус, это означает, что покупатель платит больше, чем должен.
			// Позволим уйти в минус (аванс) либо ограничим 0. Установим ограничение 0 для логичности долга.
			if newBalance < 0 {
				newBalance = 0
			}
		}
	} else {
		newBalance = currentBalance + amount
		if newBalance > creditLimit {
			return fmt.Errorf("превышен кредитный лимит по договору (лимит: %.2f, текущий долг: %.2f, сумма покупки: %.2f)", 
				creditLimit, currentBalance, amount)
		}
	}

	// 3. Обновляем транзакцию
	_, err = tx.Exec(ctx, 
		`UPDATE transactions 
		 SET status = 'completed', signature_sms_id = $2, confirmed_at = NOW(), updated_at = NOW() 
		 WHERE id = $1`,
		transactionID,
		signatureSmsID,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса транзакции: %w", err)
	}

	// 4. Обновляем баланс договора
	_, err = tx.Exec(ctx,
		"UPDATE agreements SET balance = $2, updated_at = NOW() WHERE id = $1",
		agreementID,
		newBalance,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления баланса договора: %w", err)
	}

	// Фиксируем транзакцию БД
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("ошибка фиксации (commit) транзакции БД: %w", err)
	}

	return nil
}
