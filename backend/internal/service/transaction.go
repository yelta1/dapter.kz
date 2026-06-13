package service

import (
	"context"
	"errors"
	"fmt"

	"dapter-kz/internal/models"
	"dapter-kz/internal/repository"
)

// TransactionService описывает методы работы с покупками и погашениями долгов
type TransactionService interface {
	CreateTransaction(ctx context.Context, ownerID, agreementID string, txType models.TransactionType, amount float64, receiptImg *string) (*models.Transaction, error)
	ConfirmTransaction(ctx context.Context, customerID, transactionID, smsCode, ip, ua string) error
	RejectTransaction(ctx context.Context, customerID, transactionID string) error
	GetTransactionsByAgreement(ctx context.Context, userID string, agreementID string) ([]*models.Transaction, error)
}

type transactionService struct {
	transactionRepo repository.TransactionRepository
	agreementRepo   repository.AgreementRepository
	shopRepo        repository.ShopRepository
	userRepo        repository.UserRepository
	smsService      SmsService
}

// NewTransactionService создает новый экземпляр сервиса транзакций
func NewTransactionService(
	transactionRepo repository.TransactionRepository,
	agreementRepo repository.AgreementRepository,
	shopRepo repository.ShopRepository,
	userRepo repository.UserRepository,
	smsService SmsService,
) TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
		agreementRepo:   agreementRepo,
		shopRepo:        shopRepo,
		userRepo:        userRepo,
		smsService:      smsService,
	}
}

func (s *transactionService) CreateTransaction(ctx context.Context, ownerID, agreementID string, txType models.TransactionType, amount float64, receiptImg *string) (*models.Transaction, error) {
	// 1. Получаем договор и проверяем доступ владельца
	agreement, err := s.agreementRepo.GetByID(ctx, agreementID)
	if err != nil {
		return nil, err
	}

	shop, err := s.shopRepo.GetByID(ctx, agreement.ShopID)
	if err != nil {
		return nil, err
	}

	if shop.OwnerID != ownerID {
		return nil, errors.New("доступ запрещен: этот магазин не принадлежит вам")
	}

	if agreement.Status != models.StatusActive {
		return nil, fmt.Errorf("договор неактивен (текущий статус: %s), проведение операций невозможно", agreement.Status)
	}

	// 2. Валидация в зависимости от типа транзакции
	if txType == models.TypePurchase {
		if receiptImg == nil || *receiptImg == "" {
			return nil, errors.New("при фиксации покупки в долг фото чека является обязательным")
		}

		// Проверяем лимит перед тем как отправить SMS (быстрая проверка в памяти)
		if agreement.Balance+amount > agreement.CreditLimit {
			return nil, fmt.Errorf("лимит превышен (лимит: %.2f, долг: %.2f, покупка: %.2f)", 
				agreement.CreditLimit, agreement.Balance, amount)
		}
	}

	// 3. Создаем транзакцию в БД со статусом pending
	transaction := &models.Transaction{
		AgreementID:     agreementID,
		Type:            txType,
		Amount:          amount,
		ReceiptImageUrl: receiptImg,
		Status:          models.StatusPending,
	}

	err = s.transactionRepo.Create(ctx, transaction)
	if err != nil {
		return nil, err
	}

	// 4. Отправляем SMS-код подтверждения покупателю
	var purpose string
	if txType == models.TypePurchase {
		purpose = models.PurposeConfirmPurchase
	} else {
		purpose = models.PurposeConfirmRepayment
	}

	_, err = s.smsService.SendCode(ctx, agreement.CustomerPhone, purpose, &transaction.ID)
	if err != nil {
		return nil, fmt.Errorf("транзакция создана, но произошла ошибка при отправке SMS-кода: %w", err)
	}

	return transaction, nil
}

func (s *transactionService) ConfirmTransaction(ctx context.Context, customerID, transactionID, smsCode, ip, ua string) error {
	// 1. Получаем транзакцию
	transaction, err := s.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return err
	}

	// 2. Получаем договор для сверки покупателя
	agreement, err := s.agreementRepo.GetByID(ctx, transaction.AgreementID)
	if err != nil {
		return err
	}

	if agreement.CustomerID != customerID {
		return errors.New("доступ запрещен: эта операция принадлежит другому покупателю")
	}

	if transaction.Status != models.StatusPending {
		return fmt.Errorf("операция уже обработана (статус: %s)", transaction.Status)
	}

	// 3. Определяем тип SMS-подтверждения
	var purpose string
	if transaction.Type == models.TypePurchase {
		purpose = models.PurposeConfirmPurchase
	} else {
		purpose = models.PurposeConfirmRepayment
	}

	// 4. Проверяем SMS-код
	verification, err := s.smsService.VerifyCode(ctx, agreement.CustomerPhone, purpose, &transactionID, smsCode, ip, ua)
	if err != nil {
		return fmt.Errorf("ошибка верификации SMS-кода: %w", err)
	}

	// 5. Проводим операцию в БД-транзакции
	isRepayment := transaction.Type == models.TypeRepayment
	err = s.transactionRepo.ConfirmTransactionTx(ctx, transactionID, verification.ID, isRepayment)
	if err != nil {
		// В случае неудачи откатываем верификацию в модели для корректного отображения? 
		// Транзакция в БД сама откатится благодаря `tx.Rollback`.
		return fmt.Errorf("ошибка проведения транзакции: %w", err)
	}

	return nil
}

func (s *transactionService) RejectTransaction(ctx context.Context, customerID, transactionID string) error {
	// 1. Получаем транзакцию
	transaction, err := s.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return err
	}

	// 2. Сверяем покупателя
	agreement, err := s.agreementRepo.GetByID(ctx, transaction.AgreementID)
	if err != nil {
		return err
	}

	if agreement.CustomerID != customerID {
		return errors.New("доступ запрещен: вы не можете управлять чужой транзакцией")
	}

	if transaction.Status != models.StatusPending {
		return fmt.Errorf("транзакция уже обработана (статус: %s)", transaction.Status)
	}

	// 3. Обновляем статус
	transaction.Status = models.StatusRejected
	err = s.transactionRepo.Update(ctx, transaction)
	if err != nil {
		return fmt.Errorf("не удалось отклонить транзакцию: %w", err)
	}

	return nil
}

func (s *transactionService) GetTransactionsByAgreement(ctx context.Context, userID string, agreementID string) ([]*models.Transaction, error) {
	// Сначала проверяем права доступа пользователя к договору
	agreement, err := s.agreementRepo.GetByID(ctx, agreementID)
	if err != nil {
		return nil, err
	}

	shop, err := s.shopRepo.GetByID(ctx, agreement.ShopID)
	if err != nil {
		return nil, err
	}

	if agreement.CustomerID != userID && shop.OwnerID != userID {
		return nil, errors.New("доступ запрещен: вы не можете просматривать операции этого договора")
	}

	return s.transactionRepo.GetByAgreementID(ctx, agreementID)
}
