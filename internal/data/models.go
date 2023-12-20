package data

import (
	"errors"
	"log"
	"os"

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
	Vouchers VoucherModel
	Users    UserModel
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
