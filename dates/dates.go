package dates

import (
	"fmt"
	"time"
)

type StringToDateOptions struct {
	Format string
}

func IsStringToday(stringDate string) bool {

	date := StringToDate(stringDate, StringToDateOptions{})
	return IsDateToday(date)
}

func StringToDate(stringDate string, options StringToDateOptions) time.Time {

	var dateFormat string
	if options.Format != "" {
		dateFormat = options.Format
	} else {
		dateFormat = "2006-01-02"
	}

	date, err := time.Parse(dateFormat, stringDate)
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
