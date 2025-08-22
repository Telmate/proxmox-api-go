package cli

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/Telmate/proxmox-api-go/proxmox"
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

func ValidateGuestIDset(args []string, text string) proxmox.GuestID {
	id, err := strconv.Atoi(RequiredIDset(args, 0, text))
	if err != nil && id <= 0 {
		log.Fatal(fmt.Errorf("error: %s must be a positive integer", text))
	}
	if id > proxmox.GuestIdMaximum {
		log.Fatal(errors.New(proxmox.GuestID_Error_Maximum))
	}
	guestID := proxmox.GuestID(id)
	if err := guestID.Validate(); err != nil {
		log.Fatal(err)
	}
	return guestID
}
