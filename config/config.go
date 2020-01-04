package config

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// Config holds the main configuration information
type Config struct {
	Twitter Twitter
	Discord Discord
}

// Twitter holds the information for the twitter API
type Twitter struct {
	Token          string
	Secret         string
	ConsumerKey    string
	ConsumerSecret string
}

// Discord holds the information for the discord bot
type Discord struct {
	Token    string
	Username string
	Avatar   string
}

// NewConfig creates a new config struct for you based off of the json file
func NewConfig() *Config {
	b, err := ioutil.ReadFile("./config/config.json")
	if err != nil {
		log.Fatal("An error occured in obtaining configuration information: ", err)
	}
	config := &Config{}
	json.Unmarshal(b, config)

	// If an image link is given then change the config avatar property to base64 encoding
	if config.Discord.Avatar == "" || !strings.Contains(config.Discord.Avatar, "http") {
		return config
	}
	res, err := http.Get(config.Discord.Avatar)
	if err == nil {
		b, _ := ioutil.ReadAll(res.Body)
		config.Discord.Avatar = "data:image/png;base64," + base64.StdEncoding.EncodeToString(b)
	}
	res.Body.Close()

	return config
}
