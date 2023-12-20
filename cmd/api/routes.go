package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

// routes is our main application's router.
func (app *application) routes() http.Handler {
	router := mux.NewRouter()
	router.Use(app.recoverPanic)
	router.Use(app.enableCORS)
	router.Use(app.authenticate)

	router.NotFoundHandler = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowedHandler = http.HandlerFunc(app.methodNotAllowedResponse)

	// Admin routes
	/*
		router.HandleFunc("/v1/vouchers", app.requireAuthenticatedUser(app.listVouchersHandler)).Methods(http.MethodGet)
		router.HandleFunc("/v1/vouchers", app.requireAuthenticatedUser(app.createVoucherHandler)).Methods(http.MethodPost)
		router.HandleFunc("/v1/vouchers/{id}", app.requireAuthenticatedUser(app.showVoucherHandler)).Methods(http.MethodGet)
		router.HandleFunc("/v1/vouchers/{id}", app.requireAuthenticatedUser(app.deleteVoucherHandler)).Methods(http.MethodDelete)
		router.HandleFunc("/v1/vouchers/{id}/use", app.requireAuthenticatedUser(app.updateVoucherUsageCountHandler)).Methods(http.MethodPut)
	*/

	// User routes
	router.HandleFunc("/v1/users/register", app.registerUserHandler).Methods(http.MethodPost)
	router.HandleFunc("/v1/users/login", app.loginUserHandler).Methods(http.MethodPost)
	router.HandleFunc("/v1/users/vouchers", app.requireAuthenticatedUser(app.getUserVouchersHandler)).Methods(http.MethodGet)
	router.HandleFunc("/v1/users/vouchers", app.requireAuthenticatedUser(app.redeemUserVoucherHandler)).Methods(http.MethodPut)
	router.HandleFunc("/v1/users/vouchers/use", app.requireAuthenticatedUser(app.useUserVoucherHandler)).Methods(http.MethodPut)
	router.HandleFunc("/v1/users/points", app.requireAuthenticatedUser(app.getUserPointsHandler)).Methods(http.MethodGet)
	router.HandleFunc("/v1/users/points", app.requireAuthenticatedUser(app.addUserPointsHandler)).Methods(http.MethodPut)
	router.HandleFunc("/v1/users/points/redeem", app.requireAuthenticatedUser(app.redeemPointsForVoucherHandler)).Methods(http.MethodPut)
	//TODO: router.HandleFunc("/v1/users/vouchers/best", app.requireAuthenticatedUser(app.getUserBestVoucherHandler)).Methods(http.MethodGet)

	return router
}
