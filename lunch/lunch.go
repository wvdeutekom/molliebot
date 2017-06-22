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
	lunchNotFoundMessages = []string{
		"404 Lunch not found",
		"There will be bread.",
		"Elementary, my dear Watson. It looks like bread.",
		"Keep your friends close, but your bread closer.",
		"Bread. Shaken, not stirred.",
		"We'll always have bread.",
		"They call it a royale with cheese. That means bread.",
		"Nothing on the menu, but I will have my lunch, in this life or the next.",
		"This bread seems somewhat familiar; have I eaten this before?",
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

func (lunches *Lunches) GetLunchMessageOfToday() string {

	var lunchMessage string
	lunchOfToday := lunches.getLunchOfToday()
	if lunchOfToday != nil {
		lunchMessage = "Today we eat: " + lunchOfToday.Description
	} else {
		lunchMessage = helpers.RandomStringFromArray(lunchNotFoundMessages)
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

func (lunches *Lunches) GetLunchMessageOfThisWeek() string {

	lunchMessage := "This week the following is on the menu:\n"
	availableLunch := lunches.getLunchOfThisWeek()
	if len(availableLunch) == 0 {
		lunchMessage = helpers.RandomStringFromArray(lunchNotFoundMessages)
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
