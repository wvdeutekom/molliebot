package helpers

import (
	"math/rand"
	"sort"
	"time"
)

func RandomStringFromArray(array []string) string {

	rand.Seed(time.Now().UTC().UnixNano())
	return array[rand.Intn(len(array))]
}

func ArrayContainsString(array []string, searchString string) bool {

	sort.Strings(array)
	i := sort.SearchStrings(array, searchString)
	if i < len(array) && array[i] == searchString {
		return true
	}
	return false
}
