package models

import (
	"time"
)

// TransactionType тип транзакции (покупка в долг или погашение)
type TransactionType string

const (
	TypePurchase  TransactionType = "purchase"
	TypeRepayment TransactionType = "repayment"
)

// ConfirmationStatus статус подтверждения транзакции покупателем
type ConfirmationStatus string

const (
	StatusPending   ConfirmationStatus = "pending"
	StatusCompleted ConfirmationStatus = "completed"
	StatusExpired   ConfirmationStatus = "expired"
	StatusRejected  ConfirmationStatus = "rejected"
)

// Transaction описывает отдельную покупку или платеж по договору
type Transaction struct {
	ID              string             `json:"id"`
	AgreementID     string             `json:"agreement_id"`
	Type            TransactionType    `json:"type"`
	Amount          float64            `json:"amount"`
	ReceiptImageUrl *string            `json:"receipt_image_url,omitempty"` // Только для покупок в долг
	Status          ConfirmationStatus `json:"status"`
	SignatureSmsID  *string            `json:"signature_sms_id,omitempty"` // Ссылка на ЭЦП SMS-верификации
	ConfirmedAt     *time.Time         `json:"confirmed_at,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`

	// Вспомогательные поля
	ShopName      string `json:"shop_name,omitempty"`
	CustomerPhone string `json:"customer_phone,omitempty"`
}
