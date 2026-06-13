package handler

import (
	"encoding/json"
	"net/http"

	"dapter-kz/internal/config"
	"dapter-kz/internal/models"
	"dapter-kz/internal/service"
)

// Handler агрегирует все обработчики запросов
type Handler struct {
	AuthService        service.AuthService
	ShopService        service.ShopService
	AgreementService   service.AgreementService
	TransactionService service.TransactionService
}

// NewHandler создает новый экземпляр главного обработчика
func NewHandler(
	auth service.AuthService,
	shop service.ShopService,
	agreement service.AgreementService,
	transaction service.TransactionService,
) *Handler {
	return &Handler{
		AuthService:        auth,
		ShopService:        shop,
		AgreementService:   agreement,
		TransactionService: transaction,
	}
}

// RegisterRoutes регистрирует все HTTP-эндпоинты
func (h *Handler) RegisterRoutes(mux *http.ServeMux, cfg *config.Config) {
	// Инициализируем middleware
	authMW := AuthMiddleware(cfg)
	ownerOnlyMW := RequireRole(models.RoleOwner)
	customerOnlyMW := RequireRole(models.RoleCustomer)
	adminOnlyMW := RequireRole(models.RoleAdmin)

	// --- Публичные эндпоинты аутентификации ---
	mux.HandleFunc("POST /api/v1/auth/register-customer", h.InitiateCustomerRegister)
	mux.HandleFunc("POST /api/v1/auth/verify-registration", h.VerifyCustomerRegister)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)

	// --- Защищенные эндпоинты ---
	// Административные
	mux.Handle("POST /api/v1/auth/register-owner", authMW(adminOnlyMW(http.HandlerFunc(h.RegisterOwner))))
	mux.Handle("GET /api/v1/admin/owners", authMW(adminOnlyMW(http.HandlerFunc(h.GetOwners))))
	mux.Handle("GET /api/v1/admin/customers", authMW(adminOnlyMW(http.HandlerFunc(h.GetCustomers))))
	mux.Handle("GET /api/v1/admin/shops", authMW(adminOnlyMW(http.HandlerFunc(h.GetAllShops))))

	// Профиль и PIN-код
	mux.Handle("POST /api/v1/auth/set-pin", authMW(customerOnlyMW(http.HandlerFunc(h.SetCustomerPin))))
	mux.Handle("GET /api/v1/auth/me", authMW(http.HandlerFunc(h.GetMe)))

	// Магазины
	mux.Handle("POST /api/v1/shops", authMW(adminOnlyMW(http.HandlerFunc(h.CreateShop))))
	mux.Handle("GET /api/v1/shops", authMW(ownerOnlyMW(http.HandlerFunc(h.GetShops))))

	// Договоры
	mux.Handle("POST /api/v1/agreements", authMW(ownerOnlyMW(http.HandlerFunc(h.CreateAgreement))))
	mux.Handle("POST /api/v1/agreements/{id}/confirm", authMW(customerOnlyMW(http.HandlerFunc(h.ConfirmAgreement))))
	mux.Handle("GET /api/v1/agreements", authMW(http.HandlerFunc(h.GetAgreements)))
	mux.Handle("GET /api/v1/agreements/{id}", authMW(http.HandlerFunc(h.GetAgreementByID)))
	mux.Handle("GET /api/v1/agreements/active", authMW(ownerOnlyMW(http.HandlerFunc(h.GetActiveAgreementByCID))))

	// Транзакции
	mux.Handle("POST /api/v1/transactions", authMW(ownerOnlyMW(http.HandlerFunc(h.CreateTransaction))))
	mux.Handle("POST /api/v1/transactions/{id}/confirm", authMW(customerOnlyMW(http.HandlerFunc(h.ConfirmTransaction))))
	mux.Handle("POST /api/v1/transactions/{id}/reject", authMW(customerOnlyMW(http.HandlerFunc(h.RejectTransaction))))
	mux.Handle("GET /api/v1/transactions", authMW(http.HandlerFunc(h.GetTransactionsByAgreement)))

	// Загрузка чеков и раздача статики
	mux.Handle("POST /api/v1/upload", authMW(ownerOnlyMW(http.HandlerFunc(h.UploadReceipt))))
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
}

// Вспомогательные функции отправки JSON-ответов
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
