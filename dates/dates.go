package dates

import (
	"time"
	"fmt"
)

func IsStringToday(stringDate string) bool {

	date := StringToDate(stringDate)
	return IsDateToday(date)
}

func StringToDate(stringDate string) time.Time {

	date, err := time.Parse("2006-01-02", stringDate)
	if err != nil {
		fmt.Println(err)
	}
	return date
}

func IsDateToday(date time.Time) bool {

	var isToday bool
	today := time.Now().Local()

	if date.Year() == today.Year() && date.Month() == today.Month() && date.Day() == today.Day() {
		isToday = true
	}

	return isToday
}

func NumberOfTheWeekInMonth(now time.Time) int {
	beginningOfTheMonth := time.Date(now.Year(), now.Month(), 1, 1, 1, 1, 1, time.UTC)
	_, thisWeek := now.ISOWeek()
	_, beginningWeek := beginningOfTheMonth.ISOWeek()
	return 1 + thisWeek - beginningWeek
}
