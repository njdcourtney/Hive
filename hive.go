package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func hiveAuth(hiveconfig *Hive) error {
	// Format the url and the POST JSON.
	url := fmt.Sprintf("%s/omnia/auth/sessions", hiveconfig.Url)
	authStr := []byte(fmt.Sprintf("{\"sessions\": [{ \"username\": \"%s\", \"password\": \"%s\", \"caller\": \"WEB\"}]}",
		hiveconfig.User,
		hiveconfig.Pass))

	// Get the response from Hive
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(authStr))
	body, err := hiveSendRest(req, "")
	if err != nil {
		log.Println(err)
		return err
	}

	// Finally, unmarshall the JSON and return the result.
	var result struct {
		Sessions []struct {
			Id string `json:"id"`
		} `json:"sessions"`
	}
	json.Unmarshal(body, &result)

	// Update the config with the session id
	hiveconfig.SessionId = result.Sessions[0].Id
	return nil

}

func hiveGetNode(hiveconfig Hive, nodeId string, nodeType string) (map[string]string, map[string]interface{}, error) {
	// Format the url
	url := fmt.Sprintf("%s/omnia/nodes/%s", hiveconfig.Url, nodeId)

	// Get the response from Hive
	req, err := http.NewRequest("GET", url, nil)
	body, err := hiveSendRest(req, hiveconfig.SessionId)
	if err != nil {
		return nil, nil, err
	}

	// Define the struct to decode the JSON.
	var result struct {
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

	// Unmarshall the JSON.
	json.Unmarshal(body, &result)

	// Define the Influx tags
	tags := map[string]string{
		"name": result.Nodes[0].Name,
		"id":   nodeId,
	}

	// The data to return depends on the type of device
	switch nodeType {
	case "thermostat":
		fields := map[string]interface{}{
			"current": result.Nodes[0].Attributes.Temperature.ReportedValue,
			"target":  result.Nodes[0].Attributes.TargetTemperature.ReportedValue,
		}
		return tags, fields, nil
	case "radiator":
		fields := map[string]interface{}{
			"current": result.Nodes[0].Attributes.Temperature.ReportedValue,
		}
		return tags, fields, nil
	default:
		return nil, nil, errors.New(fmt.Sprintf("Unknown device type %s", nodeType))
	}

}

// Shared function for sending the REST call
func hiveSendRest(req *http.Request, session string) ([]byte, error) {
	// Add the required headers
	req.Header.Set("Content-Type", "application/vnd.alertme.zoo-6.1+json")
	req.Header.Set("Accept", "application/vnd.alertme.zoo-6.1+json")
	req.Header.Set("X-Omnia-Client", "Hive Web Dashboard")
	// Add the sess header if already authenticated
	if session != "" {
		req.Header.Set("X-Omnia-Access-Token", session)
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
