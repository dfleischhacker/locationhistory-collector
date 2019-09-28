package main

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	locationhistory "github.com/dfleischhacker/locationhistory-collector/locationdb"

	"github.com/dfleischhacker/locationhistory-collector/configuration"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)

	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <config file>", os.Args[0])
	}

	history := NewLocationHistory(os.Args[1])
	history.Run()
}

// LocationHistory contains all relevant data of the locationhistory program
type LocationHistory struct {
	configuration    *configuration.Configuration
	mqttClient       mqtt.Client
	locationDatabase locationhistory.LocationDatabase
}

// UnixTime wraps a time.Time to provide JSON unmarshalling from unix timestamps
type UnixTime struct {
	time.Time
}

// OwntracksMessage represents a message sent by Owntracks
type OwntracksMessage struct {
	Battery              int      `json:"batt"`
	Longitude            float64  `json:"lon"`
	Latitude             float64  `json:"lat"`
	Accuracy             int      `json:"acc"`
	Pressure             float64  `json:"p"`
	BatteryState         int      `json:"bs"`
	VerticalAccuracy     int      `json:"vac"`
	Trigger              string   `json:"t"`
	InternetConnectivity string   `json:"conn"`
	Timestamp            UnixTime `json:"tst"`
	Altitude             int      `json:"alt"`
	TrackerID            string   `json:"tid"`
}

// NewLocationHistory creates a new location history configured from the given configFile
func NewLocationHistory(configFile string) LocationHistory {
	history := LocationHistory{}
	history.configuration = configuration.LoadConfiguration(configFile)

	log.Debug("Connecting to database")
	history.locationDatabase = locationhistory.OpenLocationDatabase(history.configuration.Database)
	log.Debug("Connected to database")

	log.Debug("Connecting to MQTT broker")
	clientOptions := mqtt.NewClientOptions().AddBroker(history.configuration.Mqtt.URL)
	history.mqttClient = mqtt.NewClient(clientOptions)

	return history
}

// Run starts the connection to the MQTT broker and writes retrieved into the database
func (lh *LocationHistory) Run() {
	lh.mqttClient.Connect()
	log.Debug("Connected to MQTT broker")

	for !lh.mqttClient.IsConnected() {
	}

	for {
		token := lh.mqttClient.Subscribe(lh.configuration.Mqtt.Topic, byte(1), lh.handleLocationMessage)
		if token.Wait() && token.Error() != nil {
			log.Errorf("Fail to subscribe... %v", token.Error())
			time.Sleep(5 * time.Second)

			log.Errorf("Retry to subscribe\n")
			continue
		} else {
			log.Info("Subscribed successfully!\n")
			break
		}
	}

	select {}
}

func (lh *LocationHistory) handleLocationMessage(client mqtt.Client, message mqtt.Message) {
	log.Debugf("Got message on topic '%s': %s\n", message.Topic(), message.Payload())
	var owntracksMessage OwntracksMessage
	json.Unmarshal(message.Payload(), &owntracksMessage)
	log.Infof("Received message with timestamp %s, lat %f, lon %f", owntracksMessage.Timestamp, owntracksMessage.Latitude, owntracksMessage.Longitude)
	lh.locationDatabase.AddWaypoint(message.Topic(), owntracksMessage.Latitude, owntracksMessage.Longitude, owntracksMessage.Timestamp.Time)
}

// UnmarshalJSON parses a unix timestamp into a UnixTime
func (unixtime *UnixTime) UnmarshalJSON(data []byte) (err error) {
	if string(data) == "null" {
		return nil
	}
	i, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	tm := time.Unix(i, 0)
	unixtime.Time = tm
	return
}
