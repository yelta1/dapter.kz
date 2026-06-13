package database

import (
	"context"
	"fmt"
	"time"

	"dapter-kz/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectPostgres инициализирует пул соединений с базой данных PostgreSQL
func ConnectPostgres(cfg *config.Config) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSslMode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("не удалось распарсить строку подключения: %w", err)
	}

	// Настройка параметров пула
	poolConfig.MaxConns = 15
	poolConfig.MinConns = 2
	poolConfig.MaxConnIdleTime = 15 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания пула соединений: %w", err)
	}

	// Проверяем соединение с БД
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("база данных недоступна (ошибка Ping): %w", err)
	}

	return pool, nil
}
