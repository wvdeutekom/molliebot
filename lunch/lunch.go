package lunch

import (
	"fmt"
	"log"
	"time"

	"github.com/grsmv/goweek"
	"github.com/wvdeutekom/molliebot/dates"
	"github.com/wvdeutekom/molliebot/helpers"
)

var (
	introMessages = []string{
		"Hi there! Are you as excited about lunch as I am? Let me see what's on the menu.",
		"Hola, it's almost time for lunch! Let's see what I can hustle up.",
	}

	//LNF means Lunch Not Found
	midLNFMessages = []string{
		"Hmmm. I'm not sure what's on the menu today, all I can say is:",
		"Don't see anything on the menu today, but my logs say:",
		"Hmmm. I Couldn't find any lunchings. Slogan time!",
		"No special lunch found, can someone insert my batteries? -BEEP-",
		"Where is that lunch? Maybe you should try turning me off and on again",
	}

	postLNFMessages = []string{
		`"404 Lunch not found" - Mollie monolith backend`,
		`"There will be bread." - a bread fanatic`,
		`"Elementary, my dear Watson. It looks like bread." - Mollie Holmes, probably.`,
		`"Keep your friends close, but your bread closer." - Sun Tzu`,
		`"Bread. Shaken, not stirred." - James Bread`,
		"We'll always have bread.",
		`"They call it a royale with cheese. That means bread."`,
		`"Nothing on the menu, but I will have my lunch, in this life or the next." - Me. 100%`,
		`"This bread seems somewhat familiar; have I eaten this before?" - Captain Jack Sparrow`,
		`"I'll always have bread, bread with peanutbutter." - Tjeerd`,
	}
)

type Lunches struct {
	Lunches []Lunch `mapstructure:"lunches"`
}

// TODO: Dirty DateString solution, if you  manage to unmarshal this struct
// with Viper directly into a time.Time type be my guest!
type Lunch struct {
	DateTime    time.Time
	DateString  string `mapstructure:"date"`
	Description string `mapstructure:"description"`
}

func (lunches *Lunches) ConvertLunchStringsToDate() {
	for i := 0; i < len(lunches.Lunches); i++ {
		lunches.Lunches[i].DateTime = dates.StringToDate(lunches.Lunches[i].DateString)
	}
}

// GetLunchMessageOfToday Get the lunch message today.
// If introduction is set to true then a short introduction message will be prepended
func (lunches *Lunches) GetLunchMessageOfToday(introduction bool) string {

	var lunchMessage string
	lunchOfToday := lunches.getLunchOfToday()
	if lunchOfToday != nil {
		lunchMessage = "Today we eat: " + lunchOfToday.Description
	} else {
		if introduction {
			lunchMessage += helpers.RandomStringFromArray(introMessages)
			lunchMessage += "\n"
		}
		lunchMessage += helpers.RandomStringFromArray(midLNFMessages)
		lunchMessage += "\n\n"
		lunchMessage += helpers.RandomStringFromArray(postLNFMessages)
	}
	return lunchMessage
}

func (lunches *Lunches) getLunchOfToday() *Lunch {

	for _, lunch := range lunches.Lunches {
		if dates.IsDateToday(lunch.DateTime) {
			return &lunch
		}
	}
	return nil
}

// GetLunchMessageOfThisWeek Get the lunch message for this week.
// If introduction is set to true then a short introduction message will be prepended
func (lunches *Lunches) GetLunchMessageOfThisWeek(introduction bool) string {

	lunchMessage := "This week the following is on the menu:\n"
	availableLunch := lunches.getLunchOfThisWeek()
	if len(availableLunch) == 0 {
		if introduction {
			lunchMessage += helpers.RandomStringFromArray(introMessages)
			lunchMessage += "\n"
		}
		lunchMessage += helpers.RandomStringFromArray(midLNFMessages)
		lunchMessage += "\n\n"
		lunchMessage += helpers.RandomStringFromArray(postLNFMessages)
	} else {
		for _, lunch := range availableLunch {
			lunchMessage += fmt.Sprintf("%v: %v\n", lunch.DateTime.Weekday(), lunch.Description)
		}
	}

	return lunchMessage
}

func (lunches *Lunches) getLunchOfThisWeek() []Lunch {

	week, err := goweek.NewWeek(time.Now().ISOWeek())
	if err != nil {
		log.Fatal("could not create NewWeek")
	}

	// Loop over each weekday
	// This should be refactored so its done only once and stored in a struct
	var lunchesThisWeek []Lunch
	for _, day := range week.Days {

		//Loop over each meal
		for _, lunch := range lunches.Lunches {
			if lunch.DateTime == day {
				lunchesThisWeek = append(lunchesThisWeek, lunch)
			}
		}
	}
	return lunchesThisWeek
}
