package parser

import (
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// As each flight level in represented as Fxxx (e.g. F350), this function simply parse the flight level to an int with an eventual error (if the int conversion fails for instance)
func extractFlightLevel(in string) (int, error) {
	fl, err := strconv.Atoi(in[1:])

	if err != nil {
		log.Errorf("flight level %v cannot be parsed", fl)
		return -1, fmt.Errorf("flight level %v cannot be parsed", fl)
	}

	return fl, err
}
