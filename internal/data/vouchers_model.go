package data

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/toduluz/savingsquadsbackend/internal/validator"
)

type Voucher struct {
	Code         string    `json:"id" bson:"_id"`
	CreatedAt    time.Time `json:"-" bson:"created_at"`
	ModifiedAt   time.Time `json:"-" bson:"updated_at"`
	Description  string    `json:"description" bson:"description"`
	Discount     int       `json:"discount" bson:"discount"`
	IsPercentage bool      `json:"isPercentage" bson:"isPercentage"`
	Starts       time.Time `json:"start" bson:"start"`
	Expires      time.Time `json:"expires" bson:"expires"`
	Active       bool      `json:"active" bson:"active"`
	UsageLimit   int       `json:"usageLimit" bson:"usageLimit"`
	UsageCount   int       `json:"usageCount" bson:"usageCount"`
	MinSpend     int       `json:"minSpend" bson:"minSpend"`
	Category     string    `json:"category" bson:"category"`
}

func (v *Voucher) VocuherCodeGenerator() error {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

	b := make([]byte, 15) // Generate 15 random bytes
	for i := range b {
		max := big.NewInt(int64(len(letters)))
		randNum, err := rand.Int(rand.Reader, max)
		if err != nil {
			return err
		}
		b[i] = letters[randNum.Int64()]
	}

	v.Code = string(b)
	// Return the code
	return nil
}

// ValidateVoucher runs validation checks on the Voucher type.
func ValidateVoucher(v *validator.Validator, voucher *Voucher) {
	v.Check(voucher.Code != "", "code", "must be provided")
	v.Check(len(voucher.Code) <= 20, "code", "must not be more than 20 characters long")

	v.Check(voucher.Description != "", "description", "must be provided")
	v.Check(len(voucher.Description) <= 500, "description", "must not be more than 500 characters long")

	v.Check(voucher.Discount >= 0, "discount", "must be a positive number")
	v.Check(voucher.Discount <= 100, "discount", "must not be more than 100")

	v.Check(voucher.UsageLimit >= 0, "usageLimit", "must be a positive number")

	v.Check(voucher.Starts.Before(voucher.Expires), "start", "must be before the expiry date")
}
