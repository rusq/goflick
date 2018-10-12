// Package goflick package for interrogating Flick Electric API, a NZ power retailer
package goflick

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// FlickConnect stores the authentication information
type FlickConnect struct {
	Token       string `json:"id_token"`     // Token contains OAuth token from server
	TokenType   string `json:"token_type"`   // TokenType type of the token, usually "bearer"
	Expires     int    `json:"expires_in"`   // Expires not used
	AccessToken string `json:"access_token"` // AccessToken not used
}

const (
	//APIURL Flick API URL
	APIURL       = "https://api.flick.energy"
	clientID     = "le37iwi3qctbduh39fvnpevt1m2uuvz"
	clientSecret = "ignwy9ztnst3azswww66y9vd9zt6qnt"
)

// FlickAPIs flick API endpoints
var FlickAPIs = map[string]string{
	"auth":  "/identity/oauth/token",
	"price": "/customer/mobile_provider/price",
}

// NewConnect returns a new connection to Flick
func NewConnect(username, password string) (*FlickConnect, error) {
	resp, err := http.PostForm(getAPIURL("auth"), url.Values{
		"grant_type":    {"password"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"username":      {username},
		"password":      {password},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %s", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if loginFailed(body) {
		return nil, fmt.Errorf("invalid credentials")
	}
	fc, err := parseAuth(body)
	if err != nil {
		return nil, err
	}
	return fc, nil
}

// GetPrice returns current electricity price forecast
func (fc *FlickConnect) GetPrice() (price float64, err error) {
	body, err := fc.APIcall("price")
	if err != nil {
		return
	}
	price, err = parsePrice(body)
	return
}

// MustGetPrice returns the current price, but if error occurs - panics
func (fc *FlickConnect) MustGetPrice() (price float64) {
	price, err := fc.GetPrice()
	if err != nil {
		panic(err)
	}
	return
}

//getAPIURL returns the proper URL for the API
func getAPIURL(endpoint string) string {
	endpoint, ok := FlickAPIs[endpoint]
	if !ok {
		return ""
	}
	return APIURL + endpoint
}

// APIcall calls a Flick API augmenting headers with auth token
func (fc *FlickConnect) APIcall(endpoint string) ([]byte, error) {
	if fc.Token == "" || fc.TokenType == "" {
		return nil, fmt.Errorf("unauthorized apiCall(), Token and/or TokenType are empty")
	}
	client := http.Client{}
	url := getAPIURL(endpoint)
	if url == "" {
		return nil, fmt.Errorf("invalid endpoint")
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", strings.Title(fc.TokenType)+" "+fc.Token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err = checkForErrors(body); err != nil {
		return nil, err
	}
	return body, err
}
