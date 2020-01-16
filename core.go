package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"./config"
	"./discordhandle"
	"./twitterhandle"
	"github.com/ChimeraCoder/anaconda"
	"github.com/bwmarrin/discordgo"
)

func main() {
	// Obtain config and create twitter and discord API clients
	conf := config.NewConfig()
	twitterhandle.Config = conf
	twitterhandle.Twitter = anaconda.NewTwitterApiWithCredentials(conf.Twitter.Token, conf.Twitter.Secret, conf.Twitter.ConsumerKey, conf.Twitter.ConsumerSecret)
	discord, err := discordgo.New("Bot " + conf.Discord.Token)
	if err != nil {
		log.Fatal("An error occured in creating the discord client: ", err)
	}

	// Add handler and turn on discord client
	discord.AddHandler(discordhandle.Message)
	for {
		err = discord.Open()
		if err == nil {
			break
		}
	}
	log.Println("Twitter feed bot is now online.")

	// Twitter tracking
	twitterhandle.Track(discord)

	// Create a channel to keep the bot running until a prompt is given to close
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Kill)
	<-sc

	// Close sessions
	discord.Close()
	twitterhandle.Twitter.Close()
}
