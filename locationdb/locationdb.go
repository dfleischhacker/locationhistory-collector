package locationhistory

import (
	"database/sql"
	"time"

	"github.com/dfleischhacker/locationhistory-collector/configuration"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

type locationDatabaseInterface interface {
	AddWaypoint(username string, latitude float64, longitude float64, datetime time.Time)
	HandleLocationMessage(client mqtt.Client, message mqtt.Message)
}

// LocationDatabase provides a way to store and retrieve location data
type LocationDatabase struct {
	db *sql.DB
}

// OpenLocationDatabase opens a new LocationDatabase based on the connection information provided in the given config.
func OpenLocationDatabase(config configuration.DatabaseConfig) LocationDatabase {
	db, err := sql.Open(config.DriverName, config.Dsn)
	if err != nil {
		log.Fatalf("Unable to open the database connection: %s", err)
	}
	//defer db.Close()

	initDb(db)

	locationDatabase := LocationDatabase{}
	locationDatabase.db = db
	return locationDatabase
}

func initDb(db *sql.DB) {
	_, err := db.Exec(
		`CREATE TABLE IF NOT EXISTS WAYPOINTS (id integer primary key autoincrement,
					topic text not null,
					latitude float not null,
					longitude float not null,
					time timestamp not null)`)
	if err != nil {
		log.Fatal("Error creating database tables", err)
	}
}

// AddWaypoint stores a waypoint with the given information into the database
func (ldb *LocationDatabase) AddWaypoint(topic string, latitude float64, longitude float64, datetime time.Time) {
	tx, err := ldb.db.Begin()
	if err != nil {
		log.Fatal("Unable to open transaction", err)
	}
	stmt, err := tx.Prepare(`INSERT INTO WAYPOINTS(topic, latitude, longitude, time) VALUES (?, ?, ?, ?)`)
	if err != nil {
		log.Fatal("Unable to prepare insert statement", err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(topic, latitude, longitude, datetime)
	if err != nil {
		log.Fatal("Unable to write entry to database", err)
	}
	tx.Commit()

}
