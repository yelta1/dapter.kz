package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"dapter-kz/internal/models"
)

type createAgreementReq struct {
	ShopID        string  `json:"shop_id"`
	CustomerPhone string  `json:"customer_phone"`
	CreditLimit   float64 `json:"credit_limit"`
	DueDate       string  `json:"due_date"` // Ожидается в формате "2006-01-02"
}

type confirmAgreementReq struct {
	Code string `json:"code"`
}

// CreateAgreement создает новый договор (только для владельцев)
func (h *Handler) CreateAgreement(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	var req createAgreementReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	if req.ShopID == "" || req.CustomerPhone == "" || req.CreditLimit <= 0 || req.DueDate == "" {
		writeError(w, http.StatusBadRequest, "Поля shop_id, customer_phone, credit_limit (>0) и due_date обязательны")
		return
	}

	// Парсим дату погашения
	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат даты due_date (ожидается ГГГГ-ММ-ДД, например: 2026-12-31)")
		return
	}

	agreement, err := h.AgreementService.CreateAgreement(r.Context(), ownerID, req.ShopID, req.CustomerPhone, req.CreditLimit, dueDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, agreement)
}

// ConfirmAgreement подтверждает и активирует договор по SMS-коду (только для покупателей)
func (h *Handler) ConfirmAgreement(w http.ResponseWriter, r *http.Request) {
	customerID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	agreementID := r.PathValue("id")
	if agreementID == "" {
		writeError(w, http.StatusBadRequest, "Не передан ID договора")
		return
	}

	var req confirmAgreementReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "Поле code обязательно")
		return
	}

	ip := r.RemoteAddr
	ua := r.UserAgent()

	err := h.AgreementService.ConfirmAgreement(r.Context(), customerID, agreementID, req.Code, ip, ua)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Договор успешно подписан и активирован"})
}

// GetAgreements возвращает список договоров пользователя (для В — по его магазинам, для П — личные)
func (h *Handler) GetAgreements(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	role, okRole := GetUserRoleFromContext(r.Context())
	if !ok || !okRole {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	agreements, err := h.AgreementService.GetAgreementsForUser(r.Context(), userID, role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, agreements)
}

// GetAgreementByID возвращает детальную информацию о конкретном договоре (с проверкой прав)
func (h *Handler) GetAgreementByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	agreementID := r.PathValue("id")
	if agreementID == "" {
		writeError(w, http.StatusBadRequest, "Не передан ID договора")
		return
	}

	agreement, err := h.AgreementService.GetAgreementByID(r.Context(), userID, agreementID)
	if err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, agreement)
}

// GetActiveAgreementByCID возвращает активный договор по CID покупателя и ID магазина (только для владельцев)
func (h *Handler) GetActiveAgreementByCID(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	cid := r.URL.Query().Get("cid")
	shopID := r.URL.Query().Get("shop_id")

	if cid == "" || shopID == "" {
		writeError(w, http.StatusBadRequest, "Параметры запроса cid и shop_id обязательны")
		return
	}

	// 1. Проверяем, что магазин принадлежит данному владельцу
	shop, err := h.ShopService.GetShopByID(r.Context(), shopID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Магазин не найден")
		return
	}
	if shop.OwnerID != ownerID {
		writeError(w, http.StatusForbidden, "Доступ запрещен: этот магазин не принадлежит вам")
		return
	}

	// 2. Ищем покупателя по CID
	customer, err := h.AuthService.GetUserByCID(r.Context(), cid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if customer == nil {
		writeError(w, http.StatusNotFound, "Покупатель с таким ID не найден")
		return
	}

	// 3. Ищем активный договор по паре магазин-покупатель
	agreements, err := h.AgreementService.GetAgreementsForUser(r.Context(), customer.ID, models.RoleCustomer)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var activeAgreement *models.Agreement
	for _, a := range agreements {
		if a.ShopID == shopID && (a.Status == models.StatusActive || a.Status == models.StatusPendingConfirmation) {
			activeAgreement = a
			break
		}
	}

	if activeAgreement == nil {
		writeError(w, http.StatusNotFound, "Активный договор не найден для этого покупателя в данном магазине")
		return
	}

	writeJSON(w, http.StatusOK, activeAgreement)
}
