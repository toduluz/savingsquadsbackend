package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

// Defined API Server that listen to port 3000 and has connection to the MongoDB Client
type APIServer struct {
	listenAddr string
	client     *mongo.Client
}

func NewAPIServer(listenAddr string, client *mongo.Client) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		client:     client,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()
	//	test := http.NewServeMux()

	// Endpoints				Method 		Function				Description
	// '/voucher' 				GET 		getAllVoucher() 		- return all vouchers
	//							POST 		createVoucher() 		- create new voucher
	router.HandleFunc("/voucher", makeHTTPHandleFunc(s.handleVoucher))

	// Endpoints				Method 		Function				Description
	// '/voucher/{id}' 			GET 		getVoucherById() 		- return voucher by ID
	// 							PUT	  		updateVoucherByID() 	- update voucher to isDeleted by ID
	router.HandleFunc("/voucher/{id}", makeHTTPHandleFunc(s.handleVoucherById))

	// Endpoints				Method 		Function				Description
	// '/voucher/{id}/delete' 	PUT		 	handleUpdateVoucher()	- update voucher (isDeleted = true)
	//router.HandleFunc("/voucher/{id}/delete", makeHTTPHandleFunc(s.handleDeleteVoucher)) // commented out for now

	// Endpoints				Method 		Function				Description
	// '/voucher/{id}/usage' 	PUT		 	handleUpdateVoucher()	- update voucher (usageCount + 1)
	router.HandleFunc("/voucher/{id}/usage", makeHTTPHandleFunc(s.handleUpdateVoucherUsage))

	// Endpoints				Method 		Function				Description
	// '/user' 					GET 		getAllUser() 			- return all users
	//							POST 		createUser() 			- create new user
	router.HandleFunc("/user", makeHTTPHandleFunc(s.handleUser))

	// Endpoints				Method 		Function				Description
	// '/user/{id}' 			GET 		getUserById() 			- return user by ID
	router.HandleFunc("/user/{id}", makeHTTPHandleFunc(s.handleUserById))

	// Endpoints				Method 		Function				Description
	// '/voucher/{id}/usageLimit' 	PUT		 	handleUpdateVoucher()	- update voucher limit (usageLimit + 10)
	router.HandleFunc("/voucher/{id}/usageLimit", makeHTTPHandleFunc(s.handleVoucherUsageLimit))

	log.Println("JSON API server running on port:", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		fmt.Println("Hello from handleUser() GET /")
		users, err := getAllUser(w, r, s.client)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusOK, users)
	}

	if r.Method == "POST" {
		fmt.Println("Hello from handleUser() POST /")
		userId, err := createUser(w, r, s.client)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusCreated, "id: "+userId)
	}

	return WriteJSON(w, http.StatusBadRequest, nil)
}

func (s *APIServer) handleUserById(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		fmt.Println("Hello from handleUserById GET /user/{id}")
		user, err := getUserById(w, r, s.client)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusOK, user)
	}
	return WriteJSON(w, http.StatusBadRequest, nil)
}

// Endpoints				Method 		Function				Description
// '/voucher' 				GET 		getAllVoucher() 		- return all vouchers
//							POST 		createVoucher() 		- create new voucher
//							PUT	  		updateVoucherByID() 	- update voucher by ID xx
//router.HandleFunc("/voucher", makeHTTPHandleFunc(s.handleVoucher))

func (s *APIServer) handleVoucher(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		fmt.Println("Hello from handleVoucher() GET /")
		vouchers, err := getAllVoucher(w, r, s.client)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusOK, vouchers)
	}

	if r.Method == "POST" {
		fmt.Println("Hello from handleVoucher() POST /")
		voucherId, err := createVoucher(w, r, s.client)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusCreated, "id: "+voucherId)
	}
	return WriteJSON(w, http.StatusBadRequest, nil)
}

// Endpoints				Method 		Function				Description
// '/voucher/{id}' 			GET 		getVoucherById() 		- return voucher by ID
//							PUT	  		updateVoucherByID() 	- update voucher by ID xx - to be updated
//
// router.HandleFunc("/voucher/{id}", makeHTTPHandleFunc(s.handleVoucherById))

// this function returns voucher by ID
func (s *APIServer) handleVoucherById(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		fmt.Println("Hello from handleVoucherById GET /voucher/{id}")
		voucher, err := getVoucherById(w, r, s.client)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusOK, voucher)
	}
	// should be PUT to update voucher isDelete = True
	// if r.Method == "DELETE" {
	// 	fmt.Println("Hello from handleVoucherById DELETE /voucher/{id}")
	// 	err := deleteVoucherById(w, r, s.client)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return WriteJSON(w, http.StatusOK, nil)
	// }

	// this updates the isDeleted field to true
	// this stimulates the deletion of the voucher from the database

	// ######## the returned voucher here is the voucher before the update ########
	// ### take note ###
	if r.Method == "PUT" {
		fmt.Println("Hello from handleVoucherById PUT /voucher/{id} here")
		voucher, err := updateVoucherIsDeletedByID(w, r, s.client)
		if err != nil {
			return err
		}
		fmt.Println("voucher is deleted")
		return WriteJSON(w, http.StatusOK, voucher)
	}
	return WriteJSON(w, http.StatusBadRequest, nil)
}

func (s *APIServer) handleUpdateVoucherUsage(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "PUT" {
		fmt.Println("Hello from handle PUT /voucher/{id}/usage")
		voucher, err := updateVoucherUsageByID(w, r, s.client)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusOK, voucher)
	}

	return WriteJSON(w, http.StatusBadRequest, nil)
}

func (s *APIServer) handleVoucherUsageLimit(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "PUT" {
		fmt.Println("Hello from handle PUT /voucher/{id}/usageLimit")
		voucher, err := updateVoucherUsageLimitByID(w, r, s.client)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusOK, voucher)
	}

	return WriteJSON(w, http.StatusBadRequest, nil)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			// handle error
			WriteJSON(w, http.StatusBadRequest, nil)
		}
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
