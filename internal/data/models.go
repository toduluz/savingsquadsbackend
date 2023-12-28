package data

import (
	"errors"
	"io"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	// ErrRecordNotFound is returned when a movie record doesn't exist in database.
	ErrRecordNotFound = errors.New("record not found")

	// ErrEditConflict is returned when a there is a data race, and we have an edit conflict.
	ErrEditConflict = errors.New("edit conflict")
)

// Models struct is a single convenient container to hold and represent all our database models.
type Models struct {
	Vouchers interface {
		Insert(voucher *Voucher) error
		Get(string) (*Voucher, error)
		GetVoucherList([]string) ([]Voucher, error)
		UpdateUsageCount(string) error
		Delete(string) error
		GetAllVouchers(string, time.Time, time.Time, bool, int, string, *Filters) ([]Voucher, *Metadata, error)
	}
	Users interface {
		Insert(user *User) (string, error)
		Get(string) (*User, error)
		GetByEmail(string) (*User, error)
		GetAllVouchers(string) (map[string]int, error)
		RedeemVoucher(string, string, int) error
		GetPoints(string) (int, error)
		AddPoints(string, int) error
		DeductPointsAndCreateVoucher(string, int, *Voucher) error
		UpdateVoucherList(string, map[string]int) error
	}
}

func NewModels(db *mongo.Database) Models {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	return Models{
		Vouchers: VoucherModel{
			DB:       db,
			InfoLog:  infoLog,
			ErrorLog: errorLog,
		},
		Users: UserModel{
			DB:       db,
			InfoLog:  infoLog,
			ErrorLog: errorLog,
		},
	}
}

func NewMockModels() Models {
	return Models{
		Vouchers: MockVoucherModel{},
		Users:    MockUserModel{},
	}
}

func NewTestModels(db *mongo.Database) Models {
	infoLog := log.New(io.Discard, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(io.Discard, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	return Models{
		Vouchers: VoucherModel{
			DB:       db,
			InfoLog:  infoLog,
			ErrorLog: errorLog,
		},
		Users: UserModel{
			DB:       db,
			InfoLog:  infoLog,
			ErrorLog: errorLog,
		},
	}
}
