package schedules

import (
	"fmt"
	"log"
	"time"

	pagerduty "github.com/wvdeutekom/go-pagerduty"
)

type Client struct {
	pagerdutyClient *pagerduty.Client
	Schedules       []pagerduty.Schedule
	onCallUsers     []pagerduty.User
}

//TODO:
// * Chat: Get persons on call for a given date. E.g. "Who is on call next week", return the entire week and all the people on call in every team. USE client.ListOnCalls
// * Administration: collect the entire pagerduty schedule from pagerduty. Make a list and send it to @wijnand every month

func New(apiKey string) *Client {
	client := &Client{}
	client.pagerdutyClient = pagerduty.NewClient(apiKey)
	return client
}

func (client *Client) GetCurrentOnCallUsersMessage() string {
	onCallMessage := "Currently on call:\n"

	var users []pagerduty.User

	// For now, do not fetch new data every time this function is called. New data is fetched every 10 minutes in main.go to update this data.
	if len(client.onCallUsers) > 0 {
		users = client.onCallUsers
	} else {
		users = client.GetCurrentOnCallUsers()
	}

	for _, user := range users {
		if len(user.Teams) > 0 {
			onCallMessage = onCallMessage + user.Teams[0].APIObject.Summary + ": "
		}
		onCallMessage = onCallMessage + user.Name + " - " + client.extractContactAddressFromContactMethods(user.ContactMethods, "phone_contact_method") + "\n"
	}

	return onCallMessage
}

func (client *Client) GetCurrentOnCallUsers() []pagerduty.User {

	client.getAllSchedules(false)

	var onCallUsers []pagerduty.User

	for _, schedule := range client.Schedules {
		var onCallOpts pagerduty.ListOnCallUsersOptions
		var currentTime = time.Now()
		onCallOpts.Since = currentTime.Format("2006-01-02T15:04:05Z07:00")
		hours, _ := time.ParseDuration("1s")
		onCallOpts.Until = currentTime.Add(hours).Format("2006-01-02T15:04:05Z07:00")

		if eps, err := client.pagerdutyClient.ListOnCallUsers(schedule.ID, onCallOpts); err != nil {
			panic(err)
		} else {
			for _, user := range eps {
				user = client.getUserInfo(user.ID)
				user.ContactMethods = client.GetUserContactMethods(user.ID)
				onCallUsers = append(onCallUsers, user)
			}
		}
	}
	client.onCallUsers = onCallUsers
	return onCallUsers
}

func (client *Client) getUserInfo(userID string) pagerduty.User {
	if user, err := client.pagerdutyClient.GetUser(userID, pagerduty.GetUserOptions{}); err != nil {
		panic(err)
	} else {
		return *user
	}
}

func (client *Client) GetUserContactMethods(userID string) []pagerduty.ContactMethod {
	if contactMethodResponse, err := client.pagerdutyClient.GetUserContactMethod(userID); err != nil {
		panic(err)
	} else {
		return contactMethodResponse.ContactMethods
	}
}

func (client *Client) extractContactAddressFromContactMethods(userContactMethods []pagerduty.ContactMethod, contactType string) string {

	//Phonenumber is a string because it could potentially start with "+31"
	for _, contactMethod := range userContactMethods {
		if contactMethod.Type == contactType {
			return contactMethod.Address
		}
	}
	return ""
}

func (client *Client) updatePagerdutyChannels() {

	for _, schedule := range client.Schedules {
		fmt.Println(schedule.FinalSchedule)
	}
}

func (client *Client) getAllSchedules(withDetail bool) {
	var c chan pagerduty.Schedule = make(chan pagerduty.Schedule)
	client.getScheduleList()

	if withDetail {
		for _, schedule := range client.Schedules {
			go client.getSchedule(schedule, c)
		}
		client.storeSchedules(c)
	}
}

func (client *Client) storeSchedules(c <-chan pagerduty.Schedule) {
	for i, _ := range client.Schedules {
		schedule := <-c
		client.Schedules[i] = schedule
	}
}

func (client *Client) getScheduleList() {
	if eps, err := client.pagerdutyClient.ListSchedules(pagerduty.ListSchedulesOptions{}); err != nil {
		panic(err)
	} else {
		client.Schedules = eps.Schedules
	}
}

func (client *Client) getSchedule(schedule pagerduty.Schedule, c chan<- pagerduty.Schedule) {

	if schedule, err := client.pagerdutyClient.GetSchedule(schedule.ID, pagerduty.GetScheduleOptions{}); err != nil {
		log.Fatal(err)
	} else {
		c <- *schedule
	}
}
