package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	
	"nodepay-facilitator/internal/models"
	"nodepay-facilitator/internal/store"
)

func TestVerify(t *testing.T) {
	h := New(store.New())

	// Case 1: Valid
	reqBody := models.VerifyRequest{
		PaymentPayload: models.PaymentPayload{
			Payload: map[string]interface{}{"signature": "0xValidsig"},
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/v1/verify", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	h.Verify(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %v", rr.Code)
	}
	var resp models.VerifyResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if !resp.IsValid {
		t.Errorf("expected valid, got invalid")
	}

	// Case 2: Invalid (missing signature)
	reqBodyVal := models.VerifyRequest{
		PaymentPayload: models.PaymentPayload{
			Payload: map[string]interface{}{"signature": ""},
		},
	}
	bodyVal, _ := json.Marshal(reqBodyVal)
	reqVal, _ := http.NewRequest("POST", "/v1/verify", bytes.NewBuffer(bodyVal))
	rrVal := httptest.NewRecorder()
	h.Verify(rrVal, reqVal)

	var respVal models.VerifyResponse
	json.NewDecoder(rrVal.Body).Decode(&respVal)
	if respVal.IsValid {
		t.Errorf("expected invalid, got valid")
	}
}

func TestSettle(t *testing.T) {
	h := New(store.New())

	reqBody := models.VerifyRequest{
		PaymentRequirements: models.PaymentRequirements{Network: "base"},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/v1/settle", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	h.Settle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %v", rr.Code)
	}
	var resp models.SettleResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if !resp.Success {
		t.Errorf("expected success")
	}
	if resp.Transaction == "" {
		t.Error("expected transaction hash")
	}
}

func TestSupported(t *testing.T) {
	h := New(store.New())

	req, _ := http.NewRequest("GET", "/v1/supported", nil)
	rr := httptest.NewRecorder()
	h.Supported(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %v", rr.Code)
	}
	var resp models.SupportedResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	
	if len(resp.Kinds) == 0 {
		t.Error("expected supported kinds")
	}
}
