package dao

import "io"

type Dao interface {
	// Embedded a Closer interface
	io.Closer

	// GetCalendar returns events for band for period
	// between two Unix time in seconds.
	// It returns empty array if no events.
	GetCalendar(band string, from, to int64) ([]Event, error)
}
