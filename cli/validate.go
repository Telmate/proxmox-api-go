package cli

import (
	"fmt"
	"log"
	"strconv"
)

func ValidateIDset(args []string, indexPos int, text string) string {
	if indexPos+1 > len(args) {
		log.Fatal(fmt.Errorf("error: no %s has been provided", text))
	}
	return args[indexPos]
}

// Should be used for Optional IDs.
// returns if the indexd arg if it is set. It returns an empty string when the indexed arg is not set.
func OptionalIDset(args []string, indexPos uint) (out string) {
	if int(indexPos+1) <= len(args) {
		out = args[indexPos]
	}
	return
}

func ValidateIntIDset(args []string, text string) int {
	id, err := strconv.Atoi(ValidateIDset(args, 0, text))
	if err != nil && id <= 0 {
		log.Fatal(fmt.Errorf("error: %s must be a positive integer", text))
	}
	return id
}

func ValidateExistinGuestID(args []string, indexPos int) int {
	id, err := strconv.Atoi(ValidateIDset(args, indexPos, "GuestID"))
	if err != nil || id < 100 {
		log.Fatal(fmt.Errorf("error: GuestID must be a positive integer of 100 or greater"))
	}
	return id
}
