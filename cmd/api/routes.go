package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

// routes is our main application's router.
func (app *application) routes() http.Handler {
	// router := httprouter.New()

	// // Convert the app.notFoundResponse helper to a http.Handler using the http.HandlerFunc()
	// // adapter, and then set it as the custom error handler for 404 Not Found responses.
	// router.NotFound = http.HandlerFunc(app.notFoundResponse)

	// // Convert app.methodNotAllowedResponse helper to a http.Handler and set it as the custom
	// // error handler for 405 Method Not Allowed responses
	// router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// // Vouchers handlers. Note, that these movie endpoints use the `requireActivatedUser` middleware.
	// router.HandlerFunc(http.MethodGet, "/v1/vouchers", app.listVouchersHandler)
	// router.HandlerFunc(http.MethodPost, "/v1/vouchers", app.createVoucherHandler)
	// router.HandlerFunc(http.MethodGet, "/v1/vouchers/:id", app.showVoucherHandler)
	// router.HandlerFunc(http.MethodDelete, "/v1/vouchers/:id", app.deleteVoucherHandler)
	// router.HandlerFunc(http.MethodPut, "/v1/vouchers/:id/use", app.updateVoucherUsageCountHandler)

	// // Users handlers
	// router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)

	// // Tokens handlers
	// router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	// // Wrap the router with the panic recovery middleware and rate limit middleware.
	// return app.recoverPanic(app.enableCORS(app.authenticate(router)))
	router := mux.NewRouter()
	router.Use(app.recoverPanic)
	router.Use(app.enableCORS)
	router.Use(app.authenticate)

	router.NotFoundHandler = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowedHandler = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandleFunc("/v1/vouchers", app.listVouchersHandler).Methods(http.MethodGet)
	router.HandleFunc("/v1/vouchers", app.createVoucherHandler).Methods(http.MethodPost)
	router.HandleFunc("/v1/vouchers/{id}", app.showVoucherHandler).Methods(http.MethodGet)
	router.HandleFunc("/v1/vouchers/{id}", app.deleteVoucherHandler).Methods(http.MethodDelete)
	router.HandleFunc("/v1/vouchers/{id}/use", app.updateVoucherUsageCountHandler).Methods(http.MethodPut)

	router.HandleFunc("/v1/users", app.registerUserHandler).Methods(http.MethodPost)

	router.HandleFunc("/v1/tokens/authentication", app.createAuthenticationTokenHandler).Methods(http.MethodPost)

	return router
}
