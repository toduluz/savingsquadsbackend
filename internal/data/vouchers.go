package data

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

func (m VoucherModel) GetVoucherList(voucherCodes []string) ([]Voucher, error) {
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Define a Voucher to decode the document into
	var vouchers []Voucher

	// Execute the find operation
	cursor, err := m.DB.Collection("vouchers").Find(ctx, bson.M{"_id": bson.M{"$in": voucherCodes}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode the results into a slice of Vouchers.
	for cursor.Next(ctx) {
		var voucher Voucher
		if err = cursor.Decode(&voucher); err != nil {
			return nil, err
		}
		vouchers = append(vouchers, voucher)
	}

	return vouchers, nil
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

func (m VoucherModel) GetAllVouchers(code string, starts time.Time, expires time.Time, active bool, minSpend int, category string, f *Filters) ([]Voucher, *Metadata, error) {
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
		filter = append(filter, bson.E{"_id", bson.M{"$gt": f.Cursor}})
	}
	// Execute the MongoDB find operation with limit and sort.
	opts := options.Find().SetLimit(int64(f.limit())).SetSort(bson.D{{f.sortColumn(), sortDirection}})
	cursor, err := m.DB.Collection("vouchers").Find(ctx, filter, opts)
	if err != nil {
		return nil, &Metadata{}, err
	}
	defer cursor.Close(ctx)

	// Decode the results into a slice of Vouchers.
	var vouchers []Voucher
	for cursor.Next(ctx) {
		var voucher Voucher
		if err = cursor.Decode(&voucher); err != nil {
			return nil, &Metadata{}, err
		}
		vouchers = append(vouchers, voucher)
	}

	// Generate a PaginationData struct, passing in the page size and the _id of the last document.
	var metadata Metadata
	if len(vouchers) > 0 {
		metadata = formatPaginationData(f.PageSize, vouchers[len(vouchers)-1].Code)
	}

	// If everything went OK, then return the slice of the vouchers and paginationData.
	return vouchers, &metadata, nil
}
