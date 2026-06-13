package models

import (
	"time"
)

// AgreementStatus состояние долгового соглашения
type AgreementStatus string

const (
	StatusPendingConfirmation AgreementStatus = "pending_confirmation"
	StatusActive              AgreementStatus = "active"
	StatusClosed              AgreementStatus = "closed"
	StatusSuspended           AgreementStatus = "suspended"
)

// Agreement описывает договор лимита долга между магазином и покупателем
type Agreement struct {
	ID             string          `json:"id"`
	ShopID         string          `json:"shop_id"`
	CustomerID     string          `json:"customer_id"`
	CreditLimit    float64         `json:"credit_limit"`
	Balance        float64         `json:"balance"` // Текущий долг покупателя перед этим магазином
	DueDate        time.Time       `json:"due_date"`
	Status         AgreementStatus `json:"status"`
	SignatureSmsID *string         `json:"signature_sms_id,omitempty"` // ID SMS-подтверждения (ЭЦП)
	SignedAt       *time.Time      `json:"signed_at,omitempty"`        // Время подписания договора
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`

	// Вспомогательные поля для отображения на фронтенде
	ShopName      string `json:"shop_name,omitempty"`
	CustomerPhone string `json:"customer_phone,omitempty"`
	CustomerName  string `json:"customer_name,omitempty"`
}
