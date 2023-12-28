package data

type MockUserModel struct{}

func (m MockUserModel) Insert(user *User) (string, error) {
	switch user.Email {
	case "test@example.com":
		return "testID", nil
	default:
		return "", ErrDuplicateEmail
	}
}

func (m MockUserModel) Get(id string) (*User, error) {
	return nil, nil
}

func (m MockUserModel) GetByEmail(email string) (*User, error) {
	return nil, nil
}

func (m MockUserModel) GetAllVouchers(id string) (map[string]int, error) {
	return nil, nil
}

func (m MockUserModel) RedeemVoucher(userID string, voucherCode string, usageCount int) error {
	return nil
}

func (m MockUserModel) GetPoints(id string) (int, error) {
	return 0, nil
}

func (m MockUserModel) AddPoints(id string, points int) error {
	return nil
}

func (m MockUserModel) DeductPointsAndCreateVoucher(id string, points int, voucher *Voucher) error {
	return nil
}

func (m MockUserModel) UpdateVoucherList(id string, vouchers map[string]int) error {
	return nil
}
