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

func updateVoucherUsageByID(w http.ResponseWriter, r *http.Request, client *mongo.Client) (*Voucher, error) {

	// find voucher
	var voucher Voucher
	temp := r.FormValue("query")
	id, err := primitive.ObjectIDFromHex(temp)

	//err := json.NewDecoder(r.Body).Decode(&voucher)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}

	collection := client.Database("testMongo").Collection("Voucher")

	//	err = collection.FindOne(context.Background(), bson.M{"_id": voucher.ID}).Decode(&voucher)
	err = collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&voucher)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "No voucher found with given ID", http.StatusNotFound)
			return nil, fmt.Errorf("no voucher found with ID: %v", voucher.ID)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	res, err := collection.UpdateOne(context.Background(), bson.M{"_id": voucher.ID}, bson.M{"$set": bson.M{"usageCount": voucher.UsageCount + 1}})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	if res.ModifiedCount == 0 {
		http.Error(w, "No voucher found with given ID", http.StatusNotFound)
		return nil, fmt.Errorf("no voucher found with ID: %v", voucher.ID)
	}

	fmt.Println("Voucher updated")
	return &voucher, nil
}

func getVoucherById(w http.ResponseWriter, r *http.Request, client *mongo.Client) (Voucher, error) {
	collection := client.Database("testMongo").Collection("Voucher")

	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return Voucher{}, err
	}

	var voucher Voucher
	err = collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&voucher)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return Voucher{}, err
	}

	return voucher, nil
}

func deleteVoucherById(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	collection := client.Database("testMongo").Collection("Voucher")

	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}
