package dateutil

import "time"

var BusinessLocation = mustLoadLocation("Australia/Sydney")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

func DayOf(t time.Time) time.Time {
	local := t.In(BusinessLocation)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, BusinessLocation)
}

func MondayOf(t time.Time) time.Time {
	offset := int(t.Weekday()) - int(time.Monday)
	if offset < 0 {
		offset += 7
	}
	return t.AddDate(0, 0, -offset)
}
