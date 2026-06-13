package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"dapter-kz/internal/models"
)

type createTransactionReq struct {
	AgreementID     string                 `json:"agreement_id"`
	Type            models.TransactionType `json:"type"` // "purchase" или "repayment"
	Amount          float64                `json:"amount"`
	ReceiptImageUrl *string                `json:"receipt_image_url,omitempty"`
}

type confirmTransactionReq struct {
	Code string `json:"code"`
}

// CreateTransaction создает новую транзакцию покупки или погашения (только для владельцев)
func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	var req createTransactionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	if req.AgreementID == "" || (req.Type != models.TypePurchase && req.Type != models.TypeRepayment) || req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "Поля agreement_id, type (purchase/repayment) и amount (>0) обязательны")
		return
	}

	transaction, err := h.TransactionService.CreateTransaction(r.Context(), ownerID, req.AgreementID, req.Type, req.Amount, req.ReceiptImageUrl)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, transaction)
}

// ConfirmTransaction подтверждает транзакцию по SMS-коду (только для покупателей)
func (h *Handler) ConfirmTransaction(w http.ResponseWriter, r *http.Request) {
	customerID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	transactionID := r.PathValue("id")
	if transactionID == "" {
		writeError(w, http.StatusBadRequest, "Не передан ID операции")
		return
	}

	var req confirmTransactionReq
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

	err := h.TransactionService.ConfirmTransaction(r.Context(), customerID, transactionID, req.Code, ip, ua)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Операция успешно подтверждена"})
}

// RejectTransaction отклоняет операцию покупки/погашения (только для покупателей)
func (h *Handler) RejectTransaction(w http.ResponseWriter, r *http.Request) {
	customerID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	transactionID := r.PathValue("id")
	if transactionID == "" {
		writeError(w, http.StatusBadRequest, "Не передан ID операции")
		return
	}

	err := h.TransactionService.RejectTransaction(r.Context(), customerID, transactionID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Операция отклонена"})
}

// GetTransactionsByAgreement возвращает историю транзакций по договору (с проверкой прав)
func (h *Handler) GetTransactionsByAgreement(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	agreementID := r.URL.Query().Get("agreement_id")
	if agreementID == "" {
		writeError(w, http.StatusBadRequest, "Не передан параметр запроса agreement_id")
		return
	}

	transactions, err := h.TransactionService.GetTransactionsByAgreement(r.Context(), userID, agreementID)
	if err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, transactions)
}

// UploadReceipt загружает файл чека на сервер (только для владельцев)
func (h *Handler) UploadReceipt(w http.ResponseWriter, r *http.Request) {
	// Максимальный размер файла — 5 МБ
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "Файл слишком большой. Максимальный размер: 5 МБ")
		return
	}

	file, fileHeader, err := r.FormFile("receipt")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Файл не найден (ожидается multipart-поле 'receipt')")
		return
	}
	defer file.Close()

	// Создаем папку uploads, если её нет
	uploadsDir := "./uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "Не удалось создать папку загрузок на сервере")
		return
	}

	// Генерируем уникальное имя файла
	ext := filepath.Ext(fileHeader.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadsDir, filename)

	out, err := os.Create(filePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка создания файла на сервере")
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка копирования файла")
		return
	}

	// Возвращаем путь, по которому можно получить чек
	url := fmt.Sprintf("/uploads/%s", filename)
	writeJSON(w, http.StatusOK, map[string]string{
		"url": url,
	})
}
