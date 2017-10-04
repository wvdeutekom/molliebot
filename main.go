package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/nlopes/slack"
	"github.com/robfig/cron"
	"github.com/spf13/viper"
	"github.com/wvdeutekom/molliebot/schedules"
)

type AppContext struct {
	Message        *Messages         `mapstructure:"messages"`
	Lunch          *Lunches          `mapstructure:"lunch"`
	Schedule       *schedules.Client `mapstructure:"pagerduty"`
	Options        options
	ConfigLocation string
}

type options struct {
	DebugMode bool
}

var (
	appContext AppContext
)

func init() {
	// Read config file
	if appContext.ConfigLocation = os.Getenv("CONFIG_LOCATION"); appContext.ConfigLocation == "" {
		log.Println("No CONFIG_LOCATION environment variable set. Using default: './config.json'")
		appContext.ConfigLocation = "./config.json"
	}

	// TODO: should extract path from configLocation string
	viper.SetConfigFile(appContext.ConfigLocation)

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("No configuration file loaded: %v\n", err)
	}

	// Read config into appContext struct
	err = viper.Unmarshal(&appContext)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	// Read environment variables
	// API_KEY
	var apiToken string
	if apiToken = os.Getenv("API_KEY"); apiToken == "" {
		log.Fatalln("No API_KEY environment variable set")
	}

	// DEBUG
	var debugMode bool
	debugModeString := os.Getenv("DEBUG")
	if debugModeString == "" {
		log.Println("no DEBUG environment variable set. Using default: 'false'")
		debugMode = false
	} else {
		debugMode, err = strconv.ParseBool(debugModeString)
		if err != nil {
			log.Fatalln(err)
		}
	}

	// Pagerduty
	viper.BindEnv("PAGERDUTY_API_KEY")
	pagerdutyApiKey := viper.Get("PAGERDUTY_API_KEY").(string)

	appContext.Message.Configuration.ApiToken = apiToken
	appContext.Options.DebugMode = debugMode
	appContext.Message.Configuration.VerboseLogging = debugMode
	appContext.Schedule = schedules.New(pagerdutyApiKey, appContext.ConfigLocation)
}

func main() {
	fmt.Println("starting bot")

	logger := log.New(os.Stdout, "messages-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)

	appContext.Lunch.Setup()
	appContext.Message.Setup(&appContext)

	appContext.startCrons()
	appContext.Message.Monitor()
}

func (context *AppContext) startCrons() {

	cron := cron.New()

	cron.AddFunc("0 */10 * * * *", func() {
		context.Schedule.GetCurrentOnCallUsers()
	})

	cron.AddFunc("0 1 11 18 * *", func() {
		reportMessage := context.Schedule.CompileScheduleReport()

		// Send message to report_channels
		for _, reportChannel := range appContext.Schedule.ReportChannels {
			context.Message.SendMessage(reportMessage, reportChannel)
		}
	})

	for _, cronTime := range context.Message.NotificationTimes {
		fmt.Println("adding cron ", cronTime)
		cron.AddFunc(cronTime, func() {
			lunchMessage := context.Lunch.GetLunchMessageOfToday(true)

			joinedChannelIDs := context.Message.GetJoinedChannelsIDs()
			context.Message.SendMessageToChannels(lunchMessage, joinedChannelIDs)
		})

	}
	cron.Start()
}
