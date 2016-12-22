package pg

import (
	"fmt"
	"log"
	"strings"

	"database/sql"

	"github.com/austinov/go-recipes/rocker-bot/config"
	"github.com/austinov/go-recipes/rocker-bot/store"
	_ "github.com/lib/pq"
)

const (
	bandInsert = `
	WITH s AS (
	    SELECT id
	    FROM band
	    WHERE lower(name) = $1
	), i as (
	    INSERT INTO band (name)
	    SELECT $2
	    WHERE NOT EXISTS (SELECT 1 FROM s)
	    RETURNING id
	)
	SELECT id FROM i
	UNION ALL
	SELECT id FROM s`

	cityInsert = `
	WITH s AS (
	    SELECT id
	    FROM city
	    WHERE lower(name) = $1
	), i as (
	    INSERT INTO city (name)
	    SELECT $2
	    WHERE NOT EXISTS (SELECT 1 FROM s)
	    RETURNING id
	)
	SELECT id FROM i
	UNION ALL
	SELECT id FROM s`

	eventsClear = `
	    DELETE FROM event
		WHERE band_id = $1`

	eventInsert = `
	    INSERT INTO event(title, begin_dt, end_dt, band_id, city_id, venue, link, img)
		SELECT $1::VARCHAR, $2, $3, $4, $5, $6, $7, $8
		WHERE NOT EXISTS(
			SELECT id
			FROM event
			WHERE title = $1 AND begin_dt = $2 AND end_dt = $3 AND band_id = $4 AND city_id = $5
		)`

	eventsBandInCity = `
	    SELECT title, begin_dt, end_dt, city_name, venue, link, img, string_agg(DISTINCT band_name, ', ') AS bands
		FROM vw_events
		WHERE lower(band_name) = COALESCE($1, lower(band_name)) AND 
		      lower(city_name) = COALESCE($2, lower(city_name)) AND
			  begin_dt >= $3 AND end_dt <= $4
		GROUP BY title, begin_dt, end_dt, city_name, venue, link, img
		ORDER BY begin_dt OFFSET $5 LIMIT $6`
)

var (
	bandInsertStmt       *sql.Stmt
	cityInsertStmt       *sql.Stmt
	eventsClearStmt      *sql.Stmt
	eventsInsertStmt     *sql.Stmt
	eventsBandInCityStmt *sql.Stmt
)

type Dao struct {
	db *sql.DB
}

func New(cfg config.DBConfig) store.Dao {
	db, err := sql.Open("postgres", cfg.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	bandInsertStmt, err = db.Prepare(bandInsert)
	if err != nil {
		log.Fatal(err)
	}
	cityInsertStmt, err = db.Prepare(cityInsert)
	if err != nil {
		log.Fatal(err)
	}
	eventsClearStmt, err = db.Prepare(eventsClear)
	if err != nil {
		log.Fatal(err)
	}
	eventsInsertStmt, err = db.Prepare(eventInsert)
	if err != nil {
		log.Fatal(err)
	}
	eventsBandInCityStmt, err = db.Prepare(eventsBandInCity)
	if err != nil {
		log.Fatal(err)
	}
	return &Dao{
		db,
	}
}

func (d *Dao) Close() error {
	bandInsertStmt.Close()
	cityInsertStmt.Close()
	eventsClearStmt.Close()
	eventsInsertStmt.Close()
	eventsBandInCityStmt.Close()
	d.db.Close()
	return nil
}

func (d *Dao) AddBandEvents(events []store.Event) error {
	if len(events) == 0 {
		return nil
	}
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	if err = func() error {
		var bandId int32
		// add band if not exist
		bandName := events[0].Band
		if err := tx.Stmt(bandInsertStmt).QueryRow(strings.ToLower(bandName), bandName).Scan(&bandId); err != nil {
			return fmt.Errorf("insert band failed with %#v (band's name is %#v)\n", err, events[0].Band)
		}
		// clear previouse data
		if _, err = tx.Stmt(eventsClearStmt).Exec(bandId); err != nil {
			return fmt.Errorf("clear previouse band's events failed with %#v (band's id is %#v)\n", err, bandId)
		}
		for _, event := range events {
			var cityId int32
			// add city if not exist
			if err := tx.Stmt(cityInsertStmt).QueryRow(strings.ToLower(event.City), event.City).Scan(&cityId); err != nil {
				return fmt.Errorf("insert city failed with %#v (event is %#v)\n", err, event)
			}
			// add event
			if _, err = tx.Stmt(eventsInsertStmt).Exec(event.Title, event.From, event.To, bandId, cityId, event.Venue, event.Link, event.Img); err != nil {
				return fmt.Errorf("insert band's event failed with %#v (event is %#v)\n", err, event)
			}
		}
		return nil
	}(); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (d *Dao) GetEvents(band string, city string, from, to int64, offset, limit int) ([]store.Event, error) {
	var b interface{} = nil
	var c interface{} = nil
	if band != "" {
		b = strings.ToLower(band)
	}
	if city != "" {
		c = strings.ToLower(city)
	}

	rows, err := eventsBandInCityStmt.Query(b, c, from, to, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return d.rowsToEvents(rows)
}

func (d *Dao) rowsToEvents(rows *sql.Rows) ([]store.Event, error) {
	events := make([]store.Event, 0)
	for rows.Next() {
		var (
			title, bands, city string
			venue, link, img   string
			from, to           int64
		)
		if err := rows.Scan(&title, &from, &to, &city, &venue, &link, &img, &bands); err != nil {
			return nil, err
		}
		events = append(events, store.Event{
			Band:  bands,
			Title: title,
			From:  from,
			To:    to,
			City:  city,
			Venue: venue,
			Link:  link,
			Img:   img,
		})
	}
	return events, rows.Err()
}
