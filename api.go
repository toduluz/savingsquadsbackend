package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

type APIServer struct {
	listenAddr string
	store      *mongo.Client
}

func NewAPIServer(listenAddr string, store *mongo.Client) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/voucher", makeHTTPHandleFunc(s.handleVoucher))
	//retrieve info from mongodb

	//router.HandleFunc("/", makeHTTPHandleFunc(s.handleVoucher))

	// '/voucher' 			GET 	Voucher
	//						POST 	createVoucher
	//						PUT	  	updateVoucherByID

	// '/voucher/{id}' 		GET Voucher by ID / DELETE Voucher by ID

	// router.HandleFunc("/login", makeHTTPHandleFunc(s.handleLogin))
	// router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))
	// router.HandleFunc("/account/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleGetAccountByID), s.store))
	// router.HandleFunc("/transfer", makeHTTPHandleFunc(s.handleTransfer))

	log.Println("JSON API server running on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleVoucher(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		fmt.Println("Hello from handle POST /")

		voucher, err := createVoucher(w, r, s.store)
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, nil)
		}
		return WriteJSON(w, http.StatusCreated, voucher)
	}

	if r.Method == "GET" {
		fmt.Println("Hello from handle GET /")
		vouchers, err := getAllVoucher(w, r, s.store)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusOK, vouchers)
	}

	if r.Method == "PUT" {
		fmt.Println("Hello from handle PUT /")
		voucher, err := updateVoucher(w, r, s.store)
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
