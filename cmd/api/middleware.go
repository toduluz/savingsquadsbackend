package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/toduluz/savingsquadsbackend/internal/cookies"
	"github.com/toduluz/savingsquadsbackend/internal/data"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// recoverPanic is middleware that recovers from a panic by responding with a 500 Internal Server
// Error before closing the connection. It will also log the error using our custom Logger at
// the ERROR level.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event of a panic as
		// Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a panic or not.
			if err := recover(); err != nil {
				// If there was a panic, set a "Connection: close" header on the response. This
				// acts a trigger to make Go's HTTP server automatically close the current
				// connection after a response has been sent.
				w.Header().Set("Connection:", "close")
				// The value returned by recover() has the type interface{}, so we use
				// fmt.Errorf() to normalize it into an error and call our
				// serverErrorResponse() helper. In turn, this will log the error using our
				// custom Logger type at the ERROR level and send the client a
				// 500 Internal Server Error response.
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the contextGetUser() helper that we made earlier to retrieve the user
		// information from the request context.
		user := app.contextGetUser(r)
		// If the user is anonymous, then call the authenticationRequiredResponse() to
		// inform the client that they should authenticate before trying again.
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}
		// // If the user is not activated, use the inactiveAccountResponse() helper to
		// // inform them that they need to activate their account.
		// if !user.Activated {
		// 	app.inactiveAccountResponse(w, r)
		// 	return
		// }
		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, err := app.getCookie(r, "jwt")
		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				r = app.contextSetUser(r, data.AnonymousUser)
				next.ServeHTTP(w, r)
			case errors.Is(err, cookies.ErrInvalidValue):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.logger.PrintError(err, nil)
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// w.Header().Add("Vary", "Authorization")
		// authorizationHeader := r.Header.Get("Authorization")
		// if authorizationHeader == "" {
		// 	r = app.contextSetUser(r, data.AnonymousUser)
		// 	next.ServeHTTP(w, r)
		// 	return
		// }
		// headerParts := strings.Split(authorizationHeader, " ")
		// if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		// 	app.invalidAuthenticationTokenResponse(w, r)
		// 	return
		// }
		// token := headerParts[1]

		claims, err := app.validateJWTClaims(token)
		if err != nil {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		// // Check that the issuer is our application.
		// if claims.Issuer != "" {
		// 	app.invalidAuthenticationTokenResponse(w, r)
		// 	return
		// }
		// // Check that our application is in the expected audiences for the JWT.
		// if !claims.AcceptAudience("") {
		// 	app.invalidAuthenticationTokenResponse(w, r)
		// 	return
		// }
		id, err := primitive.ObjectIDFromHex(claims.Subject)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		// Lookup the user record from the database.
		user, err := app.models.Users.Get(id)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
		// Add the user record to the request context and continue as normal.
		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

// enableCORS sets the Vary: Origin and Access-Control-Allow-Origin response headers in order to
// enabled CORS for trusted origins.
func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the "Vary: Origin" header.
		w.Header().Set("Vary", "Origin")

		// Add the "Vary: Access-Control-Request-Method" header.
		w.Header().Set("Vary", "Access-Control-Request-Method")

		// Get the value of the request's Origin header.
		origin := r.Header.Get("Origin")

		// On run this if there's an Origin request header present.
		if origin != "" {
			// Loop through the list of trusted origins, checking to see if the request
			// origin exactly matches one of them. If there are no trusted origins, then the
			// loop won't be iterated.
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					// If there is a match, then set an "Access-Control-Allow-Origin" response
					// header with the request origin as the value and break out of the loop.
					w.Header().Set("Access-Control-Allow-Origin", origin)

					// Check if the request has the HTTP method OPTIONS and contains the
					// "Access-Control-Request-Method" header. If it does, then we treat it as a
					// preflight request.
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						// Set the necessary preflight response headers.
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						// Set max cached times for headers for 60 seconds.
						w.Header().Set("Access-Control-Max-Age", "60")

						// Write the headers along with a 200 OK status and return from the
						// middleware with no further action.
						w.WriteHeader(http.StatusOK)
						return
					}

					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}
