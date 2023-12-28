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

	router.NotFoundHandler = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowedHandler = http.HandlerFunc(app.methodNotAllowedResponse)

	// Public routes
	publicRouter := router.PathPrefix("/v1/user").Subrouter()
	publicRouter.HandleFunc("/register", app.registerUserHandler).Methods(http.MethodPost)
	publicRouter.HandleFunc("/login", app.loginUserHandler).Methods(http.MethodPost)

	// Authenticated routes
	authRouter := router.PathPrefix("/v1").Subrouter()
	authRouter.Use(app.authenticate)
	authRouter.Use(app.requireAuthenticatedUser)

	// Admin routes
	adminRouter := authRouter.PathPrefix("/voucher").Subrouter()
	adminRouter.HandleFunc("", app.listVouchersHandler).Methods(http.MethodGet)
	adminRouter.HandleFunc("", app.createVoucherHandler).Methods(http.MethodPost)
	adminRouter.HandleFunc("/{id}", app.showVoucherHandler).Methods(http.MethodGet)
	adminRouter.HandleFunc("/{id}", app.deleteVoucherHandler).Methods(http.MethodDelete)

	// User routes
	userRouter := authRouter.PathPrefix("/user").Subrouter()
	userRouter.HandleFunc("/logout", app.logoutUserHandler).Methods(http.MethodPost)
	userRouter.HandleFunc("/voucher", app.getUserVouchersHandler).Methods(http.MethodGet)
	userRouter.HandleFunc("/voucher/{id}/redeem", app.redeemUserVoucherHandler).Methods(http.MethodPut)
	userRouter.HandleFunc("/voucher/{id}/use", app.useUserVoucherHandler).Methods(http.MethodPut)
	userRouter.HandleFunc("/point", app.getUserPointsHandler).Methods(http.MethodGet)
	userRouter.HandleFunc("/point", app.addUserPointsHandler).Methods(http.MethodPut)
	userRouter.HandleFunc("/point/exchange", app.exchangePointsForVoucherHandler).Methods(http.MethodPost)
	// TODO: userRouter.HandleFunc("/vouchers/best", app.getUserBestVoucherHandler).Methods(http.MethodGet)

	return router
}
