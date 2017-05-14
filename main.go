package main

import (
	"encoding/json"
	"fmt"
	"github.com/grsmv/goweek"
	"github.com/nlopes/slack"
	"github.com/wvdeutekom/molliebot/dates"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"
)

type Lunch struct {
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
}

type Config struct {
	Lunch                    []Lunch  `json:"lunch"`
	Channels                 []string `json:"channels"`
	RestrictToConfigChannels bool
}

var (
	api      *slack.Client
	apiToken string
	config   Config

	botNameRgx  = regexp.MustCompile(`^\b(mol|mollie)|\b(mol|mollie)\??$`)
	helpRgx     = regexp.MustCompile(`\bhelp\b`)
	lunchRgx    = regexp.MustCompile(`\blunch\w*|\beten\b|\beating\b`)
	thisWeekRgx = regexp.MustCompile(`\b(this|deze)\b\s+\bweek\b`)
	todayRgx    = regexp.MustCompile(`\bvandaag\b|\btoday\b`)

	lunchNotFoundMessages = []string{
		"404 Lunch not found",
		"There will be bread.",
		"Elementary, my dear Watson. It looks like bread.",
		"Keep your friends close, but your bread closer.",
		"Bread. Shaken, not stirred.",
		"We'll always have bread.",
		"They call it a royale with cheese. That means bread.",
		"Nothing on the menu, but I will have my lunch, in this life or the next.",
	}

	emojis = []string{
		"ヾ(⌐■_■)ノ♪",
		"ヽ(°◇° )ノ",
		"\\(^~^)/",
		"•ᴗ•",
		"(⌐■_■)",
		"(☞ﾟヮﾟ)☞",
		"(•‿•) ",
		"(」ﾟﾛﾟ)｣ ",
	}
)

func (l *Lunch) UnmarshalJSON(data []byte) error {
	type Alias Lunch
	aux := &struct {
		Date string `json:"date"`
		*Alias
	}{
		Alias: (*Alias)(l),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	l.Date = dates.StringToDate(aux.Date)
	return nil
}

func init() {

	// Read config file
	var configLocation string
	if configLocation = os.Getenv("CONFIG_LOCATION"); configLocation == "" {
		log.Println("No CONFIG_LOCATION environment variable set. Using default: './config.json'")
		configLocation = "./config.json"
	}

	raw, err := ioutil.ReadFile(configLocation)
	if err != nil {
		log.Fatalln(err.Error())
	}

	if err := json.Unmarshal(raw, &config); err != nil {
		log.Fatalln(err.Error())
	}
	fmt.Printf("config: %v\n", config)

	// Read environment variables
	if apiToken = os.Getenv("API_KEY"); apiToken == "" {
		log.Fatalln("No API_KEY environment variable set")
	}

	restrictToConfigChannelsString := os.Getenv("RESTRICT_TO_CONFIG_CHANNELS")
	if restrictToConfigChannelsString == "" {
		log.Println("No RESTRICT_TO_CONFIG_CHANNELS environment variable set. Using default: 'false")
		config.RestrictToConfigChannels = false
	} else {
		config.RestrictToConfigChannels, err = strconv.ParseBool(restrictToConfigChannelsString)
		if err != nil {
			log.Fatalln(err)
		}
	}

}

func main() {
	fmt.Println("starting bot")

	api = slack.New(apiToken)

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

					if config.RestrictToConfigChannels == true {
						if arrayContainsString(config.Channels, ev.Channel) {
							manageResponse(ev)
						}
					} else {
						manageResponse(ev)
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

func arrayContainsString(array []string, searchString string) bool {

	sort.Strings(array)
	i := sort.SearchStrings(array, searchString)
	if i < len(array) && array[i] == searchString {
		return true
	}
	return false
}

func manageResponse(msg *slack.MessageEvent) {

	// Sentence starts or ends with 'mollie' or 'mol'
	if botNameRgx.MatchString(msg.Text) {
		trimmedText := botNameRgx.ReplaceAllString(msg.Text, "")

		fmt.Printf("TRIMMED TEXT: %s\n", trimmedText)

		//Handle help requests
		// Sentence contains 'help'
		if helpRgx.MatchString(trimmedText) == true {
			sendMessage("Need my help? Ask for lunch by asking along the lines of:\n"+
				"> gobot what's for lunch today\n"+
				"> what are we having for lunch this week gobot\n"+
				"Or try asking me that in dutch, I'll probably listen.", msg.Channel)
		}

		//Handle general requests
		// Sentence contains 'go' and 'away'
		goAwayRgx := regexp.MustCompile(`(\bgo\b\s+\baway\b|\bleave\b|\bfuck\b\s+\boff\b)`)
		if goAwayRgx.MatchString(trimmedText) == true {

			sendMessage(fmt.Sprintf("I'm sorry %v, I'm afraid can't do that", retrieveSlackUsername(msg.User)), msg.Channel)
		}

		// Handle lunch requests
		// Sentence contains 'lunch(ing,es)' or 'eten'
		if lunchRgx.MatchString(trimmedText) == true {
			log.Println("lunch match")

			switch {

			// Sentence contains 'this'/'deze' 'week'
			case thisWeekRgx.MatchString(trimmedText):

				lunchMessage := "This week the following is on the menu:\n"
				for _, lunch := range getLunchThisWeek() {
					lunchMessage += fmt.Sprintf("%v: %v\n", lunch.Date.Weekday(), lunch.Description)
				}
				fmt.Printf("LUNCH MESSAGE %v\n", lunchMessage)
				sendMessage(lunchMessage, msg.Channel)

			default:
				// Sentence contains 'today'/'vandaag'
				//todayRgx.MatchString(trimmedText):

				lunchFound := false
				for _, lunch := range config.Lunch {
					if dates.IsDateToday(lunch.Date) {
						message := "Today we eat: " + lunch.Description
						sendMessage(message, msg.Channel)
						lunchFound = true
					}
				}
				if !lunchFound {

					sendMessage(randomStringFromArray(lunchNotFoundMessages), msg.Channel)
				}
			}
		}
	} else {
		fmt.Println("NO MATCHES AT ALL")
	}
}

func retrieveSlackUsername(userId string) string {
	user, error := api.GetUserInfo(userId)
	if error != nil {
		log.Fatalln(error)
	}

	return user.Name
}

func sendMessage(messageText string, channelId string) {
	params := slack.PostMessageParameters{}
	footer := randomFooter()
	messageText += footer

	channelID, timestamp, err := api.PostMessage(channelId, messageText, params)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}

func randomStringFromArray(array []string) string {

	rand.Seed(time.Now().UTC().UnixNano())
	return array[rand.Intn(len(array))]
}

func randomFooter() string {

	rand.Seed(time.Now().UTC().UnixNano())

	// Append some padding
	footerString := fmt.Sprintf("\n\n%v\n", randomStringFromArray(emojis))

	return footerString
}
