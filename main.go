package main

import (
	"log"
	"time"
)

func main() {
	// Load in the config file
	config := loadConfig("config.yml")

	// Log into Hive
	err := hiveAuth(&config.Hive)
	if err != nil {
		panic(err)
	}

	// Set up the channel for IPC and do the actual polling
	influxChannel := make(chan influxDataPoint)

	// Loop over the devices
	go func() {
		// Set a ticker to keep control sending the responses
		ticker := time.NewTicker(time.Second * time.Duration(config.PollInterval))
		defer ticker.Stop()

		// Loop forever
		for range ticker.C {
			// Loop over the devices
			for id, nodetype := range config.Devices {
				// Get the data from Hive and send to influx
				tags, fields, err := hiveGetNode(config.Hive, id, nodetype)
				if err != nil {
					log.Println(id, err)
				}
				influxChannel <- influxDataPoint{nodetype, tags, fields}
			}
		}
	}()

	// Connect to the database and write results
	influxdb(config.Influx, influxChannel)

}
