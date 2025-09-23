package utils

import (
	"fmt"
    "time"
)

// DefaultTimeZone is the default timezone used by the application
var DefaultTimeZone = time.FixedZone("Asia/Jakarta", 7*60*60) // UTC+7

// GetEndOfDayDuration returns the time duration until the end of the current day
// This is used for setting cache expiration times that should expire at midnight
func GetEndOfDayDuration() time.Duration {
    now := time.Now()
    endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
    return endOfDay.Sub(now)
}

// GetStartOfDay returns the timestamp for the start of the given day
func GetStartOfDay(t time.Time) time.Time {
    return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// GetEndOfDay returns the timestamp for the end of the given day
func GetEndOfDay(t time.Time) time.Time {
    return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// FormatDuration formats a time.Duration into a human-readable string
// Examples: "2h 30m", "3d 4h", "1w 2d"
func FormatDuration(d time.Duration) string {
    days := int(d.Hours() / 24)
    hours := int(d.Hours()) % 24
    minutes := int(d.Minutes()) % 60

    if days > 7 {
        weeks := days / 7
        remainingDays := days % 7
        if remainingDays > 0 {
            return time.Duration(weeks*7*24).String() + " " + time.Duration(remainingDays*24).String()
        }
        return time.Duration(weeks * 7 * 24).String()
    }

    if days > 0 {
        if hours > 0 {
            return time.Duration(days*24).String() + " " + time.Duration(hours).String()
        }
        return time.Duration(days * 24).String()
    }

    if hours > 0 {
        if minutes > 0 {
            return time.Duration(hours).String() + " " + time.Duration(minutes).String()
        }
        return time.Duration(hours).String()
    }

    return d.String()
}

// ParseDuration parses a duration string with support for days and weeks
// Format: "1d" = 1 day, "1w" = 1 week
// Falls back to time.ParseDuration for standard duration strings (h, m, s)
func ParseDuration(s string) (time.Duration, error) {
    // First try standard parsing
    d, err := time.ParseDuration(s)
    if err == nil {
        return d, nil
    }

    // Custom parsing for days and weeks
    var duration time.Duration
    var value int
    var unit string

    _, err = fmt.Sscanf(s, "%d%s", &value, &unit)
    if err != nil {
        return 0, fmt.Errorf("invalid duration format: %s", s)
    }

    switch unit {
    case "d":
        duration = time.Duration(value) * 24 * time.Hour
    case "w":
        duration = time.Duration(value) * 7 * 24 * time.Hour
    default:
        return 0, fmt.Errorf("unsupported duration unit: %s", unit)
    }

    return duration, nil
}