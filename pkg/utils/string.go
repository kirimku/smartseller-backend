package utils

import (
	"strconv"
	"time"
)

// ParseDateTime parses a string like "2024-02-12 10:04:28" into time.Time in WIB (Asia/Jakarta)
func ParseDateTime(value string) (time.Time, error) {
	layout := "2006-01-02 15:04:05"

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.Time{}, err
	}

	return time.ParseInLocation(layout, value, loc)
}

func ToFloat64(value string) float64 {
	if val, err := strconv.ParseFloat(value, 64); err == nil {
		return val
	}

	return 0
}

func ToFloat64Ptr(value string) *float64 {
	if val, err := strconv.ParseFloat(value, 64); err == nil {
		return &val
	}

	return nil
}

// SumFloat64Ptr sums multiple strings and returns the result as *float64
func SumFloat64Ptr(values ...string) *float64 {
	var sum float64
	valid := false

	for _, v := range values {
		if f := ToFloat64Ptr(v); f != nil {
			sum += *f
			valid = true
		}
	}

	if !valid {
		return nil // no valid numbers
	}
	return &sum
}
