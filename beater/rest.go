package beater

import (
	"fmt"
	"log"
	"net/http"
)

const baseURL = "/api/v1"

// Start API server
func (a *dbeat) initAPI() {
	log.Printf("Start REST server on port %d", a.config.RESTPort)
	go func() {
		http.HandleFunc(baseURL+"/health", a.agentHealth)
		err := http.ListenAndServe(fmt.Sprintf(":%d", a.config.RESTPort), nil)
		if err != nil {
			log.Fatalln("Unable to start REST server: ", err)
		}
	}()
}

// or HEALTHCHECK Dockerfile instruction
func (a *dbeat) agentHealth(resp http.ResponseWriter, req *http.Request) {
	if a.eventStreamReading {
		resp.WriteHeader(200)
	} else {
		log.Println("Error: health check failed")
		resp.WriteHeader(400)
	}
}
