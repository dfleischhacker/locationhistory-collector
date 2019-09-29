package main

import (
	"encoding/json"
	"fmt"
	"github.com/dfleischhacker/locationhistory-collector/importer"
	"github.com/dfleischhacker/locationhistory-collector/utils"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"github.com/urfave/cli"

	locationhistory "github.com/dfleischhacker/locationhistory-collector/locationdb"

	"github.com/dfleischhacker/locationhistory-collector/configuration"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

func main() {
	var configFile string
	var debug bool
	var history LocationHistory

	app := cli.NewApp()
	app.Name = "lohico"
	app.Version = "0.0.1"
	app.Usage = "Collects OwnTracks MQTT messages"
	app.ArgsUsage = "[config file]"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config,c",
			Required:    true,
			Usage:       "Load configuration file from `FILE`",
			Destination: &configFile,
		},
		cli.BoolFlag{
			Name:        "debug,d",
			Usage:       "Enable debug output",
			Destination: &debug,
		},
	}

	app.Before = func(c *cli.Context) error {
		if debug {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}
		history = NewLocationHistory(configFile)
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:  "run",
			Usage: "Start the history collector",
			Action: func(c *cli.Context) error {
				history.Run()
				return nil
			},
		},
		{
			Name:  "export",
			Usage: "Exports the SQLite database to JSON",
			Action: func(c *cli.Context) error {
				return nil
			},
		},
		{
			Name:  "topics",
			Usage: "List all topics for which location data is stored in the database",
			Action: func(c *cli.Context) error {
				topics, err := history.locationDatabase.GetTopics()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("There is data available for the following topics:\n")
				for _, topic := range topics {
					fmt.Println(" - " + topic)
				}
				return nil
			},
		},
		{
			Name:      "query",
			Usage:     "Writes all waypoints for the given `TOPIC` into the given `FILE`",
			ArgsUsage: "TOPIC FILE",
			Action: func(c *cli.Context) error {
				if c.NArg() != 2 {
					return cli.NewExitError("Provide both TOPIC and FILE parameter", -2)
				}
				topic := c.Args().Get(0)
				fileName := c.Args().Get(1)
				waypoints, err := history.locationDatabase.GetWaypoints(topic)
				if err != nil {
					return err
				}
				data, err := json.MarshalIndent(waypoints, "", " ")
				if err == nil {
					err = ioutil.WriteFile(fileName, data, 0644)
				}
				return err
			},
		},
		{
			Name:      "import",
			Usage:     "Imports a Google Timeline from the given export (JSON) `FILE` for the given `TOPIC`",
			ArgsUsage: "TOPIC FILE",
			Action: func(c *cli.Context) error {
				if c.NArg() != 2 {
					return cli.NewExitError("Provide both TOPIC and FILE parameter", -3)
				}
				topic := c.Args().Get(0)
				fileName := c.Args().Get(1)
				count, err := importer.ImportTimeline(&history.locationDatabase, fileName, topic)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Imported %d waypoints", count)
				return nil
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

// LocationHistory contains all relevant data of the locationhistory program
type LocationHistory struct {
	configuration    *configuration.Configuration
	mqttClient       mqtt.Client
	locationDatabase locationhistory.LocationDatabase
}

// OwntracksMessage represents a message sent by Owntracks
type OwntracksMessage struct {
	Battery              int            `json:"batt"`
	Longitude            float64        `json:"lon"`
	Latitude             float64        `json:"lat"`
	Accuracy             int            `json:"acc"`
	Pressure             float64        `json:"p"`
	BatteryState         int            `json:"bs"`
	VerticalAccuracy     int            `json:"vac"`
	Trigger              string         `json:"t"`
	InternetConnectivity string         `json:"conn"`
	Timestamp            utils.UnixTime `json:"tst"`
	Altitude             int            `json:"alt"`
	TrackerID            string         `json:"tid"`
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
	lh.locationDatabase.AddWaypointData(message.Topic(), owntracksMessage.Latitude, owntracksMessage.Longitude, owntracksMessage.Timestamp.Time)
}
