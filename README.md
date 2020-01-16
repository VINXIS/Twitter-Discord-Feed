# Twitter-Discord-Feed
A very simple setup of posting tweets from people onto a discord channel. May add more features later in life.

## Installation
 1. [Install golang](https://golang.org/doc/install). Ideally you have Go version 1.13 or newer. 
 2. Clone the repository using `git clone https://github.com/VINXIS/Twitter-Discord-Feed.git` to wherever you want.
 3. Go to the folder and open a console. Install the dependencies with `go get`.
 4. Go to the config folder and duplicate `config.example.json`. Name the duplicate `config.json` and fill in the twitter API credentials, and discord information
	 1. You can obtain twitter API credentials [here](https://developer.twitter.com/en/docs).
	 2. For the discord information. You add the discord bot token which is obtained from creating a discord bot [here](https://discordapp.com/developers/applications). Put the username as anything you want, preferably your bot's username, and put the avatar field to some image link, preferably the same image as your discord bot's.
 5. Invite the bot to your server by replacing `PUT_CLIENT_ID_HERE` in the URL below with the discord application's client ID obtained here [here](https://discordapp.com/developers/applications). https://discordapp.com/api/oauth2/authorize?client_id=PUT_CLIENT_ID_HERE&permissions=536870912&scope=bot.
 6. The bot by default **NEEDS** webhook manage permissions for the channel you want to follow twitter users on. The bot invite link above already does that for you, but if you would like to not have it create a role, replace `536870912` with `0`, and give the bot webhook manage permissions however you want to.
 7. Run the program by either running  `go build -o twitter core.go` and then `./twitter` in your instance / computer.
 8. Once running, type and send `~help` to a discord channel the bot is in for explanation regarding how to follow and unfollow twitter accounts
