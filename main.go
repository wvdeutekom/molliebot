package main

import (
	"./dates"
	"encoding/json"
	"fmt"
	"github.com/grsmv/goweek"
	"github.com/nlopes/slack"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"time"
)

type Token struct {
	Token string `json:"token"`
}

type Lunch struct {
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
}

type Config struct {
	Lunch []Lunch `json:"lunch"`
}

var (
	api       *slack.Client
	botToken  Token
	config    Config
	channelId string

	botRgx      = regexp.MustCompile(`^\bgobot|\bgobot$`)
	helpRgx     = regexp.MustCompile(`\bhelp\b`)
	lunchRgx    = regexp.MustCompile(`\blunch\w*|\beten\b`)
	thisWeekRgx = regexp.MustCompile(`\bdeze\b\s+\bweek\b|\bthis\b+s\bweek\b`)
	todayrgx    = regexp.MustCompile(`\bvandaag\b|\btoday\b`)
)

func (u *Lunch) UnmarshalJSON(data []byte) error {
	type Alias Lunch
	aux := &struct {
		Date string `json:"date"`
		*Alias
	}{
		Alias: (*Alias)(u),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	u.Date = dates.StringToDate(aux.Date)
	return nil
}

func init() {
	file, err := ioutil.ReadFile("./api_key.json")
	if err != nil {
		log.Fatal("Cannot read config.json")
	}

	if err := json.Unmarshal(file, &botToken); err != nil {
		log.Fatal("Cannot unmarshal json file")
	}

	raw, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if err := json.Unmarshal(raw, &config); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Printf("config: %v\n", config)
}

func getLunchThisWeek() []Lunch {
	fmt.Printf("This is day %d, week number: %d in the month\n", time.Now().Day(), dates.NumberOfTheWeekInMonth(time.Now()))

	week, err := goweek.NewWeek(time.Now().ISOWeek())
	if err != nil {
		log.Fatal("could not create NewWeek")
	}

	// Loop over each weekday
	// This should be refactored so its done only once and stored in a struct
	var lunchesToday []Lunch
	for _, day := range week.Days {

		//Loop over each meal
		for _, lunch := range config.Lunch {
			if lunch.Date == day {
				fmt.Printf("this week we eat: %v\n", lunch.Description)
				lunchesToday = append(lunchesToday, lunch)
			}
		}
	}
	return lunchesToday
}

func main() {
	fmt.Println("starting bot")

	api = slack.New(botToken.Token)
	channelId = "C594N2UHG"

	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	api.SetDebug(true)

	rtm := api.NewRTM()
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

				if ev.Msg.BotID == "" {
					manageResponse(ev)
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

func manageResponse(msg *slack.MessageEvent) {

	if msg.Channel == channelId {

		if botRgx.MatchString(msg.Text) {
			// Sentence starts or ends with 'gobot'
			trimmedText := botRgx.ReplaceAllString(msg.Text, "")

			fmt.Printf("TRIMMED TEXT: %s\n", trimmedText)

			//Handle help requests
			// Sentence contains 'help'
			if helpRgx.MatchString(trimmedText) == true {
				sendMessage("Need my help? Ask for lunch by asking:\n> gobot wat eten we vandaag", "")
			}

			// Handle lunch requests
			// Sentence contains 'lunch(ing,es)' or 'eten'
			if lunchRgx.MatchString(trimmedText) == true {

				switch {
				// Sentence contains 'this'/'deze' 'week'
				case thisWeekRgx.MatchString(trimmedText):

					lunchMessage := "This week the following is on the menu:\n"
					for _, lunch := range getLunchThisWeek() {
						lunchMessage += fmt.Sprintf("%v: %v\n", lunch.Date.Weekday(), lunch.Description)
					}
					fmt.Printf("LUNCH MESSAGE %v\n", lunchMessage)
					sendMessage(lunchMessage, "")

				// Sentence contains 'today'/'vandaag'
				case todayRgx.MatchString(trimmedText):

					for _, lunch := range config.Lunch {
						if dates.IsDateToday(lunch.Date) {
							message := "Today we eat: " + lunch.Description
							sendMessage(message, "")
						}
					}
				}
			}
		} else {
			fmt.Println("NO MATCHES AT ALL")
		}
	}
}

func sendMessage(messageText string, subMessage string) {
	params := slack.PostMessageParameters{}
	footer := "\n\nヽ(°◇° )ノ\n"
	messageText += footer

	if subMessage != "" {
		attachment := slack.Attachment{
			Text: subMessage,
		}
		params.Attachments = []slack.Attachment{attachment}
	}

	// C594N2UHG = devtest channel
	channelID, timestamp, err := api.PostMessage(channelId, messageText, params)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}