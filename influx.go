package main

import (
	"log"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
)

var influxdbConnection client.Client

func influxdbLogin(config Influx) {
	//Start DB connection
	conn, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     config.Url,
		Username: config.User,
		Password: config.Pass,
	})
	if err != nil {
		log.Fatal(err)
	}
	influxdbConnection = conn
}

func influxDataPoint(config Influx, nodetype string, tags map[string]string, fields map[string]interface{}) {

	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  config.Database,
		Precision: "s",
	})
	if err != nil {
		log.Println(err)
		return
	}

	// Generate the data point
	pt, err := client.NewPoint(nodetype, tags, fields, time.Now())
	bp.AddPoint(pt)

	// Write the batch
	// err = influxdbConnection.Write(bp)
	if err != nil {
		log.Println(err)
		return
	}

}
