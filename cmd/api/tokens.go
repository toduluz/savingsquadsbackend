package main

import (
	"errors"
	"time"

	"github.com/pascaldekloe/jwt" //
	"github.com/toduluz/savingsquadsbackend/internal/data"
)

func (app *application) createJWTClaims(user *data.User) ([]byte, error) {
	// Create a JWT claims struct containing the user ID as the subject, with an issued
	// time of now and validity window of the next 24 hours. We also set the issuer and
	// audience to a unique identifier for our application.
	var claims jwt.Claims
	claims.Subject = user.ID.Hex()
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add(24 * time.Hour))
	// claims.Issuer = "greenlight.alexedwards.net"
	// claims.Audiences = []string{"greenlight.alexedwards.net"}
	claims.Set = map[string]interface{}{"version": user.Version}

	// Sign the JWT claims using the HMAC-SHA256 algorithm and the secret key from the
	// application config. This returns a []byte slice containing the JWT as a base64-
	// encoded string.
	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secret))
	if err != nil {
		return nil, err
	}
	return jwtBytes, nil
}

func (app *application) validateJWTClaims(token string) (*jwt.Claims, error) {
	// Parse the JWT and extract the claims. This will return an error if the JWT
	// contents doesn't match the signature (i.e. the token has been tampered with)
	// or the algorithm isn't valid.
	claims, err := jwt.HMACCheck([]byte(token), []byte(app.config.jwt.secret))
	if err != nil {
		return nil, err
	}
	// Check if the JWT is still valid at this moment in time.
	if !claims.Valid(time.Now()) {
		return nil, errors.New("token has expired")
	}
	return claims, nil
}
