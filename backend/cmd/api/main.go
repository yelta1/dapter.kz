package main

import (
	"context"
	"log"
	"net/http"

	"dapter-kz/internal/config"
	"dapter-kz/internal/handler"
	"dapter-kz/internal/pkg/database"
	"dapter-kz/internal/repository/postgres"
	"dapter-kz/internal/service"
)

func main() {
	log.Println("Инициализация приложения...")

	// 1. Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// 2. Подключаемся к базе данных
	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()
	log.Println("Успешное подключение к PostgreSQL")

	// 3. Инициализируем репозитории
	userRepo := postgres.NewUserRepository(db)
	shopRepo := postgres.NewShopRepository(db)
	agreementRepo := postgres.NewAgreementRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)
	smsRepo := postgres.NewSmsRepository(db)

	// 4. Инициализируем сервисы
	// В качестве соли для хэширования SMS используем часть JWTSecret или фиксированную соль
	smsSalt := cfg.JWTSecret
	smsService := service.NewSmsService(
		smsRepo, 
		smsSalt, 
		cfg.GreenAPIIDInstance, 
		cfg.GreenAPITokenInstance, 
		cfg.GreenAPIURL,
	)
	
	authService := service.NewAuthService(userRepo, smsService, cfg)
	
	// Выполняем посев (seeding) суперадминистратора по умолчанию
	if err := authService.SeedAdmin(context.Background()); err != nil {
		log.Fatalf("Ошибка посева суперадминистратора: %v", err)
	}
	
	shopService := service.NewShopService(shopRepo)
	
	agreementService := service.NewAgreementService(
		agreementRepo, 
		shopRepo, 
		userRepo, 
		smsService, 
		authService,
	)
	
	transactionService := service.NewTransactionService(
		transactionRepo, 
		agreementRepo, 
		shopRepo, 
		userRepo, 
		smsService,
	)

	// 5. Инициализируем обработчики (handlers) и роутер
	h := handler.NewHandler(authService, shopService, agreementService, transactionService)
	mux := http.NewServeMux()
	
	// Регистрируем маршруты в роутере
	h.RegisterRoutes(mux, cfg)

	// 6. Запускаем веб-сервер
	addr := ":" + cfg.Port
	log.Printf("Сервер успешно запущен в окружении '%s' на порту %s", cfg.Env, cfg.Port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Ошибка запуска веб-сервера: %v", err)
	}
}
