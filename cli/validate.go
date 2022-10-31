package cli

import (
	"fmt"
	"log"
	"strconv"
)

// Should be used for Required IDs.
// returns if the indexed arg if it is set. It throws and error when the indexed arg is not set.
func RequiredIDset(args []string, indexPos uint, text string) string {
	if int(indexPos+1) > len(args) {
		log.Fatal(fmt.Errorf("error: no %s has been provided", text))
	}
	return args[indexPos]
}

// Should be used for Optional IDs.
// returns if the indexed arg if it is set. It returns an empty string when the indexed arg is not set.
func OptionalIDset(args []string, indexPos uint) (out string) {
	if int(indexPos+1) <= len(args) {
		out = args[indexPos]
	}
	return
}

func ValidateIntIDset(args []string, text string) int {
	id, err := strconv.Atoi(RequiredIDset(args, 0, text))
	if err != nil && id <= 0 {
		log.Fatal(fmt.Errorf("error: %s must be a positive integer", text))
	}
	return id
}

func ValidateExistingGuestID(args []string, indexPos uint) int {
	id, err := strconv.Atoi(RequiredIDset(args, indexPos, "GuestID"))
	if err != nil || id < 100 {
		log.Fatal(fmt.Errorf("error: GuestID must be a positive integer of 100 or greater"))
	}
	return id
}
