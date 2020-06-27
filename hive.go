package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var hiveSession string = ""

// Authenticate to Hive now and every half an hour
func hiveLogin(hiveconfig Hive) {

	// Authenticate immediately
	hiveAuth(hiveconfig)

	go func() {
		// Set a ticker
		authTicker := time.NewTicker(time.Minute * 30)
		defer authTicker.Stop()

		// Reauthenticate every ticker interval
		for range authTicker.C {
			hiveAuth(hiveconfig)
		}
	}()
}

func hiveAuth(hiveconfig Hive) {

	// Format the url and the POST JSON.
	url := fmt.Sprintf("%s/omnia/auth/sessions", hiveconfig.Url)
	authStr := []byte(fmt.Sprintf("{\"sessions\": [{ \"username\": \"%s\", \"password\": \"%s\", \"caller\": \"WEB\"}]}",
		hiveconfig.User,
		hiveconfig.Pass))

	// Get the response from Hive
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(authStr))
	body, err := hiveSendRest(req)
	if err != nil {
		log.Fatal("Error Authenticating", err)
	}

	// Finally, unmarshall the JSON and return the result.
	var result struct {
		Sessions []struct {
			Id string `json:"id"`
		} `json:"sessions"`
	}
	json.Unmarshal(body, &result)

	// Update the session varaible with the new session id
	hiveSession = result.Sessions[0].Id

}

// Map out the JSON structure in the returned data
type HiveJsonStructure struct {
	Nodes []struct {
		Id         string `json:"id"`
		Name       string `json:"name"`
		Attributes struct {
			Temperature struct {
				ReportedValue float32 `json:"reportedValue"`
				DisplayValue  float32 `json:"displayValue"`
			} `json:"temperature"`
			TargetTemperature struct {
				ReportedValue float32 `json:"reportedValue"`
				DisplayValue  float32 `json:"displayValue"`
			} `json:"targetHeatTemperature"`
		} `json:"attributes"`
	} `json:"nodes"`
}

func hiveGetNode(hiveconfig Hive, nodeId string, nodeType string) (tags map[string]string, fields map[string]interface{}, err error) {
	// Generic error handler (https://blog.golang.org/defer-panic-and-recover)
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("Recovered from panic in hiveGetNode %s", r))
			return
		}
	}()

	// Format the url
	url := fmt.Sprintf("%s/omnia/nodes/%s", hiveconfig.Url, nodeId)

	// Get the response from Hive
	req, err := http.NewRequest("GET", url, nil)
	body, err := hiveSendRest(req)
	if err != nil {
		return nil, nil, err
	}

	// Unmarshall the JSON.
	var result HiveJsonStructure
	json.Unmarshal(body, &result)

	// Define the Influx tags
	tags = map[string]string{
		"name": result.Nodes[0].Name,
		"id":   nodeId,
	}
	// The data to return depends on the type of device
	switch nodeType {
	case "thermostat":
		fields = map[string]interface{}{
			"current": result.Nodes[0].Attributes.Temperature.ReportedValue,
			"target":  result.Nodes[0].Attributes.TargetTemperature.ReportedValue,
		}
	case "radiator":
		fields = map[string]interface{}{
			"current": result.Nodes[0].Attributes.Temperature.ReportedValue,
		}
	default:
		err = errors.New(fmt.Sprintf("Unknown device type %s", nodeType))
		return nil, nil, err
	}

	return tags, fields, err
}

// Shared function for sending the REST call
func hiveSendRest(req *http.Request) ([]byte, error) {
	// Add the required headers
	req.Header.Set("Content-Type", "application/vnd.alertme.zoo-6.1+json")
	req.Header.Set("Accept", "application/vnd.alertme.zoo-6.1+json")
	req.Header.Set("X-Omnia-Client", "Hive Web Dashboard")
	// Add the sess header if already authenticated
	if hiveSession != "" {
		req.Header.Set("X-Omnia-Access-Token", hiveSession)
	}

	// Send the actual REST request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read in the request body.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Return the body
	return body, nil

}
