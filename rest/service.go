package rest

import (
	"net/http"
	"strconv"
	"time"

	"github.com/dfleischhacker/locationhistory-collector/configuration"
	locationhistory "github.com/dfleischhacker/locationhistory-collector/locationdb"
	"github.com/dfleischhacker/locationhistory-collector/rest/static"
	log "github.com/sirupsen/logrus"
)

// LocationHistoryService provides the data structures for the history service
type LocationHistoryService struct {
	ldb *locationhistory.LocationDatabase
}

// NewRestService returns a new REST service using the given config and location database
func NewRestService(config *configuration.Configuration, ldb *locationhistory.LocationDatabase) {
	router := http.NewServeMux()

	router.HandleFunc("/locations/", func(writer http.ResponseWriter, request *http.Request) {
		topic := request.URL.Path[11:]
		log.Infof("Retrieving data for topic '%s'", topic)
		startTime := time.Date(2015, time.January, 1, 0, 0, 0, 0, time.Local)
		endTime := time.Date(2020, time.August, 27, 0, 0, 0, 0, time.Local)
		maxCount := 1000000
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

	router.Handle("/", http.FileServer(static.AssetFile()))

	/*router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		log.Info("Got request for /")
		_, err := writer.Write(GetIndexFile())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	})*/

	router.HandleFunc("/token", func(writer http.ResponseWriter, request *http.Request) {
		log.Info("Got token request")
		_, err := writer.Write(GetMapboxToken(config.Map.Token))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Infof("Starting up server on port %d", config.Map.Port)
	log.Fatal(http.ListenAndServe(config.Map.BindAddress+":"+strconv.Itoa(config.Map.Port), router))
}

// GetMapboxToken returns the mapbox token
func GetMapboxToken(token string) []byte {
	return []byte(token)
}
