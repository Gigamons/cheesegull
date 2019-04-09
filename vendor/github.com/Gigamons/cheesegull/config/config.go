package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Config is the Configuration struct for our Config.
type Config struct {
	Server struct {
		ShouldDiscover   bool
		UnrankedBeatmaps bool
		Website          string
		BMCacheSize      float64
		Hostname         string
		Port             int
		Debug            bool
		RemoveNonZip     bool
	}
	Osu struct {
		DownloadHostname string // Good for CDN Support.
		APIKey           string
		Username         string
		Password         string
	}
	MySQL struct {
		Username string
		Password string
		Database string
		Hostname string
		Port     int
	}
	SphinxQL struct {
		Username string
		Password string
		Database string
		Hostname string
		Port     int
	}
}

// Parse the Config
func Parse() *Config {
	create()

	if _, err := os.Stat("./config.json"); os.IsNotExist(err) {
		panic("Error! Could not create Config!")
	}

	d, err := ioutil.ReadFile("./config.json")
	if err != nil {
		panic("Error! Could not load Config!")
	}

	conf := Config{}

	err = json.Unmarshal(d, &conf)
	if err != nil {
		panic("Error! Could not parse Config!")
	}

	return &conf
}

// Creates the Config
func create() {
	if _, err := os.Stat("./config.json"); os.IsNotExist(err) {
		var c Config
		c.MySQL.Port = 3306
		c.MySQL.Database = "cheesegull"
		c.MySQL.Hostname = "127.0.0.1"
		c.Server.ShouldDiscover = true
		c.Server.BMCacheSize = 10
		c.Server.Port = 62011
		c.Server.Hostname = "127.0.0.1"
		c.SphinxQL.Database = "cheesegull"
		c.SphinxQL.Hostname = "127.0.0.1"
		c.SphinxQL.Port = 9306
		c.Server.Website = fmt.Sprintf("<font face='Comic Sans MS' size='2'><h2>CHEESEGULL V2.1.5G*</h2></font><font face='Consolas' size='2'><b>Mirror Server:</b> Unknown <br><b>Maintainer:</b> <a href='mailto:UNKNOWN'>UNKNOWN</a><br><b>Original Source:</b> <a href='https://github.com/osuripple/cheesegull'>https://github.com/osuripple/cheesegull</a><br><b>Custom Fork:</b> <a href='https://github.com/gigamons/cheesegull'>https://github.com/gigamons/cheesegull</a><br><br><i>*: Using the custom gigamons fork.</i></font>")

		j, err := json.MarshalIndent(&c, "", "    ")
		if err != nil {
			panic(err)
		}

		err = ioutil.WriteFile("./config.json", j, 0644)
		if err != nil {
			panic(err)
		}

		fmt.Println("I've just created a config.json! please edit.")
		os.Exit(0)
	}
}
