package data

import "github.com/paulcockrell/greenlight/internal/validator"

type Filters struct {
	Sort         string
	SortSafeList []string
	Page         int
	PageSize     int
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page < 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize < 100, "page_size", "must be a maximum of 100")
	v.Check(validator.In(f.Sort, f.SortSafeList...), "sort", "invalid sort value")
}
