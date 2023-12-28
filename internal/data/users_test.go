package data

import (
	"reflect"
	"testing"
	"time"
)

func TestInsert(t *testing.T) {
	t.Parallel()

	db, closeDB := newTestDB(t)
	defer closeDB()

	testModel := NewTestModels(db)

	var password Password
	password.Set("password")

	testUser := &User{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      "Test User",
		Email:     "test@example.com",
		Password:  password,
		Addresses: []Address{},
		Phone:     []Phone{},
		Vouchers:  map[string]int{},
		Points:    0,
		Version:   1,
	}

	id, err := testModel.Users.Insert(testUser)
	if err != nil {
		t.Fatal(err)
	}

	testUser.ID = id
	testUser.Password.plaintext = nil

	got, err := testModel.Users.Get(id)
	if err != nil {
		t.Fatal(err)
	}

	// Compare the individual fields of the structs.
	fields := []string{"ID", "Name", "Email", "Password", "Addresses", "Phone", "Vouchers", "Points", "Version"}
	for _, field := range fields {
		wantValue := reflect.ValueOf(testUser).Elem().FieldByName(field).Interface()
		gotValue := reflect.ValueOf(got).Elem().FieldByName(field).Interface()
		if !reflect.DeepEqual(wantValue, gotValue) {
			t.Fatalf("Field %s is not the same. Want: %+v, got: %+v", field, wantValue, gotValue)
		}
	}
}
