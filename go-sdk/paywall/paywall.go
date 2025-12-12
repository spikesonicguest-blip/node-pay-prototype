package paywall

import (
	"encoding/json"
	"net/http"
	
	"nodepay-go-sdk/client"
	"nodepay-go-sdk/models"
)

type Paywall struct {
	Client *client.Client
}

func New(c *client.Client) *Paywall {
	return &Paywall{Client: c}
}

// PaymentConfig defines the cost of a resource
type PaymentConfig struct {
	Amount      string
	Currency    string
	Network     string
	Asset       string
	PayTo       string
	Description string
}

// RequirePayment is a middleware that enforces x402 payment
func (p *Paywall) Handler(cfg PaymentConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			
			requirements := models.PaymentRequirements{
				Scheme:            "exact",
				Network:           cfg.Network,
				Amount:            cfg.Amount,
				Asset:             cfg.Asset,
				PayTo:             cfg.PayTo,
				MaxTimeoutSeconds: 60,
				Extra: map[string]interface{}{
					"name":    cfg.Currency,
					"version": "2",
				},
			}

			// 1. Check for Payment (Payment-Signature)
			paymentSig := r.Header.Get("Payment-Signature")
			if paymentSig != "" {
				// Verify via Client
				valid, err := p.Client.Verify(paymentSig, requirements)
				if err == nil && valid {
					next.ServeHTTP(w, r)
					return
				}
			}

			// 2. Return 402 Payment Required
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPaymentRequired)

			desc := cfg.Description
			if desc == "" {
				desc = "Premium Resource"
			}

			// Construct PaymentRequired response (Wire Protocol)
			reqs := models.PaymentRequired{
				X402Version: 2,
				Resource: models.ResourceInfo{
					URL:         r.URL.String(),
					Description: desc,
					MimeType:    "application/json",
				},
				Accepts: []models.PaymentRequirements{requirements},
			}
			
			json.NewEncoder(w).Encode(reqs)
		})
	}
}
