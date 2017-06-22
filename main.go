package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/nlopes/slack"
	"github.com/spf13/viper"
	"github.com/wvdeutekom/molliebot/lunch"
	"github.com/wvdeutekom/molliebot/messages"
)

type AppContext struct {
	Message *messages.Messages `mapstructure:"messages"`
	Lunch   *lunch.Lunches     `mapstructure:"lunch"`
	Options options
}

type options struct {
	DebugMode                bool
	RestrictToConfigChannels bool
}

var (
	appContext AppContext
)

func init() {
	// Read config file
	var configLocation string
	if configLocation = os.Getenv("CONFIG_LOCATION"); configLocation == "" {
		log.Println("No CONFIG_LOCATION environment variable set. Using default: './config.json'")
		configLocation = "./config.json"
	}

	// TODO: should extract path from configLocation string
	viper.AddConfigPath("./")
	viper.SetConfigFile(configLocation)

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

	// RESTRICT_TO_CONFIG_CHANNELS
	var restrictToConfigChannels bool
	restrictToConfigChannelsString := os.Getenv("RESTRICT_TO_CONFIG_CHANNELS")
	if restrictToConfigChannelsString == "" {
		log.Println("No RESTRICT_TO_CONFIG_CHANNELS environment variable set. Using default: 'false")
		restrictToConfigChannels = false
	} else {
		restrictToConfigChannels, err = strconv.ParseBool(restrictToConfigChannelsString)
		if err != nil {
			log.Fatalln(err)
		}
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

	appContext.Message.Configuration.ApiToken = apiToken
	appContext.Options.DebugMode = debugMode
	appContext.Message.Configuration.VerboseLogging = debugMode
	appContext.Options.RestrictToConfigChannels = restrictToConfigChannels
}

func main() {
	fmt.Println("starting bot")

	appContext.Lunch.ConvertLunchStringsToDate()

	logger := log.New(os.Stdout, "messages-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)

	appContext.Message.Setup(appContext.Lunch)
	appContext.Message.Monitor()
}
