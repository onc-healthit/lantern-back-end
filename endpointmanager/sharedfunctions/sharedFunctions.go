package sharedfunctions

import "log"

func failOnError(errString string, err error) {
	if err != nil {
		if errString == "" {
			log.Fatalf("%s", err)
		} else {
			log.Fatalf("%s %s", errString, err)
		}
	}
}
