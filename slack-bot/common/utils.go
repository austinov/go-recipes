package common

import "time"

const oneDay = (24*60*60 - 1) * time.Second

// BeginOfDate returns begin of date
func BeginOfDate(d time.Time) time.Time {
	t := d.Truncate(time.Hour)
	return t.Add(time.Duration(-1*t.Hour()) * time.Hour)
}

// EndOfDate returns end of date - begin date minus one second
func EndOfDate(d time.Time) time.Time {
	t := d.Truncate(time.Hour)
	return t.Add(time.Duration(-1*t.Hour()) * time.Hour).Add(oneDay)
}
