package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/toduluz/savingsquadsbackend/internal/data"
	"github.com/toduluz/savingsquadsbackend/internal/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// Create an anonymous struct to hold the expected data from the request body.
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Parse the request body into the anonymous struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the data from the request body into a new User struct. Notice also that
	// we set the Activated field to false, which isn't strictly necessary because
	// the Activated field will have the zero-value of false by default. But setting
	// this explicitly helps to make our intentions clear to anyone reading the code.
	user := &data.User{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      input.Name,
		Email:     input.Email,
		Addresses: []data.Address{},
		Phone:     []data.Phone{},
		Vouchers:  []string{},
		Points:    0,
		Version:   1,
	}

	// Use the Password.Set() method to generate and store the hashed and plaintext
	// passwords.
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	// Validate the user struct and return the error messages to the client if
	// any of the checks fail.
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the user data into the database.
	id, err := app.models.Users.Insert(user)
	if err != nil {
		switch {
		// If we get an ErrDuplicateEmail error, use the v.AddError() method to manually add
		// a message to the validator instance, and then call our failedValidationResponse
		// helper().
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	user.ID = id

	jwtBytes, err := app.createJWTClaims(user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Note that we also change this to send the client a 202 Accepted status code which
	// indicates that the request has been accepted for processing, but the processing has
	// not been completed.
	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user, "authentication_token": string(jwtBytes)}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	jwtBytes, err := app.createJWTClaims(user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Convert the []byte slice to a string and return it in a JSON response.
	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": string(jwtBytes)}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getUserVouchersHandler(w http.ResponseWriter, r *http.Request) {
	idStr := app.readIDParam(r)
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	vouchers, err := app.models.Users.GetAllVouchers(id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	vouchersWithDetails, newVoucherList, err := app.models.Vouchers.GetVoucherList(vouchers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Users.UpdateVouchersList(id, newVoucherList)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"vouchers": vouchersWithDetails}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) addPointsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := app.readIDParam(r)
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var input struct {
		Points int `json:"points"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.models.Users.AddPoints(id, input.Points)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "points added"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) addVoucherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := app.readIDParam(r)
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var input struct {
		Code string `json:"code"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	voucher, err := app.models.Vouchers.Get(input.Code)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Users.AddVoucher(id, input.Code)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"voucher": voucher}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
