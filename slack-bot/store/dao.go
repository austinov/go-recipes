package store

import "io"

type Dao interface {
	// Embedded a Closer interface
	io.Closer

	// AddBandEvents saves band's events
	AddBandEvents(events []Event) error

	// GetEvents returns band's events in city for period.
	// Period is two Unix time in seconds.
	// It returns empty array if no events.
	GetEvents(band string, city string, from, to int64, offset, limit int) ([]Event, error)
}
