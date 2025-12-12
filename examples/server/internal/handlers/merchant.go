package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"fmt"
	"math/rand"
	
	"nodepay-example-merchant/internal/models"
	"nodepay-example-merchant/internal/store"
)

type Handler struct {
	Store *store.Store
}

func New(s *store.Store) *Handler {
	return &Handler{Store: s}
}

func (h *Handler) CreateCharge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateChargeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	charge := models.Charge{
		ID:              fmt.Sprintf("ch_%d", rand.Int63()),
		Status:          "new",
		PricingCurrency: req.PricingCurrency,
		AmountPricing:   req.AmountPricing,
		Asset:           req.Asset,
		Network:         req.Network,
		Amount:          "1000000", // Mocked atomic amount
		CreatedAt:       time.Now(),
		ExpiresAt:       time.Now().Add(1 * time.Hour),
		PaymentURL:      "https://pay.nodepay.ai/checkout",
		PayTo:           "0xMerchantAddress",
	}

	if err := h.Store.CreateCharge(charge); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // CORS for demo
	json.NewEncoder(w).Encode(charge)
}

func (h *Handler) GetCharge(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

    // Simplistic extraction for demo
    id := r.URL.Path // e.g. "ch_123" if stripped

	charge, err := h.Store.GetCharge(id)
	if err == store.ErrNotFound {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(charge)
}

func (h *Handler) Product(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"data": "You have purchased the Digital Widget! Thank you for using NodePay."}`))
}
