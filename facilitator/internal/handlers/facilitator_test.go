package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	
	"math/big"
	"nodepay-facilitator/internal/models"
	
	
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func toBig(s string) *big.Int {
	n, _ := new(big.Int).SetString(s, 10)
	return n
}

func TestVerify(t *testing.T) {
	h := New()

	// Case 1: Valid structure (Signature will fail, but handler should not crash)
	reqBody := models.VerifyRequest{
		PaymentPayload: models.PaymentPayload{
			X402Version: 2,
			Payload: map[string]interface{}{
				"from":        "0x0000000000000000000000000000000000000001",
				"to":          "0x0000000000000000000000000000000000000002",
				"value":       "0x1",
				"validAfter":  "0x0",
				"validBefore": "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
				"nonce":       "0x0000000000000000000000000000000000000000000000000000000000000001",
				"v":           28,
				"r":           "0x0000000000000000000000000000000000000000000000000000000000000001",
				"s":           "0x0000000000000000000000000000000000000000000000000000000000000001",
			},
		},
		PaymentRequirements: models.PaymentRequirements{
			Amount: "1",
			PayTo:  "0x0000000000000000000000000000000000000002",
		},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/v1/verify", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	h.Verify(rr, req)

	// The handler returns 200 OK even for invalid signatures, just with IsValid=false
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %v", rr.Code)
	}
	// We expect false because signature is fake
	var resp models.VerifyResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.IsValid {
		t.Errorf("expected invalid (signature not real), got valid")
	}

	// Case 2: Invalid (missing version)
	reqBodyVal := models.VerifyRequest{
		PaymentPayload: models.PaymentPayload{
			X402Version: 1, // Wrong version
			Payload: map[string]interface{}{},
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
	h := New()

	reqBody := models.VerifyRequest{
		PaymentPayload: models.PaymentPayload{
			X402Version: 2,
			Payload: map[string]interface{}{
				"from":        "0x0000000000000000000000000000000000000001",
				"to":          "0x0000000000000000000000000000000000000002",
				"value":       hexutil.EncodeBig(toBig("1")),
				"validAfter":  "0x0",
				"validBefore": "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
				"nonce":       hexutil.Encode([]byte{1, 2, 3}),
				"v":           28,
				"r":           hexutil.Encode([]byte{1}),
				"s":           hexutil.Encode([]byte{1}),
			},
		},
		PaymentRequirements: models.PaymentRequirements{Network: "base"},
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/v1/settle", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	h.Settle(rr, req)

	// Previously we expected 200/Success. Now that we removed mock mode,
	// and we have no real key/network, we expect 500.
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 (strict mode), got %v", rr.Code)
	}
}

func TestSupported(t *testing.T) {
	h := New()

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
