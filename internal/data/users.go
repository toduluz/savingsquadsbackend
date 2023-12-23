package data

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/toduluz/savingsquadsbackend/internal/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	ID        primitive.ObjectID `json:"id,,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"-" bson:"created_at"`
	UpdatedAt time.Time          `json:"-" bson:"updated_at"`
	Name      string             `json:"name" bson:"name"`
	Email     string             `json:"email" bson:"email"`
	Password  Password           `json:"-" bson:"password"`
	Addresses []Address          `json:"addresses" bson:"addresses"`
	Phone     []Phone            `json:"phone" bson:"phone"`
	Vouchers  map[string]int     `json:"vouchers" bson:"vouchers"`
	Points    int                `json:"points" bson:"points"`
	Version   int                `json:"version" bson:"version"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

// UserModel struct wraps a sql.DB connection pool and allows us to work with the User struct type
// and the users table in our database.
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
	plaintext *string `json:"-" bson:"-"`
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

// Insert inserts a new record in the users table in our database for the user. Also, we check
// if our table already contains the same email address and if so return ErrDuplicateEmail error.
func (m UserModel) Insert(user *User) (primitive.ObjectID, error) {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Create a unique index on the email field if it doesn't exist.
	opts := options.CreateIndexes().SetMaxTime(3 * time.Second)
	keys := bson.D{{Key: "email", Value: 1}} // 1 for ascending order
	indexModel := mongo.IndexModel{Keys: keys, Options: options.Index().SetUnique(true)}
	_, err := m.DB.Collection("users").Indexes().CreateOne(ctx, indexModel, opts)
	if err != nil {
		return primitive.NilObjectID, err
	}

	// Insert the user data into the users table.
	result, err := m.DB.Collection("users").InsertOne(ctx, user)
	if err != nil {
		// Check if it's a duplicate key error (which means the email already exists).
		if writeException, ok := err.(mongo.WriteException); ok {
			for _, writeError := range writeException.WriteErrors {
				if writeError.Code == 11000 {
					return primitive.NilObjectID, ErrDuplicateEmail
				}
			}
		}
		return primitive.NilObjectID, err
	}

	return result.InsertedID.(primitive.ObjectID), nil
}

func (m UserModel) Get(id primitive.ObjectID) (*User, error) {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Define a User struct to hold the data returned by the query.
	var user User

	// Define the filter to match documents where id is id.
	filter := bson.M{"_id": id}

	// Execute the find operation
	err := m.DB.Collection("users").FindOne(ctx, filter).Decode(&user)
	if err != nil {
		// If the error is a NoDocument error, return ErrRecordNotFound
		switch {
		case err == mongo.ErrNoDocuments:
			return nil, ErrRecordNotFound
		default:
			// Otherwise, return the error
			return nil, err
		}
	}

	return &user, nil
}

// GetByEmail retrieves the User details from the database based on the user's email address.
// Because we have a UNIQUE constraint on the email column, this query will only return one record,
// or none at all, upon which we return a ErrRecordNotFound error).
func (m UserModel) GetByEmail(email string) (*User, error) {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Define a User struct to hold the data returned by the query.
	var user User

	// Define the filter to match documents where email is email.
	filter := bson.M{"email": email}

	// Execute the find operation
	err := m.DB.Collection("users").FindOne(ctx, filter).Decode(&user)
	if err != nil {
		// If the error is a NoDocument error, return ErrRecordNotFound
		switch {
		case err == mongo.ErrNoDocuments:
			return nil, ErrRecordNotFound
		default:
			// Otherwise, return the error
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) GetAllVouchers(id primitive.ObjectID) (map[string]int, error) {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Define the filter to match documents where id is id.
	filter := bson.M{"_id": id}

	// Execute the find operation
	var user User
	err := m.DB.Collection("users").FindOne(ctx, filter).Decode(&user)
	if err != nil {
		// If the error is a NoDocument error, return ErrRecordNotFound
		switch {
		case err == mongo.ErrNoDocuments:
			return nil, ErrRecordNotFound
		default:
			// Otherwise, return the error
			return nil, err
		}
	}

	return user.Vouchers, nil
}

func (m UserModel) RedeemVoucher(id primitive.ObjectID, code string, number int) error {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Define the filter to match documents where id is id and vouchers does not contain code.
	filter := bson.M{"_id": id, "vouchers." + code: bson.M{"$exists": false}}

	// Define the update document to set the new values of the fields.
	update := bson.M{
		"$set": bson.M{"vouchers." + code: number},
	}

	// Execute the update operation.
	result, err := m.DB.Collection("users").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// If no document was updated, it means the voucher was already in the vouchers array.
	if result.ModifiedCount == 0 {
		return ErrVoucherAlreadyRedeeemed
	}

	return nil
}

func (m UserModel) GetPoints(id primitive.ObjectID) (int, error) {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Define the filter to retrieve the document for the user with the specified id.
	filter := bson.M{"_id": id}

	// Define the options to set the limit to 1 document.
	opts := options.FindOne().SetProjection(bson.M{"points": 1})

	// Execute the find operation
	var user User
	err := m.DB.Collection("users").FindOne(ctx, filter, opts).Decode(&user)
	if err != nil {
		// If the error is a NoDocument error, return ErrRecordNotFound
		switch {
		case err == mongo.ErrNoDocuments:
			return 0, ErrRecordNotFound
		default:
			// Otherwise, return the error
			return 0, err
		}
	}

	return user.Points, nil
}

func (m UserModel) AddPoints(id primitive.ObjectID, points int) error {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Define the filter to match documents where id is id.
	filter := bson.M{"_id": id}

	// Define the update document to set the new values of the fields.
	update := bson.M{
		"$inc": bson.M{
			"points": points,
		},
	}

	// Execute the update operation.
	_, err := m.DB.Collection("users").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (m UserModel) DeductPointsAndCreateVoucher(id primitive.ObjectID, points int, voucher *Voucher) error {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Start a new session.
	session, err := m.DB.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// Start a transaction.
	err = session.StartTransaction()
	if err != nil {
		return err
	}

	// Define the filter to match documents where id is id and points is greater than or equal to points.
	filter := bson.M{"_id": id, "points": bson.M{"$gte": points}, "vouchers." + voucher.Code: bson.M{"$exists": false}}

	// Define the update document to set the new values of the fields.
	update := bson.M{
		"$set": bson.M{
			"vouchers." + voucher.Code: 1,
		},
		"$inc": bson.M{
			"points": -points,
		},
	}

	// Execute the update operation.
	res, err := m.DB.Collection("users").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.ModifiedCount == 0 {
		return ErrExchangePointsForVoucher
	}

	// Create a new voucher.
	_, err = m.DB.Collection("vouchers").InsertOne(ctx, voucher)
	if err != nil {
		var writeException mongo.WriteException
		if errors.As(err, &writeException) {
			for _, writeError := range writeException.WriteErrors {
				if writeError.Code == 11000 {
					// Handle duplicate key error
					return ErrVoucherAlreadyExists
				}
			}
		}
		return err
	}

	// Commit the transaction.
	err = session.CommitTransaction(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (m UserModel) UpdateVoucherList(id primitive.ObjectID, vouchers map[string]int) error {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Define the filter to retrieve the document for the user with the specified id.
	filter := bson.M{"_id": id}

	// Define the update document to set the new values of the fields.
	update := bson.M{
		"$set": bson.M{
			"vouchers": vouchers,
		},
	}

	// Execute the update operation.
	_, err := m.DB.Collection("users").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
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
