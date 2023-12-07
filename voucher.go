package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
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

// update usageCount and IsDeleted

func createVoucher(w http.ResponseWriter, r *http.Request, client *mongo.Client) (string, error) {
	var voucher Voucher

	err := json.NewDecoder(r.Body).Decode(&voucher)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return "", err
	}

	collection := client.Database("testMongo").Collection("Voucher")
	insertResult, err := collection.InsertOne(context.TODO(), voucher)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}

	// Convert the InsertedID to a string and return
	return insertResult.InsertedID.(primitive.ObjectID).Hex(), nil
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

func updateVoucherIsDeletedByID(w http.ResponseWriter, r *http.Request, client *mongo.Client) (*Voucher, error) {
	voucher, err := getVoucherById(w, r, client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}

	collection := client.Database("testMongo").Collection("Voucher")

	res, err := collection.UpdateOne(context.Background(), bson.M{"_id": voucher.ID}, bson.M{"$set": bson.M{"isDeleted": false}})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	if res.ModifiedCount == 0 {
		http.Error(w, "No updates", http.StatusNotFound)
		return nil, fmt.Errorf("no voucher found with ID: %v", voucher.ID)
	}

	fmt.Println("Voucher deleted (updated isDeleted field to true)")
	return voucher, nil
}

// update usage count and check if usage limit is reached
// if usage limit is reached, no more further vouchers can be used
func updateVoucherUsageByID(w http.ResponseWriter, r *http.Request, client *mongo.Client) (*Voucher, error) {

	voucher, err := getVoucherById(w, r, client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}
	if voucher.UsageCount > voucher.UsageLimit {
		fmt.Println("Usage limit reached")
		return nil, err
	}
	collection := client.Database("testMongo").Collection("Voucher")

	res, err := collection.UpdateOne(context.Background(), bson.M{"_id": voucher.ID}, bson.M{"$set": bson.M{"usageCount": voucher.UsageCount + 1}})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	if res.ModifiedCount == 0 {
		http.Error(w, "No voucher found with given ID", http.StatusNotFound)
		return nil, fmt.Errorf("no voucher found with ID: %v", voucher.ID)
	}

	if voucher.UsageCount+1 > voucher.UsageLimit {
		updateVoucherIsDeletedByID(w, r, client)
	}
	fmt.Println("Current usage count: ", voucher.UsageCount+1)
	fmt.Println("Usage limit: ", voucher.UsageLimit)
	fmt.Println("Voucher updated")
	return voucher, nil
}

func updateVoucherUsageLimitByID(w http.ResponseWriter, r *http.Request, client *mongo.Client) (*Voucher, error) {
	voucher, err := getVoucherById(w, r, client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}

	collection := client.Database("testMongo").Collection("Voucher")

	res, err := collection.UpdateOne(context.Background(), bson.M{"_id": voucher.ID}, bson.M{"$set": bson.M{"usageLimit": voucher.UsageLimit + 10}})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	if res.ModifiedCount == 0 {
		http.Error(w, "No voucher found with given ID", http.StatusNotFound)
		return nil, fmt.Errorf("no voucher found with ID: %v", voucher.ID)
	}
	return voucher, nil
}

func getVoucherById(w http.ResponseWriter, r *http.Request, client *mongo.Client) (*Voucher, error) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}

	var voucher Voucher
	collection := client.Database("testMongo").Collection("Voucher")
	err = collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&voucher)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	return &voucher, nil
}

// func deleteVoucherById(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
// 	collection := client.Database("testMongo").Collection("Voucher")

// 	vars := mux.Vars(r)
// 	id, err := primitive.ObjectIDFromHex(vars["id"])
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return err
// 	}

// 	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": id})
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return err
// 	}

// 	return nil
// }
