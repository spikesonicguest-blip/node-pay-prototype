package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	
	"nodepay-go-sdk/models"
)

type Client struct {
	FacilitatorURL string
	HTTPClient     *http.Client
}

func New(facilitatorURL string) *Client {
	return &Client{
		FacilitatorURL: facilitatorURL,
		HTTPClient:     &http.Client{},
	}
}

// Verify calls the Facilitator's /v1/verify endpoint
func (c *Client) Verify(signature string, requirements models.PaymentRequirements) (bool, error) {
	reqBody := models.VerifyRequest{
		PaymentPayload: models.PaymentPayload{
			X402Version: 2,
			Payload: map[string]interface{}{
				"signature": signature,
			},
		},
		PaymentRequirements: requirements,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return false, err
	}

	resp, err := c.HTTPClient.Post(c.FacilitatorURL+"/v1/verify", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("facilitator returned %d", resp.StatusCode)
	}

	var verifyResp models.VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
		return false, err
	}

	return verifyResp.IsValid, nil
}
