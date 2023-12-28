package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/toduluz/savingsquadsbackend/internal/data"
	"github.com/toduluz/savingsquadsbackend/internal/validator"
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
		Vouchers:  map[string]int{},
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

	err = app.setCookie(w, "jwt", string(jwtBytes), 3600*24)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Note that we also change this to send the client a 202 Accepted status code which
	// indicates that the request has been accepted for processing, but the processing has
	// not been completed.
	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
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

	err = app.setCookie(w, "jwt", string(jwtBytes), 3600*24)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Convert the []byte slice to a string and return it in a JSON response.
	err = app.writeJSON(w, http.StatusCreated, envelope{"message": "successfully logged in"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) logoutUserHandler(w http.ResponseWriter, r *http.Request) {
	// Set the value of the "jwt" cookie to the empty string.
	err := app.setCookie(w, "jwt", "", -1)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Send a 200 OK status code and a JSON response with the message "you have been
	// logged out successfully".
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "successfully logged out"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getUserVouchersHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the user from the request context.
	user := app.contextGetUser(r)

	var voucherCodes []string
	for code := range user.Vouchers {
		voucherCodes = append(voucherCodes, code)
	}

	var vouchersWithDetails []data.Voucher
	if len(voucherCodes) > 0 {

		// Get the details of the vouchers
		vouchersfromModel, err := app.models.Vouchers.GetVoucherList(voucherCodes)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
		vouchersWithDetails = vouchersfromModel
	}

	var vouchersWithDetailsAndCount []struct {
		Code               string    `json:"code"`
		Description        string    `json:"description"`
		Discount           int       `json:"discount"`
		IsPercentage       bool      `json:"isPercentage"`
		Starts             time.Time `json:"starts"`
		Expires            time.Time `json:"expires"`
		Active             bool      `json:"active"`
		UserUsageRemaining int       `json:"userUsageRemaining"`
		MinSpend           int       `json:"minSpend"`
		Category           string    `json:"category"`
	}

	for _, voucher := range vouchersWithDetails {
		if !voucher.Active {
			delete(user.Vouchers, voucher.Code)
		} else {
			vouchersWithDetailsAndCount = append(vouchersWithDetailsAndCount, struct {
				Code               string    `json:"code"`
				Description        string    `json:"description"`
				Discount           int       `json:"discount"`
				IsPercentage       bool      `json:"isPercentage"`
				Starts             time.Time `json:"starts"`
				Expires            time.Time `json:"expires"`
				Active             bool      `json:"active"`
				UserUsageRemaining int       `json:"userUsageRemaining"`
				MinSpend           int       `json:"minSpend"`
				Category           string    `json:"category"`
			}{
				Code:               voucher.Code,
				Description:        voucher.Description,
				Discount:           voucher.Discount,
				IsPercentage:       voucher.IsPercentage,
				Starts:             voucher.Starts,
				Expires:            voucher.Expires,
				Active:             voucher.Active,
				UserUsageRemaining: user.Vouchers[voucher.Code],
				MinSpend:           voucher.MinSpend,
				Category:           voucher.Category,
			})

		}
	}

	// Update the user's voucher list
	err := app.models.Users.UpdateVoucherList(user.ID, user.Vouchers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send the data in a JSON response.
	err = app.writeJSON(w, http.StatusOK, envelope{"vouchers": vouchersWithDetailsAndCount}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) getUserPointsHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	points, err := app.models.Users.GetPoints(user.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"points": points}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) addUserPointsHandler(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)

	var input struct {
		Points int `json:"points"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.models.Users.AddPoints(user.ID, input.Points)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "points added"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) exchangePointsForVoucherHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var input struct {
		Points       int    `json:"points"`
		Description  string `json:"description"`
		Discount     int    `json:"discount"`
		IsPercentage bool   `json:"isPercentage"`
		Duration     int    `json:"duration"`
		Category     string `json:"category"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	voucher := data.Voucher{
		Description:  input.Description,
		Discount:     input.Discount,
		IsPercentage: input.IsPercentage,
		Starts:       time.Now(),
		Expires:      time.Now().Add(time.Duration(input.Duration) * time.Hour),
		Active:       true,
		UsageLimit:   1,
		UsageCount:   0,
		MinSpend:     0,
		Category:     input.Category,
	}

	err = voucher.VocuherCodeGenerator()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	validator := validator.New()
	data.ValidateVoucher(validator, &voucher)
	data.ValidatePoints(validator, input.Points)
	if !validator.Valid() {
		app.failedValidationResponse(w, r, validator.Errors)
		return
	}

	err = app.models.Users.DeductPointsAndCreateVoucher(user.ID, input.Points, &voucher)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrVoucherAlreadyExists):
			app.voucherAlreadyExistResponse(w, r)
		case errors.Is(err, data.ErrExchangePointsForVoucher):
			app.problemExchangePointsForVoucherResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Send a success response
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "points redeemed and voucher added"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) redeemUserVoucherHandler(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)

	code := app.readIDParam(r)

	voucher, err := app.models.Vouchers.Get(code)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if !voucher.Active {
		app.voucherNotAvailableResponse(w, r)
		return
	}

	err = app.models.Users.RedeemVoucher(user.ID, voucher.Code, 1)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrVoucherAlreadyRedeeemed):
			app.voucherAlreadyRedeemedResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"voucher": "successfully redeemed voucher"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) useUserVoucherHandler(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)

	voucherCode := app.readIDParam(r)

	if voucherCount, ok := user.Vouchers[voucherCode]; ok {
		if voucherCount > 0 {
			user.Vouchers[voucherCode]--
			err := app.models.Users.UpdateVoucherList(user.ID, user.Vouchers)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}
			err = app.models.Vouchers.UpdateUsageCount(voucherCode)
			if err != nil {
				switch {
				case errors.Is(err, data.ErrRecordNotFound):
					app.notFoundResponse(w, r)
				case errors.Is(err, data.ErrEditConflict):
					app.editConflictResponse(w, r)
				default:
					app.serverErrorResponse(w, r, err)
				}
				return
			}
		} else {
			app.voucherNotAvailableResponse(w, r)
			return
		}
	} else {
		app.voucherNotAvailableResponse(w, r)
		return
	}

	err := app.writeJSON(w, http.StatusOK, envelope{"voucher": "successfully used voucher"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
