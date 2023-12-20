package data

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/toduluz/savingsquadsbackend/internal/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Voucher struct {
	Code         string    `json:"id" bson:"_id"`
	CreatedAt    time.Time `json:"-" bson:"created_at"`
	ModifiedAt   time.Time `json:"-" bson:"updated_at"`
	Description  string    `json:"description" bson:"description"`
	Discount     int       `json:"discount" bson:"discount"`
	IsPercentage bool      `json:"isPercentage" bson:"isPercentage"`
	Starts       time.Time `json:"start" bson:"start"`
	Expires      time.Time `json:"expires" bson:"expires"`
	Active       bool      `json:"active" bson:"active"`
	UsageLimit   int       `json:"usageLimit" bson:"usageLimit"`
	UsageCount   int       `json:"usageCount" bson:"usageCount"`
	MinSpend     int       `json:"minSpend" bson:"minSpend"`
	Category     string    `json:"category" bson:"category"`
}

func (v *Voucher) vocuherCodeGenerator() error {
	b := make([]byte, 15) // Generate 15 random bytes
	_, err := rand.Read(b)
	if err != nil {
		return err
	}

	code := base64.URLEncoding.EncodeToString(b)

	// Base64 encoding can include '/' and '+' characters, replace them to avoid issues
	code = strings.ReplaceAll(code, "/", "a")
	code = strings.ReplaceAll(code, "+", "b")

	v.Code = code
	// Return the code
	return nil
}

type VoucherModel struct {
	DB       *mongo.Database
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

// Insert a new voucher record into the vouchers table. If the voucher code already exists and is active, return an error. Else,
// insert the new voucher record and return nil. If the voucher code already exists but is inactive, insert the new voucher record.
func (m VoucherModel) Insert(voucher *Voucher) error {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Insert the voucher into the "vouchers" collection.
	_, err := m.DB.Collection("vouchers").InsertOne(ctx, voucher)
	if err != nil {
		// Check if it's a duplicate key error (which means the voucher code already exists).
		if writeException, ok := err.(mongo.WriteException); ok {
			for _, writeError := range writeException.WriteErrors {
				if writeError.Code == 11000 {
					return errors.New("a voucher with the provided voucher code already exists")
				}
			}
		}
		return err
	}

	return nil
}

func (m VoucherModel) InsertGeneratedVoucher(ctx mongo.SessionContext, voucher *Voucher) (string, error) {
	// Generate a voucher code
	err := voucher.vocuherCodeGenerator()
	if err != nil {
		return "", err
	}
	// Insert the voucher into the "vouchers" collection.
	_, err = m.DB.Collection("vouchers").InsertOne(ctx, voucher)
	if err != nil {
		return "", err
	}

	return voucher.Code, nil
}

// Get returns a specific Voucher based on its id.
func (m VoucherModel) Get(code string) (*Voucher, error) {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Define a Voucher to decode the document into
	var voucher Voucher

	// Execute the find operation
	err := m.DB.Collection("vouchers").FindOne(ctx, bson.M{"_id": code}).Decode(&voucher)
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

	return &voucher, nil
}

func (m VoucherModel) GetVoucherList(voucherCodes []string) ([]Voucher, []string, error) {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Define a Voucher to decode the document into
	var vouchers []Voucher
	var newVoucherList []string

	// Execute the find operation
	cursor, err := m.DB.Collection("vouchers").Find(ctx, bson.M{"_id": bson.M{"$in": voucherCodes}})
	if err != nil {
		return nil, nil, err
	}
	defer cursor.Close(ctx)

	// Decode the results into a slice of Vouchers.
	for cursor.Next(ctx) {
		var voucher Voucher
		if err = cursor.Decode(&voucher); err != nil {
			return nil, nil, err
		}
		if voucher.Active {
			newVoucherList = append(newVoucherList, voucher.Code)
			vouchers = append(vouchers, voucher)
		}
	}

	return vouchers, newVoucherList, nil
}

// UpdateUsageCount updates the usageCount and active fields of a specific voucher record in the vouchers table. If the usageCount is
// less than the usageLimit, increment the usageCount by 1 and set the active field to true. If the usageCount is equal to the usageLimit,
// set the active field to false. If the voucher code does not exist, return an error. If the voucher code exists but active is false,
// return an error.
func (m VoucherModel) UpdateUsageCount(code string) error {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	filter := bson.M{"_id": code, "active": true}
	update := []bson.M{
		{
			"$set": bson.M{
				"usageCount": bson.M{
					"$cond": []interface{}{
						bson.M{"$lt": []interface{}{"$usageCount", "$usageLimit"}},
						bson.M{"$add": []interface{}{"$usageCount", 1}},
						"$usageCount",
					},
				},
				"active": bson.M{
					"$cond": []interface{}{
						bson.M{"$eq": []interface{}{bson.M{"$add": []interface{}{"$usageCount", 1}}, "$usageLimit"}},
						false,
						true,
					},
				},
			},
		},
	}

	// Execute the MongoDB update operation.
	result, err := m.DB.Collection("vouchers").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrRecordNotFound
	}

	if result.ModifiedCount == 0 {
		return ErrEditConflict
	}

	return nil
}

// Delete is a placeholder method for deleting a specific record in the Vouchers table.
func (m VoucherModel) Delete(code string) error {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Define the filter to match the document to delete.
	filter := bson.M{"_id": code}

	// Execute the MongoDB delete operation.
	result, err := m.DB.Collection("vouchers").DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m *VoucherModel) GetAllVouchers(code string, starts time.Time, expires time.Time, active bool, minSpend int, category string, f *Filters) ([]*Voucher, Metadata, error) {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Determine the sort direction.
	sortDirection := f.sortDirection()

	// Build a filter based on the provided parameters.
	filter := bson.D{}
	if code != "" {
		filter = append(filter, bson.E{"_id", code})
	}
	if !starts.IsZero() {
		filter = append(filter, bson.E{"start", bson.M{"$gte": starts}})
	}
	if !expires.IsZero() {
		filter = append(filter, bson.E{"expires", bson.M{"$lte": expires}})
	}
	if active {
		filter = append(filter, bson.E{"active", active})
	}
	if category != "" {
		filter = append(filter, bson.E{"category", category})
	}
	if minSpend != 0 {
		filter = append(filter, bson.E{"minSpend", bson.M{"$lte": minSpend}})
	}

	// If a cursor is provided, add a condition to the filter to only find documents with an _id greater than the cursor.
	if f.Cursor != "" {
		cursorID, err := primitive.ObjectIDFromHex(f.Cursor)
		if err != nil {
			return nil, Metadata{}, err
		}
		filter = append(filter, bson.E{"_id", bson.M{"$gt": cursorID}})
	}
	// Execute the MongoDB find operation with limit and sort.
	opts := options.Find().SetLimit(int64(f.limit())).SetSort(bson.D{{f.sortColumn(), sortDirection}})
	cursor, err := m.DB.Collection("vouchers").Find(ctx, filter, opts)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer cursor.Close(ctx)

	// Decode the results into a slice of Vouchers.
	var vouchers []*Voucher
	for cursor.Next(ctx) {
		var voucher Voucher
		if err = cursor.Decode(&voucher); err != nil {
			return nil, Metadata{}, err
		}
		vouchers = append(vouchers, &voucher)
	}

	// Generate a PaginationData struct, passing in the page size and the _id of the last document.
	metadata := formatPaginationData(f.PageSize, vouchers[len(vouchers)-1].Code)

	// If everything went OK, then return the slice of the vouchers and paginationData.
	return vouchers, metadata, nil
}

// ValidateVoucher runs validation checks on the Voucher type.
func ValidateVoucher(v *validator.Validator, voucher *Voucher) {
	v.Check(voucher.Code != "", "code", "must be provided")
	v.Check(len(voucher.Code) <= 20, "code", "must not be more than 20 characters long")

	v.Check(voucher.Description != "", "description", "must be provided")
	v.Check(len(voucher.Description) <= 500, "description", "must not be more than 500 characters long")

	v.Check(voucher.Discount >= 0, "discount", "must be a positive number")
	v.Check(voucher.Discount <= 100, "discount", "must not be more than 100")

	v.Check(voucher.UsageLimit >= 0, "usageLimit", "must be a positive number")

	v.Check(voucher.Starts.Before(voucher.Expires), "start", "must be before the expiry date")
}
