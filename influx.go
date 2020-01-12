package main

import (
	"log"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
)

type influxDataPoint struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func influxdb(config Influx, dataChannel chan influxDataPoint) {
	//Start DB connection
	conn, _ := client.NewHTTPClient(client.HTTPConfig{
		Addr:     config.Url,
		Username: config.User,
		Password: config.Pass,
	})
	defer conn.Close()

	// Loop listening to the channel
	for {
		// Get the result from the channel
		datapoint := <-dataChannel

		// Create a new point batch
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  config.Database,
			Precision: "s",
		})
		if err != nil {
			log.Println(err)
		}

		pt, err := client.NewPoint(datapoint.name, datapoint.tags, datapoint.fields, time.Now())
		bp.AddPoint(pt)

		// Write the batch
		err = conn.Write(bp)
		if err != nil {
			log.Println(err)
		}

	}

}
