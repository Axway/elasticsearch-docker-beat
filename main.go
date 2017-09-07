package main

import (
	"os"

	"github.com/freignat91/dbeat/beater"
	"github.com/elastic/beats/libbeat/beat"
)

func main() {
	err := beat.Run("dbeat", "", beater.New)
	if err != nil {
		os.Exit(1)
	}
}
