package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/toduluz/savingsquadsbackend/internal/data"
	"github.com/toduluz/savingsquadsbackend/internal/validator"
)

func (app *Application) createVoucherHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the information that we expect to be in the HTTP
	// request body (not that the field names and types in the struct are a subset of the Movie
	// struct). This struct will be our *target decode destination*.
	var input struct {
		Code         string    `json:"id"`
		Description  string    `json:"description"`
		Discount     int       `json:"discount"`
		IsPercentage bool      `json:"isPercentage"`
		Starts       time.Time `json:"start"`
		Expires      time.Time `json:"expires"`
		UsageLimit   int       `json:"usageLimit,omitempty"`
		MinSpend     int       `json:"minSpend,omitempty"`
		Category     string    `json:"category"`
	}

	// Use the readJSON() helper to decode the request body into the struct.
	// If this returns an error we send the client the error message along with
	// a 400 Bad Request status code.
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var active bool

	if input.Starts.Before(time.Now()) {
		active = true
	} else {
		active = false
	}

	if input.Expires.Before(time.Now()) {
		active = false
	} else {
		active = true
	}

	// Copy the values from the input struct to a new Voucher struct.
	voucher := &data.Voucher{
		Code:         strings.ToLower(input.Code),
		CreatedAt:    time.Now(),
		ModifiedAt:   time.Now(),
		Description:  input.Description,
		Discount:     input.Discount,
		IsPercentage: input.IsPercentage,
		Starts:       input.Starts,
		Expires:      input.Expires,
		Active:       active,
		UsageLimit:   input.UsageLimit,
		UsageCount:   0,
		MinSpend:     input.MinSpend,
		Category:     input.Category,
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call the ValidateVoucher() function and return a response containing the errors if any of
	// the checks fail.
	if data.ValidateVoucher(v, voucher); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the Insert() method on our vouchers model, passing in a pointer to the validated voucher
	// struct. This will create a record in the database and update the voucher struct with the
	// system-generated information.
	err = app.Models.Vouchers.Insert(voucher)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// When sending an HTTP response,
	// we want to include a Location header to let the client know which URL they can find the
	// newly created resource at. We make an empty http.Header map and then use the Set()
	// method to add a new Location header,
	// interpolating the system-generated ID for our new movie in the URL.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/vouchers/%s", voucher.Code))

	// Write a JSON response with a 201 Created status code, the movie data in the response body,
	// and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelope{"voucher": voucher}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// showMovieHandler handles the "GET /v1/vouchers/{id}" endpoint and returns a JSON response of the
// requested voucher record. If there is an error a JSON formatted error is
// returned.
func (app *Application) showVoucherHandler(w http.ResponseWriter, r *http.Request) {
	// When httprouter is parsing a request, any interpolated URL Parameters will be stored
	// in the request context. We can use the ParamsFromContext() function to retrieve a slice
	// containing these parameter names and values.
	code := app.readIDParam(r)

	// Call the Get() method to fetch the data for a specific movie.
	// We also need to use the errors.Is()
	// function to check if it returns a data.ErrRecordNotFound error,
	// in which case we send a 404 Not Found response to the client.
	voucher, err := app.Models.Vouchers.Get(code)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Create an envelope{"movie": movie} instance and pass it to writeJSON(), instead of passing
	// the plain movie struct.
	err = app.writeJSON(w, http.StatusOK, envelope{"voucher": voucher}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// deleteVoucherHandler handles "DELETE /v1/vouchers/{id}" endpoint and returns a 200 OK status code
// with a success message in a JSON response. If there is an error a JSON formatted error is
// returned.
func (app *Application) deleteVoucherHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the movie ID from the URL.
	code := app.readIDParam(r)

	// Delete the voucher from the database. Send a 404 Not Found response to the client if
	// there isn't a matching record.
	err := app.Models.Vouchers.Delete(code)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Return a 200 OK status code along with a success message.
	err = app.writeJSON(w, 200, envelope{"message": "voucher successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *Application) listVouchersHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Code         string
		Starts       time.Time
		Expires      time.Time
		Active       bool
		MinSpend     int
		Category     string
		data.Filters // Embed the Filters struct type which holds fields for filtering and sorting.
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()

	// Use our helpers to extract the title and genres query string values, falling back to the
	// defaults of an empty string and an empty slice, respectively, if they are not provided
	// by the client.
	input.Code = strings.ToLower(app.readStrings(qs, "code", ""))
	input.Starts = app.readTime(qs, "starts", time.Time{})
	input.Expires = app.readTime(qs, "expires", time.Time{})
	input.Active = app.readBool(qs, "active", false)
	input.MinSpend = app.readInt(qs, "minSpend", 0, v)
	input.Category = app.readStrings(qs, "category", "")

	input.Filters.Cursor = app.readStrings(qs, "cursor", "")
	// Ge the page and page_size query string value as integers. Notice that we set the default
	// page value to 1 and default page_size to 20, and that we pass the validator instance
	// as the final argument.
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	// Extract the sort query string value, falling back to "id" if it is not provided
	// by the client (which will imply an ascending sort on movie ID).
	input.Filters.Sort = app.readStrings(qs, "sort", "_id")

	// Add the supported sort value for this endpoint to the sort safelist.
	input.Filters.SortSafeList = []string{
		// ascending sort values
		"_id", "starts", "expires", "active", "minSpend", "category",
		// descending sort values
		"-_id", "-starts", "-expires", "-active", "-minSpend", "-category",
	}

	// Execute the validation checks on the Filters struct and send a response
	// containing the errors if necessary.
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the MovieModel.GetAll method to retrieve the movies, passing in the various filter
	// parameters.
	vouchers, metadata, err := app.Models.Vouchers.GetAllVouchers(input.Code, input.Starts, input.Expires, input.Active, input.MinSpend, input.Category, &input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the movie data.
	if err := app.writeJSON(w, http.StatusOK, envelope{"vouchers": vouchers, "metadata": *metadata}, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
