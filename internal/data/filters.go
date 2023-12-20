package data

import (
	"strings"

	"github.com/toduluz/savingsquadsbackend/internal/validator"
)

type Filters struct {
	Cursor       string
	PageSize     int
	Sort         string
	SortSafeList []string
}

// Metadata holds pagination metadata.
type Metadata struct {
	Cursor   string `json:"cursor,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

// formatPaginationData calculates the appropriate pagination data values given the page size and
// the _id of the last document in the current page of results.
func formatPaginationData(pageSize int, lastID string) Metadata {
	return Metadata{
		Cursor:   lastID,
		PageSize: pageSize,
	}
}

// ValidateFilters runs validation checks on the Filters type.
func ValidateFilters(v *validator.Validator, f Filters) {
	// Check that page and page_size parameters contain sensible values.
	v.Check(f.PageSize > 0, "page_size", "must be greater than 0")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	// Check that the sort parameter matches a value in the safelist.
	v.Check(validator.In(f.Sort, f.SortSafeList...), "sort", "invalid sort value")
}

// sortColumn checks that the client-provided Sort field matches one of the entries in our
// SortSafeList and if it does, it extracts the column name from the Sort field by stripping the
// leading hyphen character (if one exists).
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	// The panic below should technically not happen because the Sort value should have already
	// been checked when calling the ValidateFilters helper function. However, this is a sensible
	// failsafe to help stop a NoSQL injection attack from occurring.
	panic("unsafe sort parameter:" + f.Sort)
}

// sortDirection returns the sort direction (1 for "ASC" or -1 for "DESC") depending on the prefix character
// of the Sort field.
func (f Filters) sortDirection() int {
	if strings.HasPrefix(f.Sort, "-") {
		return -1
	}
	return 1
}

func (f Filters) limit() int {
	return f.PageSize
}
