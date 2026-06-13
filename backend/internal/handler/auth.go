package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"dapter-kz/internal/models"
)

type registerOwnerReq struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
	IIN      string `json:"iin"`
	FullName string `json:"full_name"`
}

type registerCustomerReq struct {
	Phone    string `json:"phone"`
	IIN      string `json:"iin"`
	FullName string `json:"full_name"`
}

type verifyRegisterReq struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

type setPinReq struct {
	Pin string `json:"pin"`
}

type loginReq struct {
	Phone    string `json:"phone"`
	Password string `json:"password"` // Также используется для PIN-кода
}

// RegisterOwner регистрирует Владельца магазина
func (h *Handler) RegisterOwner(w http.ResponseWriter, r *http.Request) {
	var req registerOwnerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	if req.Phone == "" || req.Password == "" || req.IIN == "" || req.FullName == "" {
		writeError(w, http.StatusBadRequest, "Все поля (phone, password, iin, full_name) обязательны")
		return
	}

	user, err := h.AuthService.RegisterOwner(r.Context(), req.Phone, req.Password, req.IIN, req.FullName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

// InitiateCustomerRegister инициирует регистрацию Покупателя (отправка OTP)
func (h *Handler) InitiateCustomerRegister(w http.ResponseWriter, r *http.Request) {
	var req registerCustomerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	if req.Phone == "" || req.IIN == "" || req.FullName == "" {
		writeError(w, http.StatusBadRequest, "Все поля (phone, iin, full_name) обязательны")
		return
	}

	verificationID, err := h.AuthService.InitiateCustomerRegister(r.Context(), req.Phone, req.IIN, req.FullName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message":         "Код подтверждения отправлен по SMS",
		"verification_id": verificationID,
	})
}

// VerifyCustomerRegister подтверждает регистрацию Покупателя и возвращает JWT токен
func (h *Handler) VerifyCustomerRegister(w http.ResponseWriter, r *http.Request) {
	var req verifyRegisterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	if req.Phone == "" || req.Code == "" {
		writeError(w, http.StatusBadRequest, "Поля phone и code обязательны")
		return
	}

	ip := r.RemoteAddr
	ua := r.UserAgent()

	token, err := h.AuthService.VerifyCustomerRegister(r.Context(), req.Phone, req.Code, ip, ua)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token":     token,
		"needs_pin": true, // Покупатель должен будет установить PIN-код
	})
}

// SetCustomerPin устанавливает 4-значный PIN-код покупателя
func (h *Handler) SetCustomerPin(w http.ResponseWriter, r *http.Request) {
	customerID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	var req setPinReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	if req.Pin == "" {
		writeError(w, http.StatusBadRequest, "Поле pin обязательно")
		return
	}

	err := h.AuthService.SetCustomerPin(r.Context(), customerID, req.Pin)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "PIN-код успешно установлен"})
}

// Login выполняет вход и возвращает JWT токен
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	if req.Phone == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "Поля phone и password/pin обязательны")
		return
	}

	token, err := h.AuthService.Login(r.Context(), req.Phone, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

// GetMe возвращает профиль текущего пользователя
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	// Кастим интерфейс AuthService и вызываем метод GetProfile
	type profileGetter interface {
		GetProfile(ctx context.Context, id string) (*models.User, error)
	}

	getter, ok := h.AuthService.(profileGetter)
	if !ok {
		writeError(w, http.StatusInternalServerError, "Сервис авторизации не реализует метод получения профиля")
		return
	}

	user, err := getter.GetProfile(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// GetOwners возвращает список всех владельцев (только для админа)
func (h *Handler) GetOwners(w http.ResponseWriter, r *http.Request) {
	owners, err := h.AuthService.GetUsersByRole(r.Context(), models.RoleOwner)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, owners)
}

// GetCustomers возвращает список всех покупателей (только для админа)
func (h *Handler) GetCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := h.AuthService.GetUsersByRole(r.Context(), models.RoleCustomer)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, customers)
}
