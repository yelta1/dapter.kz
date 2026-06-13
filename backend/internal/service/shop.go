package service

import (
	"context"
	"dapter-kz/internal/models"
	"dapter-kz/internal/repository"
)

// ShopService описывает логику управления магазинами
type ShopService interface {
	CreateShop(ctx context.Context, ownerID string, name, address string) (*models.Shop, error)
	GetShopByID(ctx context.Context, id string) (*models.Shop, error)
	GetShopsByOwner(ctx context.Context, ownerID string) ([]*models.Shop, error)
	GetAllShops(ctx context.Context) ([]*models.Shop, error)
}

type shopService struct {
	shopRepo repository.ShopRepository
}

// NewShopService создает новый экземпляр сервиса магазинов
func NewShopService(shopRepo repository.ShopRepository) ShopService {
	return &shopService{shopRepo: shopRepo}
}

func (s *shopService) CreateShop(ctx context.Context, ownerID string, name, address string) (*models.Shop, error) {
	shop := &models.Shop{
		OwnerID: ownerID,
		Name:    name,
		Address: address,
	}
	err := s.shopRepo.Create(ctx, shop)
	if err != nil {
		return nil, err
	}
	return shop, nil
}

func (s *shopService) GetShopByID(ctx context.Context, id string) (*models.Shop, error) {
	return s.shopRepo.GetByID(ctx, id)
}

func (s *shopService) GetShopsByOwner(ctx context.Context, ownerID string) ([]*models.Shop, error) {
	return s.shopRepo.GetByOwnerID(ctx, ownerID)
}

func (s *shopService) GetAllShops(ctx context.Context) ([]*models.Shop, error) {
	return s.shopRepo.GetAll(ctx)
}
