package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Voucher struct
type Voucher struct {
	ID             primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	IsPercentage   bool               `json:"isPercentage" bson:"isPercentage"`
	IsUserSpecific bool               `json:"isUserSpecific" bson:"isUserSpecific"`
	UserID         string             `json:"userID" bson:"userID"`
	Discount       float64            `json:"discount" bson:"discount"`
	MaxDiscount    float64            `json:"maxDiscount" bson:"maxDiscount"`
	MinSpend       float64            `json:"minSpend" bson:"minSpend"`
	StartDate      string             `json:"startDate" bson:"startDate"`
	EndDate        string             `json:"endDate" bson:"endDate"`
	UsageLimit     int                `json:"usageLimit" bson:"usageLimit"`
	UsageCount     int                `json:"usageCount" bson:"usageCount"`
	IsDeleted      bool               `json:"isDeleted" bson:"isDeleted"`
}

func createVoucher(w http.ResponseWriter, r *http.Request, client *mongo.Client) (*Voucher, error) {
	var voucher Voucher

	err := json.NewDecoder(r.Body).Decode(&voucher)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}

	collection := client.Database("testMongo").Collection("Voucher")
	insertResult, err := collection.InsertOne(context.TODO(), voucher)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	// Assign the InsertedID to the ID field of the Voucher
	voucher.ID = insertResult.InsertedID.(primitive.ObjectID)

	return &voucher, nil
}

func getAllVoucher(w http.ResponseWriter, r *http.Request, client *mongo.Client) ([]Voucher, error) {
	collection := client.Database("testMongo").Collection("Voucher")

	cursor, err := collection.Find(context.TODO(), bson.M{})

	if err != nil {
		fmt.Println("Error in cursor")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	defer cursor.Close(context.Background())

	fmt.Println("Vouchers found")

	var vouchers []Voucher
	if err = cursor.All(context.Background(), &vouchers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	fmt.Println(vouchers)

	return vouchers, nil
}
