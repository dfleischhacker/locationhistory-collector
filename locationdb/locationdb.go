package locationhistory

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/dfleischhacker/locationhistory-collector/configuration"
	log "github.com/sirupsen/logrus"
)

// LocationDatabase provides a way to store and retrieve location data
type LocationDatabase struct {
	db *sql.DB
}

// Waypoint is the content of a single location message
type Waypoint struct {
	ID        int       `json:"id"`
	Topic     string    `json:"topic"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Datetime  time.Time `json:"time"`
}

type RunningTransaction struct {
	tx   *sql.Tx
	stmt *sql.Stmt
}

// OpenLocationDatabase opens a new LocationDatabase based on the connection information provided in the given config.
func OpenLocationDatabase(config configuration.DatabaseConfig) LocationDatabase {
	db, err := sql.Open(config.DriverName, config.Dsn)
	if err != nil {
		log.Fatalf("Unable to open the database connection: %s", err)
	}

	initDb(db)

	locationDatabase := LocationDatabase{}
	locationDatabase.db = db
	return locationDatabase
}

func initDb(db *sql.DB) {
	_, err := db.Exec(
		`CREATE TABLE IF NOT EXISTS WAYPOINTS (ID INTEGER PRIMARY KEY AUTOINCREMENT,
					Topic TEXT NOT NULL,
					Latitude DOUBLE NOT NULL,
					Longitude DOUBLE NOT NULL,
					Time TIMESTAMP NOT NULL,
					CONSTRAINT all_unique UNIQUE (Topic, Latitude, Longitude, Time))`)
	if err != nil {
		log.Fatal("Error creating database tables", err)
	}
}

// AddWaypointData stores a waypoint with the given information into the database and commits the change
func (ldb *LocationDatabase) AddWaypointData(topic string, latitude float64, longitude float64, datetime time.Time) {
	tx, err := ldb.db.Begin()
	if err != nil {
		log.Fatal("Unable to open transaction", err)
	}
	stmt, err := tx.Prepare(`INSERT INTO WAYPOINTS(topic, latitude, longitude, time) VALUES (?, ?, ?, ?)`)
	if err != nil {
		log.Fatal("Unable to prepare insert statement: ", err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(topic, latitude, longitude, datetime)
	if err != nil {
		if !strings.HasPrefix(err.Error(), "UNIQUE constraint failed:") {
			log.Warn("Unable to write entry to database: ", err)
		} else {
			log.Info("Received duplicate location message, ignoring it")
		}
		tx.Rollback()
	} else {
		tx.Commit()
	}
}

func (ldb *LocationDatabase) AddWaypoint(waypoint Waypoint) {
	ldb.AddWaypointData(waypoint.Topic, waypoint.Latitude, waypoint.Longitude, waypoint.Datetime)
}

// GetWaypoints returns all waypoints for a given topic name
func (ldb *LocationDatabase) GetWaypoints(topic string) ([]Waypoint, error) {
	stmt, err := ldb.db.Prepare(`SELECT * FROM WAYPOINTS WHERE topic = ? ORDER BY time ASC`)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(topic)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	waypoints := make([]Waypoint, 0)

	for rows.Next() {
		var waypoint Waypoint
		_ = rows.Scan(&waypoint.ID, &waypoint.Topic, &waypoint.Latitude, &waypoint.Longitude, &waypoint.Datetime)
		waypoints = append(waypoints, waypoint)
	}

	return waypoints, nil
}

// GetTopics returns all topics for which location data exists
func (ldb *LocationDatabase) GetTopics() ([]string, error) {
	rows, err := ldb.db.Query(`SELECT DISTINCT topic FROM WAYPOINTS`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	topics := make([]string, 0)

	for rows.Next() {
		var name string
		_ = rows.Scan(&name)
		topics = append(topics, name)
	}

	return topics, nil
}

func (ldb *LocationDatabase) OpenTransaction() (RunningTransaction, error) {
	runningTx := RunningTransaction{}
	tx, err := ldb.db.Begin()
	if err != nil {
		return runningTx, err
	}
	stmt, err := tx.Prepare(`INSERT INTO WAYPOINTS(topic, latitude, longitude, time) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return runningTx, err
	}
	runningTx.tx = tx
	runningTx.stmt = stmt
	return runningTx, nil
}

// AddWaypointData stores a waypoint with the given information into the database and commits the change
func (rtx *RunningTransaction) AddWaypointData(topic string, latitude float64, longitude float64, datetime time.Time) {
	_, err := rtx.stmt.Exec(topic, latitude, longitude, datetime)
	if err != nil {
		if !strings.HasPrefix(err.Error(), "UNIQUE constraint failed:") {
			log.Warn("Unable to write entry to database: ", err)
		} else {
			log.Info("Received duplicate location message, ignoring it")
		}
	}
}

func (rtx *RunningTransaction) AddWaypoint(waypoint Waypoint) {
	rtx.AddWaypointData(waypoint.Topic, waypoint.Latitude, waypoint.Longitude, waypoint.Datetime)
}

func (rtx *RunningTransaction) Commit() error {
	err := rtx.tx.Commit()
	if err != nil {
		return err
	}
	err = rtx.stmt.Close()
	return err
}

func (w *Waypoint) String() string {
	return fmt.Sprintf("%s: %f - %f @ %s", w.Topic, w.Latitude, w.Longitude, w.Datetime)
}
