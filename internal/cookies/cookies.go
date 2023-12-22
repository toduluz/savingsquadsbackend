package cookies

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
)

var (
	ErrValueTooLong = errors.New("cookie value too long")
	ErrInvalidValue = errors.New("invalid cookie value")
)

func Write(w http.ResponseWriter, cookie http.Cookie) error {
	// Encode the cookie value using base64.
	cookie.Value = base64.URLEncoding.EncodeToString([]byte(cookie.Value))

	// Check the total length of the cookie contents. Return the ErrValueTooLong
	// error if it's more than 4096 bytes.
	if len(cookie.String()) > 4096 {
		return ErrValueTooLong
	}

	// Write the cookie as normal.
	http.SetCookie(w, &cookie)

	return nil
}

func Read(r *http.Request, name string) (string, error) {
	// Read the cookie as normal.
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	// Decode the base64-encoded cookie value. If the cookie didn't contain a
	// valid base64-encoded value, this operation will fail and we return an
	// ErrInvalidValue error.
	value, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", ErrInvalidValue
	}

	// Return the decoded cookie value.
	return string(value), nil
}

func WriteSigned(w http.ResponseWriter, cookie http.Cookie, secretKey []byte) error {
	// Calculate a HMAC signature of the cookie name and value, using SHA256 and
	// a secret key (which we will create in a moment).
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(cookie.Name))
	mac.Write([]byte(cookie.Value))
	signature := mac.Sum(nil)

	// Prepend the cookie value with the HMAC signature.
	cookie.Value = string(signature) + cookie.Value

	// Call our Write() helper to base64-encode the new cookie value and write
	// the cookie.
	return Write(w, cookie)
}

func ReadSigned(r *http.Request, name string, secretKey []byte) (string, error) {
	// Read in the signed value from the cookie. This should be in the format
	// "{signature}{original value}".
	signedValue, err := Read(r, name)
	if err != nil {
		return "", err
	}

	// A SHA256 HMAC signature has a fixed length of 32 bytes. To avoid a potential
	// 'index out of range' panic in the next step, we need to check sure that the
	// length of the signed cookie value is at least this long. We'll use the
	// sha256.Size constant here, rather than 32, just because it makes our code
	// a bit more understandable at a glance.
	if len(signedValue) < sha256.Size {
		return "", ErrInvalidValue
	}

	// Split apart the signature and original cookie value.
	signature := signedValue[:sha256.Size]
	value := signedValue[sha256.Size:]

	// Recalculate the HMAC signature of the cookie name and original value.
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(name))
	mac.Write([]byte(value))
	expectedSignature := mac.Sum(nil)

	// Check that the recalculated signature matches the signature we received
	// in the cookie. If they match, we can be confident that the cookie name
	// and value haven't been edited by the client.
	if !hmac.Equal([]byte(signature), expectedSignature) {
		return "", ErrInvalidValue
	}

	// Return the original cookie value.
	return value, nil
}
