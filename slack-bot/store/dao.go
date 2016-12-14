package store

import "io"

type Dao interface {
	// Embedded a Closer interface
	io.Closer

	// AddBandEvents saves band's events
	AddBandEvents(events []Event) error

	// GetCityEvents returns events in city for period.
	// Period is two Unix time in seconds.
	// It returns empty array if no events.
	GetCityEvents(city string, from, to, offset, limit int64) ([]Event, error)

	// GetBandEvents returns band's events for period.
	// Period is two Unix time in seconds.
	// It returns empty array if no events.
	GetBandEvents(band string, from, to, offset, limit int64) ([]Event, error)

	// GetBandInCityEvents returns band's events in city for period.
	// Period is two Unix time in seconds.
	// It returns empty array if no events.
	GetBandInCityEvents(band string, city string, from, to, offset, limit int64) ([]Event, error)
}
