package data

import (
	"errors"
	"log"
	"time"

	"github.com/toduluz/savingsquadsbackend/internal/validator"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail           = errors.New("duplicate email")
	ErrVoucherAlreadyExists     = errors.New("voucher already exists")
	ErrExchangePointsForVoucher = errors.New("problem exchanging points for voucher")
	ErrVoucherAlreadyRedeeemed  = errors.New("voucher already redeemed")
)

// AnonymousUser represents an anonymous user.
var AnonymousUser = &User{}

// User type whose fields describe a user. Note, that we use the json:"-" struct tag to prevent
// the Password and Version fields from appearing in any output when we encode it to JSON.
// Also, notice that the Password field uses the custom password type defined below.
type User struct {
	ID        string         `json:"id,,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time      `json:"-" bson:"created_at"`
	UpdatedAt time.Time      `json:"-" bson:"updated_at"`
	Name      string         `json:"name" bson:"name"`
	Email     string         `json:"email" bson:"email"`
	Password  Password       `json:"-" bson:"password"`
	Addresses []Address      `json:"addresses" bson:"addresses"`
	Phone     []Phone        `json:"phone" bson:"phone"`
	Vouchers  map[string]int `json:"vouchers" bson:"vouchers"`
	Points    int            `json:"points" bson:"points"`
	Version   int            `json:"version" bson:"version"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

// UserModel struct DB and allows us to work with the User struct type
// and the users collection in our database.
type UserModel struct {
	DB       *mongo.Database
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

// password tyep is a struct containing the plaintext and hashed version of a password for a User.
// The plaintext field is a *pointer* to a string, so that we're able to distinguish between a
// plaintext password not being present in the struct at all, versus a plaintext password which
// is the empty string "".
type Password struct {
	plaintext *string `json:"-" bson:"-,omitempty"`
	Hash      []byte  `json:"-" bson:"hash"`
}

// Address type is a struct containing the address of a User.
type Address struct {
	Street     string `json:"street" bson:"street"`
	Number     string `json:"number" bson:"number"`
	PostalCode int    `json:"postal_code" bson:"postal_code"`
	City       string `json:"city" bson:"city"`
}

// Phone type is a struct containing the phone of a User.
type Phone struct {
	CountryNumber string `json:"country_number" bson:"country_number"`
	Number        string `json:"number" bson:"number"`
}

// Set calculates the bcrypt hash of a plaintext password, and stores both the has and the
// plaintext versions in the password struct.
func (p *Password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.Hash = hash
	return nil
}

// Matches checks whether the provided plaintext password matches the hashed password stored in
// the password struct, returning true if it matches and false otherwise.
func (p *Password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

// ValidatePoints checks that the Points field is not a negative integer.
func ValidatePoints(v *validator.Validator, points int) {
	v.Check(points >= 0, "points", "must be a positive integer")
}

// ValidateEmail checks that the Email field is not an empty string and that it matches the regex
// for email addresses, validator.EmailRX.
func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be valid email address")
}

// ValidatePasswordPlaintext validtes that the password is not an empty string and is between 8 and
// 72 bytes long.
func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	// validate user.Name
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	// Validate email
	ValidateEmail(v, user.Email)

	// If the plaintext password is not nil, call the standalone ValidatePasswordPlaintext helper.
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	// If the password has is ever nil, this will be due to a logic error in our codebase
	// (probably because we forgot to set a password for the user). It's a useful sanity check to
	// include here, but it's not a problem with the data provided by the client. So, rather
	// than adding an error to the validation map we raise a panic instead.
	if user.Password.Hash == nil {
		panic("missing password hash for user")
	}
}
