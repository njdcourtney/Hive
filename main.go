package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	// Load in the config file
	config := loadConfig("config.yml")

	// Authenticate to Hive
	hiveLogin(config.Hive)

	// Connect to influxDB
	influxdbLogin(config.Influx)

	// Loop over the devices
	// Set a ticker to keep control sending the responses
	ticker := time.NewTicker(time.Second * time.Duration(config.PollInterval))
	defer ticker.Stop()

	// Loop forever
	for range ticker.C {
		// Loop over the devices
		for id, nodetype := range config.Devices {
			// Get the data from Hive and send to influx
			fmt.Println("Pre fetch")
			tags, fields, err := hiveGetNode(config.Hive, id, nodetype)
			if err != nil {
				log.Println(id, err)
				continue // Skip the rest of the loop
			}
			// //Send the data to influx
			influxDataPoint(config.Influx, nodetype, tags, fields)
		}
	}
}
