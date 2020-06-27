package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Hive struct {
	Url    string
	User   string
	Pass   string
	Reauth int
}

type Influx struct {
	Url      string
	Database string
	User     string
	Pass     string
}

type ConfigFile struct {
	Hive         Hive
	Influx       Influx
	Devices      map[string]string
	PollInterval int
}

// Load in the config file.
func loadConfig(filename string) ConfigFile {
	// Deliberately don't catch errors, want script to crash on config file load error
	configFile, _ := ioutil.ReadFile(filename)
	var config ConfigFile
	yaml.Unmarshal(configFile, &config)

	// Return the struct
	return (config)
}
