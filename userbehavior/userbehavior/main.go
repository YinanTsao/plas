package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Payload struct {
	Data string `json:"data"`
}

func sendRequest(url string, jsonData []byte) {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Print the response status
	fmt.Println("Response status:", resp.Status)

	// Read and print the response body
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Response body:", string(body))
}

func main() {
	url := "http://130.104.229.12:31112/function/tngo"

	// Read the file into a string
	fileContent, err := os.ReadFile("image_base64")
	if err != nil {
		panic(err)
	}

	payload := Payload{
		Data: string(fileContent),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	for {
		sendRequest(url, jsonData)
		time.Sleep(2000 * time.Millisecond) // adjust here for different rates
	}
}
