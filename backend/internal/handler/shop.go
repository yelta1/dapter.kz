package handler

import (
	"encoding/json"
	"net/http"
)

type createShopReq struct {
	OwnerID string `json:"owner_id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

// CreateShop создает новый магазин (только для администраторов)
func (h *Handler) CreateShop(w http.ResponseWriter, r *http.Request) {
	var req createShopReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	if req.Name == "" || req.OwnerID == "" {
		writeError(w, http.StatusBadRequest, "Поля owner_id и name обязательны")
		return
	}

	shop, err := h.ShopService.CreateShop(r.Context(), req.OwnerID, req.Name, req.Address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, shop)
}

// GetShops возвращает список магазинов текущего владельца
func (h *Handler) GetShops(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	shops, err := h.ShopService.GetShopsByOwner(r.Context(), ownerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, shops)
}

// GetAllShops возвращает список вообще всех магазинов в системе (только для админа)
func (h *Handler) GetAllShops(w http.ResponseWriter, r *http.Request) {
	shops, err := h.ShopService.GetAllShops(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, shops)
}
