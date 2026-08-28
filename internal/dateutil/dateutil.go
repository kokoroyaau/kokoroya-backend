package dateutil

import "time"

// BusinessLocation is the timezone the business operates in (Sydney/
// Melbourne, AEST/AEDT). The database always stores instants in UTC —
// this is only used to decide which calendar day/week an instant falls
// on for reporting, never for storage.
var BusinessLocation = mustLoadLocation("Australia/Sydney")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

// DayOf returns the start of the calendar day (in BusinessLocation) that t
// falls on, regardless of what location t itself is currently in (e.g. a
// timestamptz value scanned back as UTC).
func DayOf(t time.Time) time.Time {
	local := t.In(BusinessLocation)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, BusinessLocation)
}

// MondayOf returns the Monday of the week containing t, used to resolve
// which week's carry-forward rate (net sales / labour) applies to a date.
func MondayOf(t time.Time) time.Time {
	offset := int(t.Weekday()) - int(time.Monday)
	if offset < 0 {
		offset += 7
	}
	return t.AddDate(0, 0, -offset)
}
