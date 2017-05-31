package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func handleRequest(rw http.ResponseWriter, req *http.Request) {

	//try and fill buffer from request body
	buffer := new(bytes.Buffer)
	_, err := buffer.ReadFrom(req.Body)

	//if successful, push into json string
	if err == nil {
		var json string
		json = buffer.String()
		if isJSON(json) {
			log.Printf(json)
			return
		} else {
			err = errors.New("No valid JSON")
		}
	}

	//if this is reached, it is unsuccessful; print error
	log.Printf("error: %v", err.Error())
}

func main() {
	http.HandleFunc("/sync", handleRequest)
	log.Fatal(http.ListenAndServe(":8082", nil))
}
