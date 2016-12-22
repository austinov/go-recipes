CREATE TABLE IF NOT EXISTS band (
    "id"   serial primary key,
    "name" varchar(100) NOT NULL
);

CREATE INDEX ind_band_id ON band USING btree (id);
CREATE UNIQUE INDEX uni_band ON band (lower(name));

CREATE TABLE IF NOT EXISTS city (
    "id"   serial primary key,
    "name" varchar(100) NOT NULL
);

CREATE INDEX ind_city_id ON city USING btree (id);
CREATE UNIQUE INDEX uni_city ON city (lower(name));

CREATE TABLE IF NOT EXISTS event (
    "id"       serial primary key,
    "title"    varchar(255) NOT NULL,
    "begin_dt" bigint,
    "end_dt"   bigint,
    "band_id"  integer NOT NULL,
    "city_id"  integer NOT NULL,
    "venue"    varchar(255),
    "link"     varchar(255),
    "img"      varchar(255)
);

CREATE INDEX ind_event_id ON event USING btree (id);
CREATE INDEX ind_event_title ON event USING btree (title);
CREATE INDEX ind_event_begin ON event USING btree (begin_dt);
CREATE INDEX ind_event_end ON event USING btree (end_dt);
CREATE INDEX ind_event_band ON event USING btree (band_id);
CREATE INDEX ind_event_city ON event USING btree (city_id);
CREATE UNIQUE INDEX uni_event ON event (title, begin_dt, end_dt, band_id, city_id);
ALTER TABLE event ADD CONSTRAINT fk_event_band FOREIGN KEY (band_id) REFERENCES band (id);
ALTER TABLE event ADD CONSTRAINT fk_event_city FOREIGN KEY (city_id) REFERENCES city (id);

CREATE OR REPLACE VIEW vw_events AS
    SELECT e.*, c.name AS city_name, b.name AS band_name
	FROM event e
	    JOIN city c ON e.city_id = c.id
	    JOIN band b ON e.band_id = b.id
