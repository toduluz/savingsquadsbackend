package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestIntegrationRegisterUserHandler(t *testing.T) {
	t.Parallel()

	app, closeDB := newIntegrationTestApplication(t)
	defer closeDB()

	tests := []struct {
		name     string
		userJSON string
		wantCode int
		wantBody string
	}{
		{"Valid User", `{"name":"Test User","email":"test@example.com","password":"password123"}`, http.StatusAccepted, `{ 
			"user": {
				"name": "Test User",
				"email": "test@example.com",
				"password": "password123",
				"addresses": [],
				"phone": [],
				"vouchers": {},
				"points": 0,
				"version": 1
			}}`},
		{"Duplicate User", `{"name":"Test User","email":"test@example.com","password":"password123"}`, http.StatusUnprocessableEntity, `{
			"error": {
				"email": "a user with this email address already exists"
			}}`},
		{"Invalid Email", `{"name":"Test User","email":"test","password":"password123"}`, http.StatusUnprocessableEntity, `{
			"error": {
				"email": "must be valid email address"
			}}`},
		{"Missing Email", `{"name":"Test User","password":"password123"}`, http.StatusUnprocessableEntity, `{
			"error": {
				"email": "must be provided"
			}}`},
		{"Invalid Password", `{"name":"Test User","email":"test@example.com","password":"123"}`, http.StatusUnprocessableEntity, `{
			"error": {
				"password": "must be at least 8 bytes long"
			}}`},
		{"Missing Password", `{"name":"Test User","email":"test@exmaple.com"}`, http.StatusUnprocessableEntity, `{
			"error": {
				"password": "must be provided"
			}}`},
		{"Missing Name", `{"email":"test@example.com","password":"password123"}`, http.StatusUnprocessableEntity, `{
			"error": {
				"name": "must be provided"
			}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/v1/user/register", bytes.NewBufferString(tt.userJSON))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			// handler := http.HandlerFunc(app.registerUserHandler)
			router := app.Routes()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.wantCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.wantCode)
			}

			var gotBody, wantBody map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &gotBody); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(tt.wantBody), &wantBody); err != nil {
				t.Fatal(err)
			}

			if tt.name == "Valid User" {
				return
			}

			if !reflect.DeepEqual(gotBody, wantBody) {
				t.Errorf("handler returned unexpected body: got %v want %v",
					gotBody, wantBody)
			}
		})
	}
}
