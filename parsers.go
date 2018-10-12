package goflick

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// loginFailed returns true if login fails
func loginFailed(response []byte) bool {
	return strings.Contains(string(response), "Auth failed invalid_grant")
}

// parseAuth parses the response and returns the FlickConnect struct properly
// populated from the output
func parseAuth(response []byte) (*FlickConnect, error) {
	var fc = FlickConnect{}
	if err := json.Unmarshal(response, &fc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %s", err.Error())
	}
	if fc.Token == "" || fc.TokenType != "bearer" {
		return nil, fmt.Errorf("failed to get the token")
	}
	return &fc, nil
}

// parsePrice parses the response extracting the current needle price
func parsePrice(response []byte) (float64, error) {
	var priceJSON map[string]interface{}
	if err := json.Unmarshal(response, &priceJSON); err != nil {
		return 0.0, err
	}
	// we need to go deeper!
	needle, ok := priceJSON["needle"].(map[string]interface{})
	if !ok {
		return 0.0, fmt.Errorf("failed to parse the response: %s", priceJSON)
	}
	// we need to go deeper!
	priceStr, ok := needle["price"].(string)
	if !ok {
		return 0.0, fmt.Errorf("failed to get the price: %s", needle)
	}
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0.0, err
	}
	if math.IsNaN(price) || math.IsInf(price, 0) || price < 0.0 {
		return 0.0, fmt.Errorf("invalid value returned by API: %v", price)
	}
	return price, nil
}

// checkForErrors checks for error messages in the response body
func checkForErrors(body []byte) error {
	strBody := string(body)
	switch {
	case strings.Contains(strBody, `{"error":"urn:flick:authentication:error:token_verification_failed"}`):
		return fmt.Errorf("invalid auth token")
	case strings.Contains(strBody, "405 Not Allowed"):
		return fmt.Errorf("something went wrong, got HTTP 405")
	}
	return nil
}
