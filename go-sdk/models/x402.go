package models

// x402 V2 Wire Protocol Models

// PaymentRequired matches spec/x402_protocol.yaml PaymentRequired
type PaymentRequired struct {
	X402Version int                   `json:"x402Version"`
	Error       string                `json:"error,omitempty"`
	Resource    ResourceInfo          `json:"resource"`
	Accepts     []PaymentRequirements `json:"accepts"`
	Extensions  map[string]interface{} `json:"extensions,omitempty"`
}

// PaymentPayload matches spec/x402_protocol.yaml PaymentPayload
type PaymentPayload struct {
	X402Version int                    `json:"x402Version"`
	Resource    *ResourceInfo          `json:"resource,omitempty"`
	Accepted    PaymentRequirements    `json:"accepted"`
	Payload     map[string]interface{} `json:"payload"` // signature, authorization
	Extensions  map[string]interface{} `json:"extensions,omitempty"`
}

type ResourceInfo struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

type PaymentRequirements struct {
	Scheme            string                 `json:"scheme"`
	Network           string                 `json:"network"` // CAIP-2
	Amount            string                 `json:"amount"`
	Asset             string                 `json:"asset"`
	PayTo             string                 `json:"payTo"`
	MaxTimeoutSeconds int                    `json:"maxTimeoutSeconds"`
	Extra             map[string]interface{} `json:"extra,omitempty"`
}

// Facilitator /verify and /settle Request
type VerifyRequest struct {
	PaymentPayload      PaymentPayload      `json:"paymentPayload"`
	PaymentRequirements PaymentRequirements `json:"paymentRequirements"`
}

type VerifyResponse struct {
	IsValid       bool   `json:"isValid"`
	InvalidReason string `json:"invalidReason,omitempty"`
	Payer         string `json:"payer,omitempty"`
}

type SettleResponse struct {
	Success     bool   `json:"success"`
	ErrorReason string `json:"errorReason,omitempty"`
	Payer       string `json:"payer,omitempty"`
	Transaction string `json:"transaction"`
	Network     string `json:"network"`
}
