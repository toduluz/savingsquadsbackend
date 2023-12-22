package main

import (
	"net/http"

	"github.com/toduluz/savingsquadsbackend/internal/cookies"
)

func (app *application) setCookie(w http.ResponseWriter, name, value string, maxAge int) error {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	// Use the WriteSigned() function, passing in the secret key as the final
	// argument.
	err := cookies.WriteSigned(w, cookie, []byte(app.config.jwt.secret))
	if err != nil {
		return err
	}
	return nil
}

func (app *application) getCookie(r *http.Request) (string, error) {
	// Use the ReadSigned() function, passing in the secret key as the final
	// argument.
	value, err := cookies.ReadSigned(r, "exampleCookie", []byte(app.config.jwt.secret))
	if err != nil {
		return "", err
	}

	return value, nil
}
