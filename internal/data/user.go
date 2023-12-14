package data

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	ID       string            `json:"id,omitempty" bson:"_id,omitempty"`
	Username string            `json:"username" bson:"username"`
	Password string            `json:"password" bson:"password"`
	Email    string            `json:"email" bson:"email"`
	Contact  int               `json:"contact" bson:"contact"`
	Address  map[string]string `json:"address" bson:"address"`
	Cashback float64           `json:"cashback" bson:"cashback"`
	DateJoin string            `json:"dateJoin" bson:"dateJoin"`
	IsAdmin  bool              `json:"isAdmin" bson:"isAdmin"`
	Vouchers []string          `json:"vouchers" bson:"vouchers"`
}

func CreateUser(w http.ResponseWriter, r *http.Request, client *mongo.Client) (string, error) {
	var user User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return "", err
	}

	collection := client.Database("testMongo").Collection("User")
	insertResult, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", err
	}

	// Convert the InsertedID to a string and return
	return insertResult.InsertedID.(primitive.ObjectID).Hex(), nil
}

func GetAllUser(w http.ResponseWriter, r *http.Request, client *mongo.Client) ([]User, error) {
	collection := client.Database("testMongo").Collection("User")

	cursor, err := collection.Find(context.TODO(), bson.M{})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var users []User
	if err = cursor.All(context.Background(), &users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	return users, nil
}

func GetUserById(w http.ResponseWriter, r *http.Request, client *mongo.Client) (*User, error) {

	// get id from url
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	// find user
	var user User
	collection := client.Database("testMongo").Collection("User")
	err = collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	return &user, nil
}
