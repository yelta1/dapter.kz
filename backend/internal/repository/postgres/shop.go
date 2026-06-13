package postgres

import (
	"context"
	"errors"
	"fmt"

	"dapter-kz/internal/models"
	"dapter-kz/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type shopRepo struct {
	db *pgxpool.Pool
}

// NewShopRepository создает новый экземпляр репозитория магазинов
func NewShopRepository(db *pgxpool.Pool) repository.ShopRepository {
	return &shopRepo{db: db}
}

func (r *shopRepo) Create(ctx context.Context, shop *models.Shop) error {
	query := `
		INSERT INTO shops (owner_id, name, address, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query, shop.OwnerID, shop.Name, shop.Address).Scan(&shop.ID, &shop.CreatedAt, &shop.UpdatedAt)
	if err != nil {
		return fmt.Errorf("ошибка создания магазина в БД: %w", err)
	}
	return nil
}

func (r *shopRepo) GetByID(ctx context.Context, id string) (*models.Shop, error) {
	query := `
		SELECT id, owner_id, name, address, created_at, updated_at
		FROM shops
		WHERE id = $1
	`
	var shop models.Shop
	err := r.db.QueryRow(ctx, query, id).Scan(
		&shop.ID,
		&shop.OwnerID,
		&shop.Name,
		&shop.Address,
		&shop.CreatedAt,
		&shop.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("магазин с ID %s не найден", id)
		}
		return nil, fmt.Errorf("ошибка получения магазина по ID: %w", err)
	}
	return &shop, nil
}

func (r *shopRepo) GetByOwnerID(ctx context.Context, ownerID string) ([]*models.Shop, error) {
	query := `
		SELECT id, owner_id, name, address, created_at, updated_at
		FROM shops
		WHERE owner_id = $1
		ORDER BY name
	`
	rows, err := r.db.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса получения магазинов владельца: %w", err)
	}
	defer rows.Close()

	shops := make([]*models.Shop, 0)
	for rows.Next() {
		var shop models.Shop
		err := rows.Scan(
			&shop.ID,
			&shop.OwnerID,
			&shop.Name,
			&shop.Address,
			&shop.CreatedAt,
			&shop.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования записи магазина: %w", err)
		}
		shops = append(shops, &shop)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после итерации по магазинам: %w", err)
	}
	return shops, nil
}

func (r *shopRepo) GetAll(ctx context.Context) ([]*models.Shop, error) {
	query := `
		SELECT id, owner_id, name, address, created_at, updated_at
		FROM shops
		ORDER BY name
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса получения всех магазинов: %w", err)
	}
	defer rows.Close()

	shops := make([]*models.Shop, 0)
	for rows.Next() {
		var shop models.Shop
		err := rows.Scan(
			&shop.ID,
			&shop.OwnerID,
			&shop.Name,
			&shop.Address,
			&shop.CreatedAt,
			&shop.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования записи магазина: %w", err)
		}
		shops = append(shops, &shop)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после итерации по всем магазинам: %w", err)
	}
	return shops, nil
}
