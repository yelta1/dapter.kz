package repository

import (
	"context"
	"dapter-kz/internal/models"
)

// UserRepository описывает работу с пользователями
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByPhone(ctx context.Context, phone string) (*models.User, error)
	GetByIINHash(ctx context.Context, iinHash string) (*models.User, error)
	GetByCID(ctx context.Context, cid string) (*models.User, error)
	GetAllByRole(ctx context.Context, role models.UserRole) ([]*models.User, error)
	Update(ctx context.Context, user *models.User) error
}

// ShopRepository описывает работу с магазинами
type ShopRepository interface {
	Create(ctx context.Context, shop *models.Shop) error
	GetByID(ctx context.Context, id string) (*models.Shop, error)
	GetByOwnerID(ctx context.Context, ownerID string) ([]*models.Shop, error)
	GetAll(ctx context.Context) ([]*models.Shop, error)
}

// AgreementRepository описывает работу с долговыми договорами
type AgreementRepository interface {
	Create(ctx context.Context, agreement *models.Agreement) error
	GetByID(ctx context.Context, id string) (*models.Agreement, error)
	GetActiveByShopAndCustomer(ctx context.Context, shopID, customerID string) (*models.Agreement, error)
	GetByCustomerID(ctx context.Context, customerID string) ([]*models.Agreement, error)
	GetByShopID(ctx context.Context, shopID string) ([]*models.Agreement, error)
	Update(ctx context.Context, agreement *models.Agreement) error
}

// TransactionRepository описывает работу с транзакциями покупок и погашений
type TransactionRepository interface {
	Create(ctx context.Context, transaction *models.Transaction) error
	GetByID(ctx context.Context, id string) (*models.Transaction, error)
	GetByAgreementID(ctx context.Context, agreementID string) ([]*models.Transaction, error)
	Update(ctx context.Context, transaction *models.Transaction) error
	
	// ConfirmTransactionTx атомарно в транзакции БД подтверждает покупку/погашение 
	// и обновляет текущий баланс договора.
	ConfirmTransactionTx(ctx context.Context, transactionID string, signatureSmsID string, isRepayment bool) error
}

// SmsRepository описывает работу с SMS-верификациями
type SmsRepository interface {
	Create(ctx context.Context, verification *models.SmsVerification) error
	GetByID(ctx context.Context, id string) (*models.SmsVerification, error)
	// GetActive возвращает действующее, не истекшее по времени SMS-подтверждение по номеру телефона
	GetActive(ctx context.Context, phone string, purpose string, referenceID *string) (*models.SmsVerification, error)
	Update(ctx context.Context, verification *models.SmsVerification) error
}
