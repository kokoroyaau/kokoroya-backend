package clock

import (
	"testing"
	"time"
)

func TestSnapClockIn(t *testing.T) {
	base := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		in   time.Duration
		want time.Duration
	}{
		{0, 0},                               // 9:00 -> 9:00
		{3 * time.Minute, 0},                 // 9:03 -> 9:00 (within grace)
		{4 * time.Minute, 15 * time.Minute},  // 9:04 -> 9:15
		{14 * time.Minute, 15 * time.Minute}, // 9:14 -> 9:15
		{15 * time.Minute, 15 * time.Minute}, // 9:15 -> 9:15
	}
	for _, c := range cases {
		got := snapClockIn(base.Add(c.in))
		want := base.Add(c.want)
		if !got.Equal(want) {
			t.Errorf("snapClockIn(+%v) = %v, want %v", c.in, got, want)
		}
	}
}
