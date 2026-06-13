package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"dapter-kz/internal/config"
	"dapter-kz/internal/models"
	"dapter-kz/internal/pkg/crypto"
	"dapter-kz/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Claims описывает структуру JWT токена
type Claims struct {
	UserID string          `json:"user_id"`
	Phone  string          `json:"phone"`
	Role   models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// AuthService описывает бизнес-логику управления аккаунтами
type AuthService interface {
	RegisterOwner(ctx context.Context, phone, password, iin, fullName string) (*models.User, error)
	InitiateCustomerRegister(ctx context.Context, phone, iin, fullName string) (string, error)
	VerifyCustomerRegister(ctx context.Context, phone, code, ip, ua string) (string, error)
	SetCustomerPin(ctx context.Context, customerID string, pin string) error
	Login(ctx context.Context, phone, passwordOrPin string) (string, error)
	DecryptUserProfile(ctx context.Context, user *models.User) error
	GetProfile(ctx context.Context, id string) (*models.User, error)
	GetUsersByRole(ctx context.Context, role models.UserRole) ([]*models.User, error)
	GetUserByCID(ctx context.Context, cid string) (*models.User, error)
	SeedAdmin(ctx context.Context) error
}

type authService struct {
	userRepo   repository.UserRepository
	smsService SmsService
	cfg        *config.Config
}

// NewAuthService создает новый экземпляр сервиса авторизации
func NewAuthService(userRepo repository.UserRepository, smsService SmsService, cfg *config.Config) AuthService {
	return &authService{
		userRepo:   userRepo,
		smsService: smsService,
		cfg:        cfg,
	}
}

func (s *authService) RegisterOwner(ctx context.Context, phone, password, iin, fullName string) (*models.User, error) {
	phone = crypto.NormalizePhone(phone)
	// Проверка на существование телефона
	existingUser, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("пользователь с таким номером телефона уже существует")
	}

	// Проверка на уникальность ИИН
	iinHash := crypto.HashSha256(iin, s.cfg.JWTSecret)
	existingIIN, err := s.userRepo.GetByIINHash(ctx, iinHash)
	if err != nil {
		return nil, err
	}
	if existingIIN != nil {
		return nil, errors.New("пользователь с таким ИИН уже зарегистрирован")
	}

	// Хэширование пароля владельца
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("ошибка хэширования пароля: %w", err)
	}
	passwordHashStr := string(passHash)

	// Шифрование персональных данных
	iinEnc, err := crypto.Encrypt(iin, s.cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("ошибка шифрования ИИН: %w", err)
	}

	nameEnc, err := crypto.Encrypt(fullName, s.cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("ошибка шифрования ФИО: %w", err)
	}

	user := &models.User{
		Phone:             phone,
		PasswordHash:      &passwordHashStr,
		Role:              models.RoleOwner,
		IINEncrypted:      iinEnc,
		IINHash:           iinHash,
		FullNameEncrypted: nameEnc,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	// Заполняем открытые поля для возврата клиенту
	user.IIN = iin
	user.FullName = fullName

	return user, nil
}

func (s *authService) generateUniqueCID(ctx context.Context) (string, error) {
	for i := 0; i < 10; i++ {
		cidVal, err := rand.Int(rand.Reader, big.NewInt(900000))
		if err != nil {
			return "", err
		}
		cid := fmt.Sprintf("%d", 100000+cidVal.Int64())

		existing, err := s.userRepo.GetByCID(ctx, cid)
		if err != nil {
			return "", err
		}
		if existing == nil {
			return cid, nil
		}
	}
	return "", errors.New("не удалось сгенерировать уникальный CID покупателя")
}

func (s *authService) InitiateCustomerRegister(ctx context.Context, phone, iin, fullName string) (string, error) {
	phone = crypto.NormalizePhone(phone)
	// 1. Проверяем, существует ли уже пользователь
	existingUser, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return "", err
	}

	// Проверка уникальности ИИН
	iinHash := crypto.HashSha256(iin, s.cfg.JWTSecret)
	existingIIN, err := s.userRepo.GetByIINHash(ctx, iinHash)
	if err != nil {
		return "", err
	}

	// Шифрование данных покупателя
	iinEnc, err := crypto.Encrypt(iin, s.cfg.EncryptionKey)
	if err != nil {
		return "", fmt.Errorf("ошибка шифрования ИИН: %w", err)
	}
	nameEnc, err := crypto.Encrypt(fullName, s.cfg.EncryptionKey)
	if err != nil {
		return "", fmt.Errorf("ошибка шифрования ФИО: %w", err)
	}

	var user *models.User

	if existingUser != nil {
		// Если пользователь существует, но не завершил регистрацию (нет PIN-кода),
		// позволяем обновить данные и повторно выслать SMS.
		if existingUser.PinCodeHash == nil {
			user = existingUser
			user.IINEncrypted = iinEnc
			user.IINHash = iinHash
			user.FullNameEncrypted = nameEnc
			if user.CID == nil {
				cid, err := s.generateUniqueCID(ctx)
				if err != nil {
					return "", err
				}
				user.CID = &cid
			}
			err = s.userRepo.Update(ctx, user)
			if err != nil {
				return "", fmt.Errorf("ошибка обновления неактивного пользователя: %w", err)
			}
		} else {
			return "", errors.New("пользователь с таким номером телефона уже существует и активен")
		}
	} else {
		// Проверяем уникальность ИИН только для новых пользователей или активных
		if existingIIN != nil && existingIIN.PinCodeHash != nil {
			return "", errors.New("пользователь с таким ИИН уже зарегистрирован")
		}

		cid, err := s.generateUniqueCID(ctx)
		if err != nil {
			return "", fmt.Errorf("ошибка генерации CID покупателя: %w", err)
		}

		user = &models.User{
			Phone:             phone,
			Role:              models.RoleCustomer,
			CID:               &cid,
			IINEncrypted:      iinEnc,
			IINHash:           iinHash,
			FullNameEncrypted: nameEnc,
		}
		err = s.userRepo.Create(ctx, user)
		if err != nil {
			return "", err
		}
	}

	// 2. Отправляем SMS-код подтверждения
	smsVerificationID, err := s.smsService.SendCode(ctx, phone, models.PurposeRegister, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка отправки SMS кода: %w", err)
	}

	return smsVerificationID, nil
}

func (s *authService) VerifyCustomerRegister(ctx context.Context, phone, code, ip, ua string) (string, error) {
	phone = crypto.NormalizePhone(phone)
	// 1. Проверяем код подтверждения
	_, err := s.smsService.VerifyCode(ctx, phone, models.PurposeRegister, nil, code, ip, ua)
	if err != nil {
		return "", err
	}

	// 2. Получаем пользователя для генерации токена авторизации
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("пользователь не найден")
	}

	// Генерируем JWT токен (после верификации OTP покупатель может установить PIN-код)
	token, err := s.generateJWT(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *authService) SetCustomerPin(ctx context.Context, customerID string, pin string) error {
	if len(pin) != 4 {
		return errors.New("PIN-код должен состоять ровно из 4 цифр")
	}

	// Получаем пользователя по ID
	user, err := s.userRepo.GetByID(ctx, customerID)
	if err != nil {
		return err
	}

	if user.Role != models.RoleCustomer {
		return errors.New("установка PIN-кода доступна только для роли Покупатель")
	}

	// Хэшируем PIN-код (используем bcrypt для криптографической стойкости)
	pinHash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("ошибка хэширования PIN-кода: %w", err)
	}

	pinHashStr := string(pinHash)
	user.PinCodeHash = &pinHashStr

	return s.userRepo.Update(ctx, user)
}

func (s *authService) Login(ctx context.Context, phone, passwordOrPin string) (string, error) {
	phone = crypto.NormalizePhone(phone)
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("неверный номер телефона или пароль/PIN-код")
	}

	if user.Role == models.RoleOwner || user.Role == models.RoleAdmin {
		if user.PasswordHash == nil {
			return "", errors.New("учетная запись пользователя повреждена (отсутствует пароль)")
		}
		err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(passwordOrPin))
		if err != nil {
			return "", errors.New("неверный номер телефона или пароль")
		}
	} else if user.Role == models.RoleCustomer {
		if user.PinCodeHash == nil {
			return "", errors.New("регистрация не завершена. Пожалуйста, подтвердите телефон по SMS")
		}
		err = bcrypt.CompareHashAndPassword([]byte(*user.PinCodeHash), []byte(passwordOrPin))
		if err != nil {
			return "", errors.New("неверный номер телефона или PIN-код")
		}
	} else {
		return "", errors.New("неизвестная роль пользователя")
	}

	// Генерация JWT токена
	return s.generateJWT(user)
}

func (s *authService) DecryptUserProfile(ctx context.Context, user *models.User) error {
	// Расшифровка ИИН
	iin, err := crypto.Decrypt(user.IINEncrypted, s.cfg.EncryptionKey)
	if err != nil {
		return fmt.Errorf("ошибка расшифровки ИИН: %w", err)
	}
	user.IIN = iin

	// Расшифровка ФИО
	fullName, err := crypto.Decrypt(user.FullNameEncrypted, s.cfg.EncryptionKey)
	if err != nil {
		return fmt.Errorf("ошибка расшифровки ФИО: %w", err)
	}
	user.FullName = fullName

	return nil
}

func (s *authService) GetProfile(ctx context.Context, id string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	err = s.DecryptUserProfile(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *authService) GetUsersByRole(ctx context.Context, role models.UserRole) ([]*models.User, error) {
	users, err := s.userRepo.GetAllByRole(ctx, role)
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		_ = s.DecryptUserProfile(ctx, u)
	}
	return users, nil
}

func (s *authService) SeedAdmin(ctx context.Context) error {
	adminPhone := "+77777777777"
	existing, err := s.userRepo.GetByPhone(ctx, adminPhone)
	if err != nil {
		return err
	}
	if existing != nil {
		if existing.Role == models.RoleAdmin {
			log.Println("Учетная запись суперадминистратора уже существует")
			return nil
		}
		return fmt.Errorf("пользователь с номером %s уже существует, но имеет другую роль: %s", adminPhone, existing.Role)
	}

	log.Println("Создание учетной записи суперадминистратора по умолчанию...")

	passHash, err := bcrypt.GenerateFromPassword([]byte("adminpassword"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	passwordHashStr := string(passHash)

	iinEnc, err := crypto.Encrypt("000000000000", s.cfg.EncryptionKey)
	if err != nil {
		return err
	}
	iinHash := crypto.HashSha256("000000000000", s.cfg.JWTSecret)

	nameEnc, err := crypto.Encrypt("Администратор Системы", s.cfg.EncryptionKey)
	if err != nil {
		return err
	}

	admin := &models.User{
		Phone:             adminPhone,
		PasswordHash:      &passwordHashStr,
		Role:              models.RoleAdmin,
		IINEncrypted:      iinEnc,
		IINHash:           iinHash,
		FullNameEncrypted: nameEnc,
	}

	err = s.userRepo.Create(ctx, admin)
	if err != nil {
		return fmt.Errorf("ошибка создания админа: %w", err)
	}

	log.Println("Учетная запись суперадминистратора успешно создана (Телефон: +77777777777, Пароль: adminpassword)")
	return nil
}

func (s *authService) GetUserByCID(ctx context.Context, cid string) (*models.User, error) {
	user, err := s.userRepo.GetByCID(ctx, cid)
	if err != nil {
		return nil, err
	}
	if user != nil {
		_ = s.DecryptUserProfile(ctx, user)
	}
	return user, nil
}

func (s *authService) generateJWT(user *models.User) (string, error) {
	expirationTime := time.Now().Add(24 * 30 * time.Hour) // Токен на 30 дней для удобства
	claims := &Claims{
		UserID: user.ID,
		Phone:  user.Phone,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("ошибка подписания токена: %w", err)
	}

	return tokenString, nil
}
