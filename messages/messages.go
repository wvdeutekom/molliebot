package messages

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/nlopes/slack"
	"github.com/wvdeutekom/molliebot/helpers"
	"github.com/wvdeutekom/molliebot/lunch"
)

var (
	botNameRegex       = regexp.MustCompile(`^\bmollie(bot)?\b|\bmollie(bot)?\??$`)
	helpRegex          = regexp.MustCompile(`\bhelp\b`)
	lunchRegex         = regexp.MustCompile(`\blunch\w*|\beten\b|\beat\w*\b`)
	thisWeekRegex      = regexp.MustCompile(`\b(this|deze)\b\s+\bweek\b`)
	todayRegex         = regexp.MustCompile(`\bvandaag\b|\btoday\b`)
	userTagRegex       = regexp.MustCompile(`\<\@(.{9})\>`)
	goAwayRegex        = regexp.MustCompile(`(\bgo\b\s+\baway\b|\bleave\b|\bfuck\b\s+\boff\b)`)
	userIdRegex        = regexp.MustCompile(`\<\@|\>`)
	directMessageRegex = regexp.MustCompile(`^D(.{8})$`)
)

type Messages struct {
	api               *slack.Client
	References        references
	Channels          []string `mapstructure:"restricted_channels"`
	NotificationTimes []string `mapstructure:"notification_times"`
	Configuration     configuration
}

type configuration struct {
	VerboseLogging           bool
	ApiToken                 string
	RestrictToConfigChannels bool
}

type references struct {
	Lunch *lunch.Lunches
}

func (m *Messages) Setup(lunch *lunch.Lunches) {
	m.api = slack.New(m.Configuration.ApiToken)
	m.api.SetDebug(m.Configuration.VerboseLogging)
	m.References.Lunch = lunch
}

func (m *Messages) Monitor() {

	rtm := m.api.NewRTM()
	go rtm.ManageConnection()

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
			case *slack.TeamJoinEvent:
				// Handle new user to client
			case *slack.MessageEvent: //
				// Handle new message to channel

				// Only respond to real users. Bots have BotIDs, users do not
				if ev.Msg.BotID == "" {

					if m.Configuration.RestrictToConfigChannels == true {
						if helpers.ArrayContainsString(m.Channels, ev.Channel) {
							m.manageResponse(ev)
						}
					} else {
						m.manageResponse(ev)
					}
				}

			case *slack.ReactionAddedEvent:
				// Handle reaction added
			case *slack.ReactionRemovedEvent:
				// Handle reaction removed
			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())
			case *slack.InvalidAuthEvent:
				fmt.Printf("Invalid credentials")
				break Loop
			default:
				// fmt.Printf("Unknown error")
			}
		}
	}
}

func (m *Messages) manageResponse(msg *slack.MessageEvent) {

	// Get <@U12345> tag(s) from text and convert them to readable names
	userTags := userTagRegex.FindAllString(msg.Text, -1)
	for _, tag := range userTags {
		retrievedUsername := m.RetrieveSlackUsername(tag)
		msg.Text = strings.Replace(msg.Text, tag, retrievedUsername, -1)
	}

	// Sentence starts or ends with 'mollie' or 'molliebot' or is a direct message
	if botNameRegex.MatchString(msg.Text) || m.IsDirectMessage(msg) {
		trimmedText := botNameRegex.ReplaceAllString(msg.Text, "")

		//Handle help requests
		// Sentence contains 'help'
		if helpRegex.MatchString(trimmedText) == true {
			m.SendMessage("Need my help? Ask for lunch by asking along the lines of:\n"+
				"> Mollie what's for lunch today\n"+
				"> What are we having for lunch this week mollie\n"+
				"Or try asking me that in dutch, I'll probably listen.\n"+
				"\n"+
				"Suggestions, bugs? Create an issue on <https://github.com/wvdeutekom/molliebot|github.com>", msg.Channel)
		}

		//Handle general requests
		// Sentence contains 'go' and 'away'
		if goAwayRegex.MatchString(trimmedText) == true {

			m.SendMessage(fmt.Sprintf("I'm sorry %v, I'm afraid can't do that", m.RetrieveSlackUsername(msg.User)), msg.Channel)
		}

		// Handle lunch requests
		// Sentence contains 'lunch(ing,es)' or 'eten'
		if lunchRegex.MatchString(trimmedText) == true {

			switch {

			// Sentence contains 'this'/'deze' 'week'
			case thisWeekRegex.MatchString(trimmedText):

				lunchMessage := m.References.Lunch.GetLunchMessageOfThisWeek()
				m.SendMessage(lunchMessage, msg.Channel)
			default:
				// Sentence contains 'today'/'vandaag'
				//todayRegex.MatchString(trimmedText):

				lunchMessage := m.References.Lunch.GetLunchMessageOfToday()
				m.SendMessage(lunchMessage, msg.Channel)
			}
		}
	} else {
		fmt.Println("NO MATCHES AT ALL")
	}
}

func (m *Messages) RetrieveSlackUsername(userId string) string {

	// If userId contains <@ >, strip it from the string.
	if userTagRegex.MatchString(userId) {
		userId = userIdRegex.ReplaceAllString(userId, "")
	}

	user, error := m.api.GetUserInfo(userId)
	if error != nil {
		log.Print(error)
	}

	return user.Name
}

func (m *Messages) GetJoinedChannelsIDs() []string {

	// Get Public channels that the user/bot is part of
	allChannels := m.retrieveAllChannels()
	// Also get the private channels. Only joined private channels can be fetched
	allGroups := m.retrieveAllGroups()

	var joinedChannels []string

	// Loop through all the channels and add IDs to joinedChannels if IsMember
	for _, v := range allChannels {
		if v.IsMember {
			joinedChannels = append(joinedChannels, v.ID)
		}
	}

	// Add all group IDs to joinedChannels
	for _, v := range allGroups {
		joinedChannels = append(joinedChannels, v.ID)
	}
	return joinedChannels
}

func (m *Messages) retrieveAllChannels() []slack.Channel {
	channels, error := m.api.GetChannels(true)
	if error != nil {
		log.Print(error)
		return nil
	}
	return channels
}

func (m *Messages) retrieveAllGroups() []slack.Group {
	groups, error := m.api.GetGroups(true)
	if error != nil {
		log.Print(error)
		return nil
	}
	return groups
}

func (m *Messages) SendMessageToChannels(messageText string, channelIDs []string) {
	for _, channelID := range channelIDs {
		m.SendMessage(messageText, channelID)
	}
}

func (m *Messages) SendMessage(messageText string, channelId string) {
	params := slack.PostMessageParameters{
		AsUser: true,
	}
	footer := randomFooter()
	messageText += footer

	channelID, timestamp, err := m.api.PostMessage(channelId, messageText, params)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s\n", channelID, timestamp)
}

func (m *Messages) IsDirectMessage(msg *slack.MessageEvent) bool {
	return directMessageRegex.MatchString(msg.Channel)
}

func randomFooter() string {

	emojis := []string{
		"ヾ(⌐■_■)ノ♪",
		"ヽ(°◇° )ノ",
		"\\(^~^)/",
		"•ᴗ•",
		"(⌐■_■)",
		"(☞ﾟヮﾟ)☞",
		"(•‿•) ",
		"(」ﾟﾛﾟ)｣ ",
	}

	// Append some padding
	footerString := fmt.Sprintf("\n\n%v\n", helpers.RandomStringFromArray(emojis))

	return footerString
}
