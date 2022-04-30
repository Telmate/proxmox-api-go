package cli

import (
	"strconv"
	"fmt"
	"log"
)

func ValidateIDset(args []string, indexPos int, text string) (string){
	if indexPos+1 > len(args) {
		log.Fatal(fmt.Errorf("error: no %s has been provided", text))
	}
	return args[indexPos]
}

func ValidateExistinGuestID(args []string, indexPos int) (int){
	id, err := strconv.Atoi(ValidateIDset(args, indexPos, "GuestID"))
	if err != nil || id < 100 {
		log.Fatal(fmt.Errorf("error: GuestID must be a positive integer of 100 or greater"))
	}
	return id
}