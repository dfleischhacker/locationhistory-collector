package configuration

import (
	"io/ioutil"

	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

// The MqttConfig defines the MQTT server used to retrieve location data
type MqttConfig struct {
	// URL to connect to MQTT broker
	URL string
	// Topic to listen for location updates
	Topic string
	// Username to use when connecting
	Username string
	// Password for the user
	Password string
}

// The DatabaseConfig defines the database used to store location data
type DatabaseConfig struct {
	// dsn to connect to database
	DriverName string
	Dsn        string
}

// The Configuration of the locationhistory app
type Configuration struct {
	Mqtt     MqttConfig
	Database DatabaseConfig
	Map      MapConfig
}

// The MapConfig used for showing the waypoint map on the web UI
type MapConfig struct {
	Token       string
	BindAddress string
	Port        int
}

// LoadConfiguration loads a config file from the given path and returns the resulting Configuration
func LoadConfiguration(path string) *Configuration {
	log.Debugln("Trying to load data from path {}", path)

	fileContent, err := ioutil.ReadFile(path)

	if err != nil {
		log.Fatal(err)
	}

	config := Configuration{}
	toml.Unmarshal(fileContent, &config)

	return &config
}
