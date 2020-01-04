package twitterhandle

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"../config"
	"github.com/ChimeraCoder/anaconda"
	"github.com/bwmarrin/discordgo"
)

// Users is a list of user IDs to track
var Users []string

// Twitter is the twitter API client
var Twitter *anaconda.TwitterApi

// Config holds information regarding the config
var Config *config.Config

// Track tracks the twitter users for the channel every minute
func Track(s *discordgo.Session) {
	startTime := time.Now()

	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-ticker.C:
			// Get channels
			var fileNames []string
			filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
				fileNames = append(fileNames, path)
				return nil
			})
			var channels []string
			channelRegex, err := regexp.Compile(`(\d+)\.json`)
			for _, file := range fileNames {
				if channelRegex.MatchString(file) {
					channels = append(channels, channelRegex.FindStringSubmatch(file)[1])
				}
			}

			for _, chID := range channels {
				// Get webhook
				var webhook *discordgo.Webhook
				webhooks, _ := s.ChannelWebhooks(chID)
				if len(webhooks) == 0 {
					webhook, err = s.WebhookCreate(chID, Config.Discord.Username, Config.Discord.Avatar)
					if err != nil {
						s.ChannelMessageSend(chID, "**I need the MANAGE_WEBHOOKS persmission in order to post!**\nI will automatically remove tracking from this channel for now. Once you have given me webhook permissions for this channel, then please copypaste the `~follow` command originally used.")
						os.Remove("./" + chID + ".json")
						continue
					}
				} else {
					webhook = webhooks[0]
				}

				// Get users
				var users []anaconda.User
				b, _ := ioutil.ReadFile("./" + chID + ".json")
				json.Unmarshal(b, &users)
				for _, user := range users {
					// Get tweets
					v := url.Values{}
					v.Set("screen_name", user.ScreenName)
					v.Set("count", "200")
					tweets, err := Twitter.GetUserTimeline(v)
					if err != nil {
						continue
					}
					for _, tweet := range tweets {
						// Check if tweet was after last run
						tweetTime, _ := tweet.CreatedAtTime()
						if startTime.After(tweetTime) {
							continue
						}
						s.WebhookExecute(webhook.ID, webhook.Token, false, &discordgo.WebhookParams{
							Content:   "https://twitter.com/" + user.ScreenName + "/status/" + tweet.IdStr,
							Username:  user.ScreenName,
							AvatarURL: user.ProfileImageUrlHttps,
						})

					}
				}
			}
			startTime = time.Now()
		}
	}
}

// Verify checks if the twitter handles given are valid
func Verify(IDs []string) (users []anaconda.User, err error) {
	users, err = Twitter.GetUsersLookup(strings.Join(IDs, ", "), nil)
	if err != nil {
		return nil, err
	}

	// Remove protected accounts
	i := 0
	for _, user := range users {
		if !user.Protected {
			users[i] = user
			i++
		}
	}
	users = users[:i]
	return users, err
}
