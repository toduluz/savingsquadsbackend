package data

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Secret key used for signing the JWTs. This should be stored securely and not exposed.
var jwtKey = []byte("your_secret_key")

// Claims struct to be encoded to JWT.
type Claims struct {
	UserID string `json:"userID"`
	jwt.StandardClaims
}

// GenerateToken generates a new JWT for a user.
func GenerateToken(userID string) (string, error) {
	// Set the claims, which includes the user ID and expiry time.
	claims := &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds.
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Create the JWT string.
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates the JWT.
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, err
		}
		return nil, err
	}

	if !token.Valid {
		return nil, err
	}

	return claims, nil
}
