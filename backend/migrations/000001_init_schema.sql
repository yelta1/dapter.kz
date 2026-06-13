-- +migrate Up

-- Включение расширения для UUID, если требуется генерация на стороне БД
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Типы ролей пользователей
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
        CREATE TYPE user_role AS ENUM ('owner', 'customer', 'admin');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'agreement_status') THEN
        CREATE TYPE agreement_status AS ENUM ('pending_confirmation', 'active', 'closed', 'suspended');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'transaction_type') THEN
        CREATE TYPE transaction_type AS ENUM ('purchase', 'repayment');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'confirmation_status') THEN
        CREATE TYPE confirmation_status AS ENUM ('pending', 'completed', 'expired', 'rejected');
    END IF;
END $$;

-- Убедимся, что роль admin добавлена, если тип уже был создан ранее
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'admin';

-- Таблица пользователей (Владельцы и Покупатели)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(20) UNIQUE NOT NULL, -- Номер телефона (индекс для авторизации)
    password_hash VARCHAR(255),        -- Хэш пароля для Владельцев магазина
    pin_code_hash VARCHAR(255),        -- Хэш 4-значного PIN-кода для Покупателей (задается после OTP)
    role user_role NOT NULL,           -- Роль (owner / customer / admin)
    cid VARCHAR(6) UNIQUE,             -- 6-значный ID покупателя (уникальный, nullable для admin/owner)
    
    -- Чувствительные персональные данные (будут шифроваться на уровне приложения)
    iin_encrypted TEXT NOT NULL,       -- Зашифрованный ИИН
    iin_hash VARCHAR(64) UNIQUE NOT NULL, -- SHA-256 хэш ИИН (для поиска/уникальности без дешифрования)
    full_name_encrypted TEXT NOT NULL,  -- Зашифрованное ФИО
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_users_iin_hash ON users(iin_hash);
CREATE INDEX IF NOT EXISTS idx_users_cid ON users(cid);

-- Таблица магазинов (принадлежат Владельцам)
CREATE TABLE IF NOT EXISTS shops (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    name VARCHAR(100) NOT NULL,
    address TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_shops_owner ON shops(owner_id);

-- Таблица долговых договоров (между Магазином и Покупателем)
CREATE TABLE IF NOT EXISTS agreements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_id UUID NOT NULL REFERENCES shops(id) ON DELETE RESTRICT,
    customer_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    credit_limit NUMERIC(15, 2) NOT NULL CHECK (credit_limit >= 0),
    balance NUMERIC(15, 2) NOT NULL DEFAULT 0.00, -- Текущий долг покупателя перед этим магазином
    due_date DATE NOT NULL,                      -- Дата погашения лимита
    status agreement_status NOT NULL DEFAULT 'pending_confirmation',
    
    -- Данные простой ЭЦП (SMS-подтверждения)
    signature_sms_id UUID, -- Будет связано с записью в sms_verifications после подтверждения
    signed_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_agreements_customer ON agreements(customer_id);
CREATE INDEX IF NOT EXISTS idx_agreements_status ON agreements(status);

-- Уникальный активный договор между конкретным магазином и покупателем.
-- Покупатель может иметь архивные (closed) договоры, но только один активный/ожидающий подтверждения в конкретном магазине.
CREATE UNIQUE INDEX IF NOT EXISTS idx_active_agreement_per_shop 
ON agreements(shop_id, customer_id) 
WHERE status IN ('pending_confirmation', 'active', 'suspended');

-- Таблица транзакций (Покупки в долг / Погашения)
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agreement_id UUID NOT NULL REFERENCES agreements(id) ON DELETE RESTRICT,
    type transaction_type NOT NULL,
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    receipt_image_url TEXT, -- Ссылка на фото чека (для покупок)
    status confirmation_status NOT NULL DEFAULT 'pending',
    
    -- Связь с SMS-подтверждением
    signature_sms_id UUID, -- Ссылка на запись верификации
    confirmed_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_transactions_agreement ON transactions(agreement_id);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);

-- Таблица SMS-верификаций (логирование кодов и фактов согласия)
CREATE TABLE IF NOT EXISTS sms_verifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(20) NOT NULL,
    code_hash VARCHAR(64) NOT NULL, -- Хэш одноразового кода (для безопасности)
    purpose VARCHAR(50) NOT NULL,   -- Назначение (sign_agreement, confirm_purchase, confirm_repayment, auth)
    reference_id UUID,              -- ID целевой сущности (agreements.id или transactions.id)
    status confirmation_status NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Метаданные для аудита простой ЭЦП
    ip_address VARCHAR(45),
    user_agent TEXT
);

CREATE INDEX IF NOT EXISTS idx_sms_phone ON sms_verifications(phone);
CREATE INDEX IF NOT EXISTS idx_sms_status_expires ON sms_verifications(status, expires_at);

-- +migrate Down
DROP TABLE IF EXISTS sms_verifications;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS agreements;
DROP TABLE IF EXISTS shops;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS confirmation_status;
DROP TYPE IF EXISTS transaction_type;
DROP TYPE IF EXISTS agreement_status;
DROP TYPE IF EXISTS user_role;
