package models

import "time"

// CreateChargeRequest matches spec/charges.yaml
type CreateChargeRequest struct {
	PricingCurrency  string                 `json:"pricing_currency"`
	AmountPricing    string                 `json:"amount_pricing"`
	Asset            string                 `json:"asset"`   // Address or ISO
	Network          string                 `json:"network"` // CAIP-2
	SettlementRuleID string                 `json:"settlement_rule_id,omitempty"`
	Description      string                 `json:"description,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// Charge matches spec/charges.yaml and facilitator.yaml
type Charge struct {
	ID               string                 `json:"id"`
	Status           string                 `json:"status"` // new, pending, paid, expired, failed
	PricingCurrency  string                 `json:"pricing_currency"`
	AmountPricing    string                 `json:"amount_pricing"`
	Asset            string                 `json:"asset"`
	Network          string                 `json:"network"`
	Amount           string                 `json:"amount"` // Payment amount in atomic units
	SettlementAsset  string                 `json:"settlement_asset,omitempty"`
	SettlementAmount string                 `json:"settlement_amount,omitempty"`
	PaymentURL       string                 `json:"payment_url,omitempty"`
	PayTo            string                 `json:"pay_to,omitempty"`
	Beneficiary      *Beneficiary           `json:"beneficiary,omitempty"`
	ExpiresAt        time.Time              `json:"expires_at"`
	CreatedAt        time.Time              `json:"created_at"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type Beneficiary struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// CreateRefundRequest matches spec/charges.yaml
type CreateRefundRequest struct {
	ChargeID string                 `json:"charge_id"`
	Amount   string                 `json:"amount"`
	Reason   string                 `json:"reason,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Refund matches spec/charges.yaml
type Refund struct {
	ID        string                 `json:"id"`
	ChargeID  string                 `json:"charge_id"`
	Status    string                 `json:"status"` // pending, succeeded, failed
	Amount    string                 `json:"amount"`
	Reason    string                 `json:"reason,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
