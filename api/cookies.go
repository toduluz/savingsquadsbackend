package api

import (
	"net/http"

	"github.com/toduluz/savingsquadsbackend/internal/cookies"
)

func (app *Application) setCookie(w http.ResponseWriter, name, value string, maxAge int) error {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/v1/users",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	// Use the WriteSigned() function, passing in the secret key as the final
	// argument.
	err := cookies.WriteSigned(w, cookie, []byte(app.Config.Jwt.Secret))
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) getCookie(r *http.Request, value string) (string, error) {
	// Use the ReadSigned() function, passing in the secret key as the final
	// argument.
	value, err := cookies.ReadSigned(r, value, []byte(app.Config.Jwt.Secret))
	if err != nil {
		return "", err
	}

	return value, nil
}
