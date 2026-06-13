package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"dapter-kz/internal/models"
	"dapter-kz/internal/pkg/crypto"
	"dapter-kz/internal/repository"
)

// SmsService описывает методы работы с SMS-сообщениями
type SmsService interface {
	SendCode(ctx context.Context, phone string, purpose string, referenceID *string) (string, error)
	VerifyCode(ctx context.Context, phone string, purpose string, referenceID *string, code string, ip, ua string) (*models.SmsVerification, error)
}

type smsService struct {
	smsRepo       repository.SmsRepository
	salt          string // соль для дополнительной безопасности при хэшировании кодов
	idInstance    string
	tokenInstance string
	apiUrl        string
}

// NewSmsService создает новый экземпляр сервиса SMS/WhatsApp
func NewSmsService(smsRepo repository.SmsRepository, salt string, idInstance string, tokenInstance string, apiUrl string) SmsService {
	return &smsService{
		smsRepo:       smsRepo,
		salt:          salt,
		idInstance:    idInstance,
		tokenInstance: tokenInstance,
		apiUrl:        apiUrl,
	}
}

func (s *smsService) SendCode(ctx context.Context, phone string, purpose string, referenceID *string) (string, error) {
	phone = crypto.NormalizePhone(phone)
	// Генерация 4-значного случайного кода
	codeVal, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", fmt.Errorf("ошибка генерации случайного числа: %w", err)
	}
	code := fmt.Sprintf("%04d", codeVal.Int64())

	// Хэширование кода для защиты от утечки БД
	codeHash := crypto.HashSha256(code, s.salt)

	verification := &models.SmsVerification{
		Phone:       phone,
		CodeHash:    codeHash,
		Purpose:     purpose,
		ReferenceID: referenceID,
		Status:      models.StatusPending,
		ExpiresAt:   time.Now().Add(5 * time.Minute), // Код действителен 5 минут
	}

	err = s.smsRepo.Create(ctx, verification)
	if err != nil {
		return "", fmt.Errorf("ошибка сохранения SMS-кода: %w", err)
	}

	// Формируем текст сообщения в зависимости от цели
	var messageText string
	switch purpose {
	case models.PurposeSignAgreement:
		messageText = fmt.Sprintf("Подтверждение договора на Дэптер.kz. Код: %s", code)
	default:
		messageText = fmt.Sprintf("Подтверждение операции на Дэптер.kz. Код: %s", code)
	}

	// Проверяем, настроен ли реальный шлюз Green API
	isMock := s.idInstance == "" || strings.HasPrefix(s.idInstance, "mock-")

	if !isMock {
		// Очищаем номер от плюса
		cleanPhone := phone
		if strings.HasPrefix(cleanPhone, "+") {
			cleanPhone = cleanPhone[1:]
		}

		// Формируем JSON-тело
		requestBody, err := json.Marshal(map[string]string{
			"chatId":  cleanPhone + "@c.us",
			"message": messageText,
		})
		if err != nil {
			return "", fmt.Errorf("ошибка кодирования тела запроса Green API: %w", err)
		}

		// Выполняем POST запрос
		apiURL := fmt.Sprintf("%s/waInstance%s/sendMessage/%s", s.apiUrl, s.idInstance, s.tokenInstance)
		
		log.Printf("[GREEN API DEBUG] Sending message to %q via %s, text: %q\n", 
			cleanPhone, apiURL, messageText)

		req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(requestBody))
		if err != nil {
			return "", fmt.Errorf("ошибка создания запроса Green API: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("ошибка отправки запроса в Green API: %w", err)
		}
		defer resp.Body.Close()

		// Читаем ответ целиком
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("ошибка чтения ответа Green API: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("ошибка отправки через Green API (статус %d): %s", resp.StatusCode, string(bodyBytes))
		}

		log.Printf("[WHATSAPP GATEWAY] Успешно отправлен реальный код на WhatsApp %s через Green API\n", phone)
	} else {
		// Заглушка отправки SMS (вывод в логи сервера)
		log.Printf("\n[SMS GATEWAY] [MOCK] Отправлено SMS на номер %s\nТекст: %s\n[VERIFICATION_ID]: %s\n", 
			phone, messageText, verification.ID)
	}

	// Запись в файл для авто-тестирования
	_ = os.MkdirAll("uploads", 0755)
	_ = os.WriteFile("uploads/latest_sms.txt", []byte(code), 0644)

	return verification.ID, nil
}

func (s *smsService) VerifyCode(ctx context.Context, phone string, purpose string, referenceID *string, code string, ip, ua string) (*models.SmsVerification, error) {
	phone = crypto.NormalizePhone(phone)
	// 1. Ищем активный SMS-код
	verification, err := s.smsRepo.GetActive(ctx, phone, purpose, referenceID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при поиске активного кода: %w", err)
	}

	if verification == nil {
		return nil, fmt.Errorf("активный код подтверждения не найден или срок его действия истек")
	}

	// 2. Сверяем хэш кода
	expectedHash := crypto.HashSha256(code, s.salt)
	if verification.CodeHash != expectedHash {
		return nil, fmt.Errorf("неверный код подтверждения")
	}

	// 3. Отмечаем код как подтвержденный в БД
	now := time.Now()
	verification.Status = models.StatusCompleted
	verification.VerifiedAt = &now
	verification.IPAddress = &ip
	verification.UserAgent = &ua

	err = s.smsRepo.Update(ctx, verification)
	if err != nil {
		return nil, fmt.Errorf("не удалось обновить статус SMS-кода: %w", err)
	}

	return verification, nil
}
