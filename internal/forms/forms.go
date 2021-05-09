package forms

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
)

// Form creates a custom form struct, embeds a url.Values object
type Form struct {
	url.Values
	Errors errors
}

// New initializes a form struct
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// Has checks if form field is in post and not empty
func (f *Form) Has(field string) bool {
	value := f.Get(field)
	return value != ""
}

// Valid returns true if there are no errors, otherwise false
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// Required checks for required fields
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

// IsAboveMinLength checks for string minimum length
func (f *Form) IsAboveMinLength(field string, length int) bool {
	value := f.Get(field)
	if len(value) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", length))
		return false
	}
	return true
}

// IsEmail checks for valid email address
func (f *Form) IsEmail(field string) {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid email address")
	}
}

// GetTimeObj gets the time.Time object date
func (f *Form) GetTimeObj(field string) (time.Time , error) {
	var date time.Time
	d := f.Get(field)
	// 01/02 03:04:05PM '06 -0700
	layout := "2006-01-02"
	date, err := time.Parse(layout, d)
	if err != nil {
		return date, err
	}
	return date, nil
}

// ValidateDate checks the date
func (f *Form) ValidateDate(field string) {
	date, err := f.GetTimeObj(field)
	if err != nil {
		f.Errors.Add(field, "Invalid date: date must be YYYY-MM-DD format")
	}

	tomorrow := time.Now().Add(24 * time.Hour)
	if date.Before(tomorrow) {
		f.Errors.Add(field, "Invalid date: date must be after tomorrow")
	}
}

func (f *Form) EndDateGreaterThanStartDate(start_date, end_date string) {
	startDate, err := f.GetTimeObj(start_date)
	if err != nil {
		f.Errors.Add(start_date, "Invalid date: date must be YYYY-MM-DD format")
	}
	endDate, err := f.GetTimeObj(end_date)
	if err != nil {
		f.Errors.Add(end_date, "Invalid date: date must be YYYY-MM-DD format")
	}

	if endDate.Before(startDate.Add(24 * time.Hour)) {
		f.Errors.Add(end_date, "Invalid date: end date must be after the start date")
	}
}
