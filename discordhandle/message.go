package discordhandle

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"../twitterhandle"
	"github.com/ChimeraCoder/anaconda"
	"github.com/bwmarrin/discordgo"
)

// Message looks for a `~follow` or `~help` command sent by users
func Message(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore ALL bot messages
	if m.Author.Bot {
		return
	}

	// Create regexes
	followRegex, _ := regexp.Compile(`~follow\s+(.+)`)
	followInfoRegex, _ := regexp.Compile(`~followinfo`)
	followRemoveRegex, _ := regexp.Compile(`~followremove\s+(.+)`)
	helpRegex, _ := regexp.Compile(`~help`)

	// Check for a command
	if helpRegex.MatchString(m.Content) || (len(m.Mentions) > 0 && m.Mentions[0].ID == s.State.User.ID) {
		help(s, m)
	} else if followRegex.MatchString(m.Content) {
		follow(s, m)
	} else if followInfoRegex.MatchString(m.Content) {
		followInfo(s, m)
	} else if followRemoveRegex.MatchString(m.Content) {
		followRemove(s, m)
	}
	return
}

// help sends a message detailing how to use the bot discord-client side
func help(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://github.com/VINXIS/Twitter-Discord-Feed",
			IconURL: s.State.User.AvatarURL(""),
			Name:    s.State.User.Username,
		},
		Description: "This bot is used to follow twitter accounts and post any new tweet they make including retweets.\n" +
			"To follow an account, simply use the following format: `~follow <link to user|userhandle>`\n" +
			"To follow multiple accounts, simply separate the links/user handles with a space.\n" +
			"To remove accounts, use `followremove` instead of `follow`. It works the same way.\n" +
			"To see which accounts are being followed: simply post `~followinfo`.",
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name: "Example of **following** accounts",
				Value: "~follow https://twitter.com/Twitter https://twitter.com/TwitterAPI?s=09 https://twitter.com/TwitterDev\n" +
					"~follow https://twitter.com/jack\n" +
					"~follow jack twitter",
			},
			&discordgo.MessageEmbedField{
				Name: "Example of **unfollowing** accounts",
				Value: "~followremove https://twitter.com/Twitter https://twitter.com/TwitterAPI?s=09 https://twitter.com/TwitterDev\n" +
					"~followremove https://twitter.com/jack\n" +
					"~followremove jack twitter",
			},
		},
	})
}

func follow(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Parse message for accounts
	followRegex, _ := regexp.Compile(`~follow\s+(.+)`)
	userIDs, err := parseMessage(s, m, followRegex)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	// If 0 users were found then return
	if len(userIDs) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No user IDs given! Please link the accounts, or give the user handle!")
		return
	}

	// Load users
	users := loadUsers(m)

	// Get users
	newUsers, err := twitterhandle.Verify(userIDs)
	if err != nil || len(newUsers) == 0 {
		s.ChannelMessageSend(m.ChannelID, "The users **"+strings.Join(userIDs, ", ")+"** either do not exist, or are protected.")
		return
	}

	// Filter out duplicates, and append new users to list
	for _, newUser := range newUsers {
		exists := false
		for _, user := range users {
			if user.ScreenName == newUser.ScreenName {
				exists = true
				break
			}
		}

		if !exists {
			users = append(users, newUser)
		}
	}

	// Save users
	saveUsers(users, m)

	text := "Now following: "
	for _, user := range users {
		text += "**" + user.ScreenName + ",** "
	}
	s.ChannelMessageSend(m.ChannelID, strings.TrimSuffix(text, ",** ")+"**. Please make sure I have the **MANAGE_WEBHOOKS** persmission to post via webhooks!\nI will still be able to post without the permission given.")
}

func followInfo(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Load users
	users := loadUsers(m)

	// Check if there even is tracking occurring for the channel
	if len(users) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No tracking for this channel currently!")
		return
	}

	// List out users
	text := "This channel is following: "
	for _, user := range users {
		text += "**" + user.ScreenName + ",** "
	}
	text = strings.TrimSuffix(text, ",** ")+"**"

	s.ChannelMessageSend(m.ChannelID, text)
}

func followRemove(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Parse message for accounts
	followRemoveRegex, _ := regexp.Compile(`~followremove\s+(.+)`)
	userIDs, err := parseMessage(s, m, followRemoveRegex)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	// If 0 users were found then return
	if len(userIDs) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No user IDs given! Please link the accounts, or give the user handle!")
		return
	}

	// Load users
	users := loadUsers(m)
	if len(users) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No users being tracked in this channel currently!")
		return
	}

	// Create initial text
	baseText := "Removed the following people: "
	text := baseText

	// Remove users
	var newUsers []anaconda.User
	for _, user := range users {
		remove := false
		for _, ID := range userIDs {
			if strings.ToLower(ID) == strings.ToLower(user.ScreenName) {
				text += "**" + user.ScreenName + ",** "
				remove = true
				break
			}
		}

		if !remove {
			newUsers = append(newUsers, user)
		}
	}

	if text == baseText {
		s.ChannelMessageSend(m.ChannelID, "Was not able to remove anyone with the given IDs!")
		return
	}

	if len(newUsers) != 0 {
		saveUsers(newUsers, m)
	} else {
		os.Remove("./" + m.ChannelID + ".json")
	}
	s.ChannelMessageSend(m.ChannelID, strings.TrimSuffix(text, ",** ")+"**")
}

// loadUsers loads a list of twitter users from a json file
func loadUsers(m *discordgo.MessageCreate) (users []anaconda.User) {
	_, err := os.Stat("./" + m.ChannelID + ".json")
	if err == nil {
		b, _ := ioutil.ReadFile("./" + m.ChannelID + ".json")
		json.Unmarshal(b, &users)
	}
	return users
}

// saveUsers saves a list of twitter users into a json file
func saveUsers(users []anaconda.User, m *discordgo.MessageCreate) {
	jsonCache, _ := json.Marshal(users)
	ioutil.WriteFile("./"+m.ChannelID+".json", jsonCache, 0644)
}

// parseMessage parses the discord message for accounts
func parseMessage(s *discordgo.Session, m *discordgo.MessageCreate, regex *regexp.Regexp) (userIDs []string, err error) {
	// Admin check
	admin := false
	if m.GuildID == "" {
		admin = true
	} else {
		server, err := s.Guild(m.GuildID)
		if err != nil {
			return nil, errors.New("could not get server information")
		}
		admin = adminCheck(s, m, *server)
	}

	if !admin {
		return nil, errors.New("you must either be an owner, administrator, or server manager")
	}

	// Get twitter accounts
	userList := regex.FindStringSubmatch(m.Content)[1]
	IDs := strings.Split(userList, " ")
	twitterRegex, _ := regexp.Compile(`twitter\.com/([a-zA-Z0-9_]+)`)
	nameRegex, _ := regexp.Compile(`[a-zA-Z0-9_]+`)
	for _, ID := range IDs {
		if twitterRegex.MatchString(ID) {
			userIDs = append(userIDs, twitterRegex.FindStringSubmatch(ID)[1])
		} else if nameRegex.MatchString(ID) {
			userIDs = append(userIDs, nameRegex.FindStringSubmatch(ID)[0])
		}
	}
	return userIDs, nil
}

// adminCheck is a utility function for checking if the user is an admin
func adminCheck(s *discordgo.Session, m *discordgo.MessageCreate, server discordgo.Guild) (admin bool) {
	if m.Author.ID == server.OwnerID {
		admin = true
	} else {
		member, _ := s.GuildMember(server.ID, m.Author.ID)
		for _, roleID := range member.Roles {
			role, err := s.State.Role(m.GuildID, roleID)
			if err == nil && (role.Permissions&discordgo.PermissionAdministrator != 0 || role.Permissions&discordgo.PermissionManageServer != 0) {
				admin = true
				break
			}
		}
	}
	return admin
}
