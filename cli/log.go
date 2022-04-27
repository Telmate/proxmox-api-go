package cli

import (
	"log"
)

func LogFatalError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func LogFatalListing(text string, err error){
	if err != nil {
		log.Fatalf("error listing %s %+v\n",text ,err)
	}
}