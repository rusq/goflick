package goflick

import (
	"reflect"
	"testing"
)

const (
	respAuthSuccess  = `{"access_token":"aCC3ss","expires_in":5184000,"id_token":"1d_T0Ken","token_type":"bearer"}`
	respAuthFailure  = `Auth failed invalid_grant The access grant you supplied is invalid`
	respPrice        = `{"kind": "mobile_provider_price", "needle": {"status": "urn:flick:market:price:forecast", "charge_methods": ["kwh", "spot_price"], "start_at": "2018-07-25T10:00:00Z", "components": [{"kind": "component", "unit_code": "cents", "charge_method": "kwh", "per": "kwh", "charge_setter": "retailer", "_links": {}, "value": "1.58"}, {"kind": "component", "unit_code": "cents", "charge_method": "kwh", "per": "kwh", "charge_setter": "ea", "_links": {}, "value": "0.113"}, {"kind": "component", "unit_code": "cents", "charge_method": "kwh", "per": "kwh", "charge_setter": "metering", "_links": {}, "value": "0.0"}, {"kind": "component", "unit_code": "cents", "charge_method": "kwh", "per": "kwh", "charge_setter": "generation", "_links": {}, "value": "0.0"}, {"kind": "component", "unit_code": "cents", "charge_method": "kwh", "per": "kwh", "charge_setter": "admin", "_links": {}, "value": "0.0"}, {"kind": "component", "unit_code": "cents", "charge_method": "kwh", "per": "kwh", "charge_setter": "network", "_links": {}, "value": "5.51"}, {"kind": "component", "unit_code": "cents", "charge_method": "spot_price", "per": "kwh", "charge_setter": "ea", "_links": {}, "value": "7.253"}], "unit_code": "cents", "price": "14.456", "now": "2018-07-25T10:01:09.013Z", "type": "rated", "per": "kwh", "end_at": "2018-07-25T10:29:59Z"}, "customer_state": "active"}`
	respPriceInvalid = `{"kind": "mobile_provider_price", "needle": {}, "customer_state": "active"}`
)

var sampleFlickConnect = FlickConnect{
	Token:       `1d_T0Ken`,
	TokenType:   `bearer`,
	Expires:     5184000,
	AccessToken: `aCC3ss`,
}

func TestNewConnect(t *testing.T) {
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    *FlickConnect
		wantErr bool
	}{
		{"invalid login", args{"xxxx", "xxxx"}, nil, true},
		// valid login would need to have live credentials here
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConnect(tt.args.username, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConnect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConnect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_loginFailed(t *testing.T) {
	type args struct {
		response []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"valid login", args{[]byte(respAuthSuccess)}, false},
		{"invalid login", args{[]byte(respAuthFailure)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := loginFailed(tt.args.response); got != tt.want {
				t.Errorf("loginFailed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseAuth(t *testing.T) {
	type args struct {
		response []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *FlickConnect
		wantErr bool
	}{
		{"valid token", args{[]byte(respAuthSuccess)}, &sampleFlickConnect, false},
		{"invalid token", args{[]byte(`{"valid_json":true}`)}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAuth(tt.args.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseAuth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parsePrice(t *testing.T) {
	type args struct {
		response []byte
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{"valid price", args{[]byte(respPrice)}, 14.456, false},
		{"empty needle", args{[]byte(respPriceInvalid)}, 0.0, true},
		{"invalid json", args{[]byte("helo world")}, 0.0, true},
		{"NaN", args{[]byte(`{"needle":{"price":"NaN"}}`)}, 0.0, true},
		{"Inf", args{[]byte(`{"needle":{"price":"Inf"}}`)}, 0.0, true},
		{"less than 0", args{[]byte(`{"needle":{"price":"-4.12"}}`)}, 0.0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePrice(tt.args.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePrice() error = %v, wantErr %v, price = %v", err, tt.wantErr, got)
				return
			}
			if got != tt.want {
				t.Errorf("parsePrice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkForErrors(t *testing.T) {
	type args struct {
		body []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"405", args{[]byte(`{"_status":"405 Not Allowed"}`)}, true},
		{"auth_err", args{[]byte(`{"error":"urn:flick:authentication:error:token_verification_failed"}`)}, true},
		{"all ok", args{[]byte(respAuthSuccess)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkForErrors(tt.args.body); (err != nil) != tt.wantErr {
				t.Errorf("checkForErrors() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFlickConnect_apiCall(t *testing.T) {
	type fields struct {
		Token       string
		TokenType   string
		Expires     int
		AccessToken string
	}
	type args struct {
		endpoint string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantBody []byte
		wantErr  bool
	}{
		{"panic", fields{Token: ""}, args{"price"}, nil, true},
		{"panic TokenType", fields{TokenType: ""}, args{"price"}, nil, true},
		{"invalid auth", fields{Token: "123", TokenType: "456"}, args{"price"}, nil, true},
		{"invalid endpoint", fields{Token: "123", TokenType: "bearer"}, args{"$$$$"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := &FlickConnect{
				Token:       tt.fields.Token,
				TokenType:   tt.fields.TokenType,
				Expires:     tt.fields.Expires,
				AccessToken: tt.fields.AccessToken,
			}

			gotBody, err := fc.APIcall(tt.args.endpoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("FlickConnect.apiCall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotBody, tt.wantBody) {
				t.Errorf("FlickConnect.apiCall() = %v, want %v", string(gotBody), string(tt.wantBody))
			}
		})
	}
}
