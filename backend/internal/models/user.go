package models

import (
	"time"
)

// UserRole тип роли пользователя (владелец магазина или покупатель)
type UserRole string

const (
	RoleOwner    UserRole = "owner"
	RoleCustomer UserRole = "customer"
	RoleAdmin    UserRole = "admin"
)

// User описывает структуру пользователя в БД
type User struct {
	ID                string     `json:"id"`
	Phone             string     `json:"phone"`
	PasswordHash      *string    `json:"-"` // Пароль скрываем при JSON-сериализации
	PinCodeHash       *string    `json:"-"` // PIN-код скрываем при JSON-сериализации
	Role              UserRole   `json:"role"`
	CID               *string    `json:"cid,omitempty"` // 6-значный ID покупателя
	IINEncrypted      string     `json:"-"` // Внутреннее зашифрованное значение ИИН
	IINHash           string     `json:"-"` // Хэш ИИН для поиска
	FullNameEncrypted string     `json:"-"` // Внутреннее зашифрованное ФИО
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	// Дешифрованные данные (заполняются в сервис-слое перед отправкой клиенту)
	IIN      string `json:"iin,omitempty"`
	FullName string `json:"full_name,omitempty"`
}
