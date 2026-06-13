package models

import (
	"time"
)

// Назначения для отправки SMS-кодов
const (
	PurposeRegister         = "register_confirmation"
	PurposeSignAgreement    = "sign_agreement"
	PurposeConfirmPurchase  = "confirm_purchase"
	PurposeConfirmRepayment = "confirm_repayment"
)

// SmsVerification представляет лог отправки и проверки SMS-кода
type SmsVerification struct {
	ID         string             `json:"id"`
	Phone      string             `json:"phone"`
	CodeHash   string             `json:"-"` // Хэш кода скрыт
	Purpose    string             `json:"purpose"`
	ReferenceID *string            `json:"reference_id,omitempty"` // ID договора или транзакции
	Status     ConfirmationStatus `json:"status"`
	ExpiresAt  time.Time          `json:"expires_at"`
	VerifiedAt *time.Time         `json:"verified_at,omitempty"`
	CreatedAt  time.Time          `json:"created_at"`
	IPAddress  *string            `json:"ip_address,omitempty"`
	UserAgent  *string            `json:"user_agent,omitempty"`
}
