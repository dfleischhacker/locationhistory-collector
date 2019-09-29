package rest

import (
	locationhistory "github.com/dfleischhacker/locationhistory-collector/locationdb"
	"log"
	"net/http"
)

type LocationHistoryService struct {
	ldb *locationhistory.LocationDatabase
}

func getData(w http.ResponseWriter, req *http.Request) {

}

func NewRestService(ldb *locationhistory.LocationDatabase) {
	http.HandleFunc("/", getData)
	log.Fatal(http.ListenAndServe(":10000", nil))
}
