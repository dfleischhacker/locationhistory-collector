package importer

import (
	"encoding/json"
	"github.com/dfleischhacker/locationhistory-collector/locationdb"
	"github.com/dfleischhacker/locationhistory-collector/utils"
	log "github.com/sirupsen/logrus"
	"os"
)

// ImportTimeline reads the exported Google Timeline data from the given `fileName` and imports it into the
// provided `database`. The return value is the number of imported waypoints or, if an error occurs, the error.
// The data is returned for the given `topic`.
func ImportTimeline(database *locationhistory.LocationDatabase, fileName string, topic string) (int, error) {
	var count = 0
	stream, err := os.Open(fileName)
	if err != nil {
		return 0, err
	}
	defer stream.Close()
	dec := json.NewDecoder(stream)

	rTx, err := database.OpenTransaction()
	if err != nil {
		return 0, nil
	}

	// read open curly brace, the locations key and the opening bracket
	_, err = dec.Token()
	if err != nil {
		log.Fatal(err)
	}
	dec.More()
	_, err = dec.Token()
	dec.More()
	_, err = dec.Token()

	for dec.More() {
		var twp timelineWaypoint
		// decode an array value (Message)
		err := dec.Decode(&twp)
		if err != nil {
			log.Fatal(err)
		}

		count += 1
		rTx.AddWaypoint(twp.toWaypoint(topic))
	}

	// read closing bracket
	_, err = dec.Token()
	if err != nil {
		log.Fatal(err)
	}

	err = rTx.Commit()
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (twp *timelineWaypoint) toWaypoint(topic string) locationhistory.Waypoint {
	datetime, err := utils.GetUnixTime(twp.TimestampMs, 1000)
	if err != nil {
		log.Fatal(err)
	}
	return locationhistory.Waypoint{
		Topic:     topic,
		Datetime:  datetime.Time,
		Longitude: float64(twp.Longitude) / 10000000,
		Latitude:  float64(twp.Latitude) / 10000000,
	}
}

type timelineWaypoint struct {
	TimestampMs      string `json:"timestampMs"`
	Latitude         int64  `json:"latitudeE7"`
	Longitude        int64  `json:"longitudeE7"`
	Accuracy         int    `json:"accuracy"`
	Velocity         int    `json:"velocity"`
	Altitude         int    `json:"altitude"`
	VerticalAccuracy int    `json:"verticalAccuracy"`
}
