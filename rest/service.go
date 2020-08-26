package rest

import (
	"net/http"
	"strconv"
	"time"

	"github.com/dfleischhacker/locationhistory-collector/configuration"
	locationhistory "github.com/dfleischhacker/locationhistory-collector/locationdb"
	log "github.com/sirupsen/logrus"
)

type LocationHistoryService struct {
	ldb *locationhistory.LocationDatabase
}

func NewRestService(config *configuration.Configuration, ldb *locationhistory.LocationDatabase, port int) {
	router := http.NewServeMux()

	router.HandleFunc("/locations/", func(writer http.ResponseWriter, request *http.Request) {
		topic := request.URL.Path[11:]
		log.Infof("Retrieving data for topic '%s'", topic)
		startTime := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.Local)
		endTime := time.Date(2020, time.August, 27, 0, 0, 0, 0, time.Local)
		maxCount := 100000
		waypoints, err := ldb.GetWaypoints(topic, &startTime, &endTime, &maxCount)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Infof("Got %d waypoints", len(waypoints))
		gpxDoc := GenerateGpx(waypoints)
		bytes, err := GetGpxStream(gpxDoc)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = writer.Write(bytes)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		log.Info("Got request for /")
		_, err := writer.Write(GetIndexFile(config.Map.Token))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Infof("Starting up server on port %d", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), router))
}
