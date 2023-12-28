package data

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Insert inserts a new record in the users table in our database for the user. Also, we check
// if our table already contains the same email address and if so return ErrDuplicateEmail error.
func (m UserModel) Insert(user *User) (string, error) {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Create a unique index on the email field if it doesn't exist.
	opts := options.CreateIndexes().SetMaxTime(3 * time.Second)
	keys := bson.D{{Key: "email", Value: 1}} // 1 for ascending order
	indexModel := mongo.IndexModel{Keys: keys, Options: options.Index().SetUnique(true)}
	_, err := m.DB.Collection("users").Indexes().CreateOne(ctx, indexModel, opts)
	if err != nil {
		return "", err
	}

	// Insert the user data into the users table.
	result, err := m.DB.Collection("users").InsertOne(ctx, user)
	if err != nil {
		// Check if it's a duplicate key error (which means the email already exists).
		if writeException, ok := err.(mongo.WriteException); ok {
			for _, writeError := range writeException.WriteErrors {
				if writeError.Code == 11000 {
					return "", ErrDuplicateEmail
				}
			}
		}
		return "", err
	}

	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (m UserModel) Get(id string) (*User, error) {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Convert the id string to a MongoDB ObjectId.
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// Define a User struct to hold the data returned by the query.
	var user User

	// Define the filter to match documents where id is id.
	filter := bson.M{"_id": oid}

	// Execute the find operation
	err = m.DB.Collection("users").FindOne(ctx, filter).Decode(&user)
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

func (m UserModel) GetAllVouchers(id string) (map[string]int, error) {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// Define the filter to match documents where id is id.
	filter := bson.M{"_id": oid}

	// Execute the find operation
	var user User
	err = m.DB.Collection("users").FindOne(ctx, filter).Decode(&user)
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

func (m UserModel) RedeemVoucher(id string, code string, number int) error {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Define the filter to match documents where id is id and vouchers does not contain code.
	filter := bson.M{"_id": oid, "vouchers." + code: bson.M{"$exists": false}}

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

func (m UserModel) GetPoints(id string) (int, error) {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, err
	}

	// Define the filter to retrieve the document for the user with the specified id.
	filter := bson.M{"_id": oid}

	// Define the options to set the limit to 1 document.
	opts := options.FindOne().SetProjection(bson.M{"points": 1})

	// Execute the find operation
	var user User
	err = m.DB.Collection("users").FindOne(ctx, filter, opts).Decode(&user)
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

func (m UserModel) AddPoints(id string, points int) error {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Define the filter to match documents where id is id.
	filter := bson.M{"_id": oid}

	// Define the update document to set the new values of the fields.
	update := bson.M{
		"$inc": bson.M{
			"points": points,
		},
	}

	// Execute the update operation.
	_, err = m.DB.Collection("users").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (m UserModel) DeductPointsAndCreateVoucher(id string, points int, voucher *Voucher) error {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

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
	filter := bson.M{"_id": oid, "points": bson.M{"$gte": points}, "vouchers." + voucher.Code: bson.M{"$exists": false}}

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

func (m UserModel) UpdateVoucherList(id string, vouchers map[string]int) error {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Define the filter to retrieve the document for the user with the specified id.
	filter := bson.M{"_id": oid}

	// Define the update document to set the new values of the fields.
	update := bson.M{
		"$set": bson.M{
			"vouchers": vouchers,
		},
	}

	// Execute the update operation.
	_, err = m.DB.Collection("users").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}
