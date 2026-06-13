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

type userRepo struct {
	db *pgxpool.Pool
}

// NewUserRepository создает новый экземпляр репозитория пользователей
func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (phone, password_hash, pin_code_hash, role, cid, iin_encrypted, iin_hash, full_name_encrypted, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		user.Phone,
		user.PasswordHash,
		user.PinCodeHash,
		user.Role,
		user.CID,
		user.IINEncrypted,
		user.IINHash,
		user.FullNameEncrypted,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("ошибка при вставке пользователя в БД: %w", err)
	}
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, phone, password_hash, pin_code_hash, role, cid, iin_encrypted, iin_hash, full_name_encrypted, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var user models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Phone,
		&user.PasswordHash,
		&user.PinCodeHash,
		&user.Role,
		&user.CID,
		&user.IINEncrypted,
		&user.IINHash,
		&user.FullNameEncrypted,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("пользователь с ID %s не найден", id)
		}
		return nil, fmt.Errorf("ошибка получения пользователя по ID: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	query := `
		SELECT id, phone, password_hash, pin_code_hash, role, cid, iin_encrypted, iin_hash, full_name_encrypted, created_at, updated_at
		FROM users
		WHERE phone = $1
	`
	var user models.User
	err := r.db.QueryRow(ctx, query, phone).Scan(
		&user.ID,
		&user.Phone,
		&user.PasswordHash,
		&user.PinCodeHash,
		&user.Role,
		&user.CID,
		&user.IINEncrypted,
		&user.IINHash,
		&user.FullNameEncrypted,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Возвращаем nil, nil если не найден — удобно для валидации при регистрации
		}
		return nil, fmt.Errorf("ошибка получения пользователя по телефону: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetByIINHash(ctx context.Context, iinHash string) (*models.User, error) {
	query := `
		SELECT id, phone, password_hash, pin_code_hash, role, cid, iin_encrypted, iin_hash, full_name_encrypted, created_at, updated_at
		FROM users
		WHERE iin_hash = $1
	`
	var user models.User
	err := r.db.QueryRow(ctx, query, iinHash).Scan(
		&user.ID,
		&user.Phone,
		&user.PasswordHash,
		&user.PinCodeHash,
		&user.Role,
		&user.CID,
		&user.IINEncrypted,
		&user.IINHash,
		&user.FullNameEncrypted,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Удобно для проверки уникальности
		}
		return nil, fmt.Errorf("ошибка получения пользователя по хэшу ИИН: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetByCID(ctx context.Context, cid string) (*models.User, error) {
	query := `
		SELECT id, phone, password_hash, pin_code_hash, role, cid, iin_encrypted, iin_hash, full_name_encrypted, created_at, updated_at
		FROM users
		WHERE cid = $1
	`
	var user models.User
	err := r.db.QueryRow(ctx, query, cid).Scan(
		&user.ID,
		&user.Phone,
		&user.PasswordHash,
		&user.PinCodeHash,
		&user.Role,
		&user.CID,
		&user.IINEncrypted,
		&user.IINHash,
		&user.FullNameEncrypted,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка получения пользователя по CID: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetAllByRole(ctx context.Context, role models.UserRole) ([]*models.User, error) {
	query := `
		SELECT id, phone, password_hash, pin_code_hash, role, cid, iin_encrypted, iin_hash, full_name_encrypted, created_at, updated_at
		FROM users
		WHERE role = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, role)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка пользователей по роли: %w", err)
	}
	defer rows.Close()

	users := make([]*models.User, 0)
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Phone,
			&user.PasswordHash,
			&user.PinCodeHash,
			&user.Role,
			&user.CID,
			&user.IINEncrypted,
			&user.IINHash,
			&user.FullNameEncrypted,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования записи пользователя: %w", err)
		}
		users = append(users, &user)
	}
	return users, nil
}

func (r *userRepo) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET phone = $2, password_hash = $3, pin_code_hash = $4, role = $5, cid = $6, 
		    iin_encrypted = $7, iin_hash = $8, full_name_encrypted = $9, updated_at = NOW()
		WHERE id = $1
	`
	result, err := r.db.Exec(ctx, query,
		user.ID,
		user.Phone,
		user.PasswordHash,
		user.PinCodeHash,
		user.Role,
		user.CID,
		user.IINEncrypted,
		user.IINHash,
		user.FullNameEncrypted,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления пользователя: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("пользователь для обновления не найден")
	}
	return nil
}
