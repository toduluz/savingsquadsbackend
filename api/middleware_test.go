package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthenticate(t *testing.T) {

	app := newTestApplication(t)
	// Define the test cases.
	tests := []struct {
		name           string
		wantStatusCode int
	}{
		{"Invalid JWT cookie", http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new instance of our application struct which uses a mocked
			// getCookie method.

			// Create a new recorder to capture the response.
			w := httptest.NewRecorder()

			// Create a new request with the "jwt" cookie.
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.name == "Invalid JWT cookie" {
				r.AddCookie(&http.Cookie{Name: "jwt", Value: "dummy"})
			}

			// Create a mock HTTP handler that we can pass to our authenticate middleware,
			// which writes a 200 status code and "OK" response body.
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("OK"))
			})

			// Pass the mock HTTP handler to our authenticate middleware.
			app.authenticate(next).ServeHTTP(w, r)

			// Check that the status code of the response is the expected one.
			rs := w.Result()
			if rs.StatusCode != tc.wantStatusCode {
				t.Errorf("want %d; got %d", tc.wantStatusCode, rs.StatusCode)
			}
		})
	}
}
