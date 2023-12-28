package data

import "time"

type MockVoucherModel struct{}

func (m MockVoucherModel) Insert(voucher *Voucher) error {
	return nil
}

func (m MockVoucherModel) Get(code string) (*Voucher, error) {
	return nil, nil
}

func (m MockVoucherModel) GetVoucherList(codes []string) ([]Voucher, error) {
	return nil, nil
}

func (m MockVoucherModel) UpdateUsageCount(code string) error {
	return nil
}

func (m MockVoucherModel) Delete(code string) error {
	return nil
}

func (m MockVoucherModel) GetAllVouchers(code string, startDate time.Time, endDate time.Time, active bool, limit int, sort string, filters *Filters) ([]Voucher, *Metadata, error) {
	return nil, nil, nil
}
