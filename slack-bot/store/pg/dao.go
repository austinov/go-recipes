package pg

import (
	"fmt"
	"log"

	"database/sql"

	"github.com/austinov/go-recipes/slack-bot/config"
	"github.com/austinov/go-recipes/slack-bot/store"
	_ "github.com/lib/pq"
)

const (
	bandInsert = `
	WITH s AS (
	    SELECT id
	    FROM band
	    WHERE name = $1
	), i as (
	    INSERT INTO band (name)
	    SELECT $1
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
	    WHERE name = $1
	), i as (
	    INSERT INTO city (name)
	    SELECT $1
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	eventsInCity = `
	    SELECT *
		FROM vw_events
		WHERE city_name = $1 AND begin_dt >= $2 AND end_dt <= $3`

	eventsBand = `
	    SELECT *
		FROM vw_events
		WHERE band_name = $1 AND begin_dt >= $2 AND end_dt <= $3`

	eventsBandInCity = `
	    SELECT *
		FROM vw_events
		WHERE band_name = $1 AND city_name = $2 AND begin_dt >= $3 AND end_dt <= $4`
)

var (
	bandInsertStmt       *sql.Stmt
	cityInsertStmt       *sql.Stmt
	eventsClearStmt      *sql.Stmt
	eventsInsertStmt     *sql.Stmt
	eventsInCityStmt     *sql.Stmt
	eventsBandStmt       *sql.Stmt
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
	eventsInCityStmt, err = db.Prepare(eventsInCity)
	if err != nil {
		log.Fatal(err)
	}
	eventsBandStmt, err = db.Prepare(eventsBand)
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
	log.Println("Close Dao in pg package")
	bandInsertStmt.Close()
	cityInsertStmt.Close()
	eventsClearStmt.Close()
	eventsInsertStmt.Close()
	eventsInCityStmt.Close()
	eventsBandStmt.Close()
	eventsBandInCityStmt.Close()
	d.db.Close()
	return nil
}

func (d *Dao) GetCalendar(band string, from, to int64) ([]store.Event, error) {
	// TODO
	return nil, nil
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
		if err := tx.Stmt(bandInsertStmt).QueryRow(events[0].Band).Scan(&bandId); err != nil {
			return fmt.Errorf("insert band failed with %#v (band's name is %#v)\n", err, events[0].Band)
		}
		// clear previouse data
		if _, err = tx.Stmt(eventsClearStmt).Exec(bandId); err != nil {
			return fmt.Errorf("clear previouse band's events failed with %#v (band's id is %#v)\n", err, bandId)
		}
		for _, event := range events {
			var cityId int32
			// add city if not exist
			if err := tx.Stmt(cityInsertStmt).QueryRow(event.City).Scan(&cityId); err != nil {
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

func (d *Dao) GetCityEvents(city string, from, to, offset, limit int64) ([]store.Event, error) {
	// TODO
	/*
		rows, err := eventsInCityStmt.Query(city, from, to, offset, limit)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			if err = rows.Scan(&secs, &calls); err != nil {
				return 0, 0, err
			}
		}
		if err := rows.Err(); err != nil {
			return 0, 0, err
		}
	*/

	return nil, nil
}

func (d *Dao) GetBandEvents(band string, from, to, offset, limit int64) ([]store.Event, error) {
	// TODO
	return nil, nil
}

func (d *Dao) GetBandInCityEvents(band string, city string, from, to, offset, limit int64) ([]store.Event, error) {
	// TODO
	return nil, nil
}
