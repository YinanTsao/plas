package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {

	sourceURL := "http://130.104.229.12:5000/endpoints"

	var endpoints Endpoints
	var err error

	for {
		endpoints, err = GetEndpoints(sourceURL)
		if err != nil {
			fmt.Printf("Failed to get endpoints: %v\n", err)
			fmt.Println("Retrying in 1 minutes...")
			time.Sleep(1 * time.Minute)
			continue
		}
		break
	}

	topology := new(NetworkTopology)

	go func() {
		for {
			*topology = Pinger(endpoints)
			fmt.Println(*topology)

			time.Sleep(3 * time.Minute)
		}
	}()

	http.HandleFunc("/networktopo", func(w http.ResponseWriter, r *http.Request) {
		NetworkTopoHandler(w, r, *topology)
	})

	addr := ":3300"
	fmt.Printf("HTTP server listening on%s\n", addr)
	http.ListenAndServe(addr, nil)
}
