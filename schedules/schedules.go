package schedules

import (
	"fmt"
	"log"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/spf13/viper"
)

type Context struct {
	apiKey      string
	client      *pagerduty.Client
	Schedules   []pagerduty.Schedule
	onCallUsers []pagerduty.User
}

//TODO:
// * Chat: Get persons on call for a given date. Return the entire week and all the people on call in every team. USE client.ListOnCalls
// * Administration: collect the entire pagerduty schedule from pagerduty. Make a list and send it to @wijnand every month

// var authtoken = "P_-iGverwWGC7S7UgrsQ" // Set your auth token here

func init() {
	fmt.Println("init pagerduty!")
	viper.BindEnv("PAGERDUTY_API_KEY")

	var context Context
	context.apiKey = viper.Get("PAGERDUTY_API_KEY").(string)
	context.client = pagerduty.NewClient(context.apiKey)
}

func (context *Context) GetCurrentOnCallUsers() []pagerduty.User {

	context.getAllSchedules(false)

	var onCallUsers []pagerduty.User

	for _, schedule := range context.Schedules {
		var onCallOpts pagerduty.ListOnCallUsersOptions
		var currentTime = time.Now()
		onCallOpts.Since = currentTime.Format("2006-01-02T15:04:05Z07:00")
		hours, _ := time.ParseDuration("1s")
		onCallOpts.Until = currentTime.Add(hours).Format("2006-01-02T15:04:05Z07:00")

		if eps, err := context.client.ListOnCallUsers(schedule.ID, onCallOpts); err != nil {
			panic(err)
		} else {
			for _, user := range eps {
				user = context.getUserInfo(user.ID)
				fmt.Println(user)
				onCallUsers = append(onCallUsers, user)
			}
		}
	}

	return onCallUsers
}

func (context *Context) getUserInfo(userID string) pagerduty.User {
	if user, err := context.client.GetUser(userID, pagerduty.GetUserOptions{}); err != nil {
		panic(err)
	} else {
		return *user
	}
}
func (context *Context) updatePagerdutyChannels() {

	for _, schedule := range context.Schedules {
		fmt.Println(schedule.FinalSchedule)
	}
}

func (context *Context) getAllSchedules(withDetail bool) {
	var c chan pagerduty.Schedule = make(chan pagerduty.Schedule)
	context.getScheduleList()

	if withDetail {
		for _, schedule := range context.Schedules {
			go context.getSchedule(schedule, c)
		}
		context.storeSchedules(c)
	}
}

func (context *Context) storeSchedules(c <-chan pagerduty.Schedule) {
	for i, _ := range context.Schedules {
		schedule := <-c
		context.Schedules[i] = schedule
	}
}

func (context *Context) getScheduleList() {
	fmt.Println("getScheduleLIST: ")

	if eps, err := context.client.ListSchedules(pagerduty.ListSchedulesOptions{}); err != nil {
		panic(err)
	} else {
		context.Schedules = eps.Schedules
	}
}

func (context *Context) getSchedule(schedule pagerduty.Schedule, c chan<- pagerduty.Schedule) {

	fmt.Println("START! --------------------", schedule.ID)
	if schedule, err := context.client.GetSchedule(schedule.ID, pagerduty.GetScheduleOptions{}); err != nil {
		log.Fatal(err)
	} else {
		c <- *schedule
	}
}
