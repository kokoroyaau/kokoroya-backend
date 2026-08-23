package dateutil

import "time"

// MondayOf returns the Monday of the week containing t, used to resolve
// which week's carry-forward rate (net sales / labour) applies to a date.
func MondayOf(t time.Time) time.Time {
	offset := int(t.Weekday()) - int(time.Monday)
	if offset < 0 {
		offset += 7
	}
	return t.AddDate(0, 0, -offset)
}
