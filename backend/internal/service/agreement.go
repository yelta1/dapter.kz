package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dapter-kz/internal/models"
	"dapter-kz/internal/repository"
)

// AgreementService описывает логику работы с долговыми договорами
type AgreementService interface {
	CreateAgreement(ctx context.Context, ownerID, shopID, customerCID string, limit float64, dueDate time.Time) (*models.Agreement, error)
	ConfirmAgreement(ctx context.Context, customerID, agreementID, smsCode, ip, ua string) error
	GetAgreementByID(ctx context.Context, userID string, id string) (*models.Agreement, error)
	GetAgreementsForUser(ctx context.Context, userID string, role models.UserRole) ([]*models.Agreement, error)
}

type agreementService struct {
	agreementRepo repository.AgreementRepository
	shopRepo      repository.ShopRepository
	userRepo      repository.UserRepository
	smsService    SmsService
	authService   AuthService // Для дешифрования данных ФИО
}

// NewAgreementService создает новый экземпляр сервиса договоров
func NewAgreementService(
	agreementRepo repository.AgreementRepository,
	shopRepo repository.ShopRepository,
	userRepo repository.UserRepository,
	smsService SmsService,
	authService AuthService,
) AgreementService {
	return &agreementService{
		agreementRepo: agreementRepo,
		shopRepo:      shopRepo,
		userRepo:      userRepo,
		smsService:    smsService,
		authService:   authService,
	}
}

func (s *agreementService) CreateAgreement(ctx context.Context, ownerID, shopID, customerCID string, limit float64, dueDate time.Time) (*models.Agreement, error) {
	// 1. Проверяем, что магазин существует и принадлежит данному владельцу
	shop, err := s.shopRepo.GetByID(ctx, shopID)
	if err != nil {
		return nil, err
	}
	if shop.OwnerID != ownerID {
		return nil, errors.New("доступ запрещен: этот магазин не принадлежит вам")
	}

	// 2. Ищем покупателя по 6-значному ID в системе
	customer, err := s.userRepo.GetByCID(ctx, customerCID)
	if err != nil {
		return nil, err
	}
	if customer == nil || customer.Role != models.RoleCustomer || customer.PinCodeHash == nil {
		return nil, errors.New("покупатель с таким ID не зарегистрирован или не активирован в системе")
	}

	// 3. Проверяем, нет ли уже активного/на рассмотрении договора для этой пары магазин-покупатель
	existing, err := s.agreementRepo.GetActiveByShopAndCustomer(ctx, shopID, customer.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("у покупателя уже есть действующий или ожидающий подтверждения договор в этом магазине (ID: %s)", existing.ID)
	}

	// 4. Создаем договор в статусе pending_confirmation
	agreement := &models.Agreement{
		ShopID:      shopID,
		CustomerID:  customer.ID,
		CreditLimit: limit,
		Balance:     0.0,
		DueDate:     dueDate,
		Status:      models.StatusPendingConfirmation,
	}

	err = s.agreementRepo.Create(ctx, agreement)
	if err != nil {
		return nil, err
	}

	// 5. Отправляем SMS-код подтверждения покупателю для подписания договора (простая ЭЦП)
	_, err = s.smsService.SendCode(ctx, customer.Phone, models.PurposeSignAgreement, &agreement.ID)
	if err != nil {
		return nil, fmt.Errorf("договор создан, но произошла ошибка при отправке SMS-кода: %w", err)
	}

	return agreement, nil
}

func (s *agreementService) ConfirmAgreement(ctx context.Context, customerID, agreementID, smsCode, ip, ua string) error {
	// 1. Получаем договор
	agreement, err := s.agreementRepo.GetByID(ctx, agreementID)
	if err != nil {
		return err
	}

	// 2. Проверяем, что договор принадлежит этому покупателю и находится в статусе ожидания
	if agreement.CustomerID != customerID {
		return errors.New("доступ запрещен: этот договор оформлен на другого пользователя")
	}
	if agreement.Status != models.StatusPendingConfirmation {
		return fmt.Errorf("договор не может быть подписан (текущий статус: %s)", agreement.Status)
	}

	// 3. Проверяем SMS-код
	verification, err := s.smsService.VerifyCode(ctx, agreement.CustomerPhone, models.PurposeSignAgreement, &agreementID, smsCode, ip, ua)
	if err != nil {
		return fmt.Errorf("ошибка проверки SMS: %w", err)
	}

	// 4. Переводим договор в активное состояние и сохраняем ЭЦП
	agreement.Status = models.StatusActive
	agreement.SignatureSmsID = &verification.ID
	now := time.Now()
	agreement.SignedAt = &now

	err = s.agreementRepo.Update(ctx, agreement)
	if err != nil {
		return fmt.Errorf("не удалось сохранить подписанный договор: %w", err)
	}

	return nil
}

func (s *agreementService) GetAgreementByID(ctx context.Context, userID string, id string) (*models.Agreement, error) {
	agreement, err := s.agreementRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Проверяем доступ: пользователь должен быть либо покупателем договора, либо владельцем магазина договора
	shop, err := s.shopRepo.GetByID(ctx, agreement.ShopID)
	if err != nil {
		return nil, err
	}

	if agreement.CustomerID != userID && shop.OwnerID != userID {
		return nil, errors.New("доступ запрещен: у вас нет прав на просмотр этого договора")
	}

	// Дешифруем ФИО покупателя для отображения
	customer, err := s.userRepo.GetByID(ctx, agreement.CustomerID)
	if err == nil {
		_ = s.authService.DecryptUserProfile(ctx, customer)
		agreement.CustomerName = customer.FullName
	}

	return agreement, nil
}

func (s *agreementService) GetAgreementsForUser(ctx context.Context, userID string, role models.UserRole) ([]*models.Agreement, error) {
	var list []*models.Agreement
	var err error

	if role == models.RoleCustomer {
		list, err = s.agreementRepo.GetByCustomerID(ctx, userID)
	} else if role == models.RoleOwner {
		// Получаем все магазины владельца
		shops, errShops := s.shopRepo.GetByOwnerID(ctx, userID)
		if errShops != nil {
			return nil, errShops
		}

		// Для каждого магазина собираем договоры
		for _, shop := range shops {
			shopAgreements, errA := s.agreementRepo.GetByShopID(ctx, shop.ID)
			if errA != nil {
				return nil, errA
			}
			list = append(list, shopAgreements...)
		}
	} else {
		return nil, errors.New("неизвестная роль пользователя")
	}

	if err != nil {
		return nil, err
	}

	// Расшифровываем имена покупателей для всех договоров в списке
	for _, agreement := range list {
		customer, errC := s.userRepo.GetByID(ctx, agreement.CustomerID)
		if errC == nil {
			_ = s.authService.DecryptUserProfile(ctx, customer)
			agreement.CustomerName = customer.FullName
		}
	}

	return list, nil
}
