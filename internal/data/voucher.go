package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	db                = "testMongo"
	voucherCollection = "Voucher"
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

// creating brand new voucher POST method
func CreateVoucher(w http.ResponseWriter, r *http.Request, client *mongo.Client) (string, error) {
	var voucher Voucher

	err := json.NewDecoder(r.Body).Decode(&voucher)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return "", err
	}

	collection := client.Database(db).Collection(voucherCollection)
	insertResult, err := collection.InsertOne(context.TODO(), voucher)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}
	log.Println("Inserted a single document: ", insertResult.InsertedID, "New vouchers created")
	// Convert the InsertedID to a string and return
	return insertResult.InsertedID.(primitive.ObjectID).Hex(), nil
}

// retrieves all vouchers GET method
func GetAllVoucher(w http.ResponseWriter, r *http.Request, client *mongo.Client) ([]Voucher, error) {
	log.Println("Retrieving all vouchers...")
	collection := client.Database(db).Collection(voucherCollection)
	cursor, err := collection.Find(context.TODO(), bson.M{})

	if err != nil {
		fmt.Println("Error in cursor")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	defer cursor.Close(context.Background())

	log.Println("Retrieved all vouchers")

	var vouchers []Voucher
	if err = cursor.All(context.Background(), &vouchers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	log.Println("Returning all vouchers...")

	return vouchers, nil
}

// update Voucher isDeleted field to true PUT method
func UpdateVoucherIsDeletedByID(w http.ResponseWriter, r *http.Request, client *mongo.Client) (*Voucher, error) {

	var voucher *Voucher
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	log.Println("ID retrieved...")
	if err != nil {
		log.Println("Error in getting ID")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}
	collection := client.Database(db).Collection(voucherCollection)
	err = collection.FindOneAndUpdate(context.Background(), bson.M{"_id": id}, bson.M{"$set": bson.M{"isDeleted": true}}).Decode(&voucher)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Error in updating voucher isDeleted field")
		return nil, err
	}

	log.Println("Voucher deleted (updated isDeleted field to true)")
	return voucher, nil
}

// update usage count and check if usage limit is reached
// if usage limit is reached, no more further vouchers can be used
func UpdateVoucherUsageByID(w http.ResponseWriter, r *http.Request, client *mongo.Client) (*Voucher, error) {
	log.Println("In updateVoucherUsageByID")
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	log.Println("ID retrieved...")
	if err != nil {
		log.Println("Error in getting ID")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}

	var voucher *Voucher
	collection := client.Database(db).Collection(voucherCollection)

	filter := bson.M{"_id": id, "$expr": bson.M{"$lt": bson.A{"$usageCount", "$usageLimit"}}}
	update := bson.M{
		"$inc": bson.M{"usageCount": 1},
	}
	err = collection.FindOneAndUpdate(context.Background(), filter, update).Decode(&voucher)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("No document found based on current filters")
			return nil, errors.New("no voucher found with given ID")
		} else {
			log.Println("Error in updating voucher usage")
		}
	}
	log.Println("Voucher updated", voucher)
	log.Println("Current usage count here: ", voucher.UsageCount+1)

	if voucher.UsageCount+1 >= voucher.UsageLimit {
		UpdateVoucherIsDeletedByID(w, r, client)
	}
	log.Println("Current usage count: ", voucher.UsageCount+1)
	log.Println("Usage limit: ", voucher.UsageLimit)
	log.Println("Voucher updated")
	return voucher, nil
}

// updates voucher limit by 10
func UpdateVoucherUsageLimitByID(w http.ResponseWriter, r *http.Request, client *mongo.Client) (*Voucher, error) {
	// voucher, err := getVoucherById(w, r, client)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return nil, err
	// }
	var voucher *Voucher
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	log.Println("ID retrieved...")
	if err != nil {
		log.Println("Error in getting ID")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}
	collection := client.Database(db).Collection(voucherCollection)

	err = collection.FindOneAndUpdate(context.Background(), bson.M{"_id": id}, bson.M{"$inc": bson.M{"usageLimit": 10}}).Decode(&voucher)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	log.Println("Voucher usage limit updated. Current limit: ", voucher.UsageLimit+10)
	return voucher, nil
}

// returns 1 voucher filtered using ID
func GetVoucherById(w http.ResponseWriter, r *http.Request, client *mongo.Client) (*Voucher, error) {
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, err
	}

	var voucher Voucher
	collection := client.Database(db).Collection(voucherCollection)
	err = collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&voucher)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	return &voucher, nil
}

// func deleteVoucherById(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
// 	collection := client.Database(db).Collection(voucherCollection)

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
