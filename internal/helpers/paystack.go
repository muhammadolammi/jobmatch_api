package helpers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type PaystackResponse struct {
	Status  bool            `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"` // remains flexible for any endpoint
}

func CallPaystack(url, method, secret string, payload any, client *http.Client) (*PaystackResponse, error) {
	payloadb, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling payload: %v", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(payloadb))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+secret)
	req.Header.Add("Content-Type", "application/json")

	httpRes, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client.Do error: %v", err)
	}
	defer httpRes.Body.Close()

	body, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %v", err)
	}

	// Accept **all 2xx success codes**
	if httpRes.StatusCode < 200 || httpRes.StatusCode >= 300 {
		return nil, fmt.Errorf("http error status: %v, body: %s", httpRes.StatusCode, string(body))
	}

	ps := PaystackResponse{}
	if err := json.Unmarshal(body, &ps); err != nil {
		return nil, fmt.Errorf("error unmarshalling Paystack response: %v", err)
	}

	// Paystack success = status == true
	if !ps.Status {
		return nil, fmt.Errorf("paystack error: %s", ps.Message)
	}

	return &ps, nil
}

func VerifyPaystackSignature(payload []byte, signature string, secret string) bool {
	h := hmac.New(sha512.New, []byte(secret))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	return expectedSignature == signature
}
