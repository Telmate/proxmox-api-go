package cli

import (
	"log"
)

func LogFatalError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}