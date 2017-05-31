package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"gopkg.in/mgo.v2"
)

func handleRequest(rw http.ResponseWriter, req *http.Request) {

	//try and fill buffer from request body
	buffer := new(bytes.Buffer)
	_, err := buffer.ReadFrom(req.Body)

	//if successful, push into JSON string
	if err == nil {
		var data string
		data = buffer.String()
		if isJSON(data) {
			//if JSON is valid, store in DB
			log.Printf(data)
			storeInDB(data)
			return
		} else {
			//if JSON is invalid, raise an error
			err = errors.New("No valid JSON")
		}
	}

	//if this is reached, the request was unsuccessful; print error
	log.Fatalf("error: %v", err.Error())
}

func storeInDB(data string) {
	//dial new db session
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}
	//close afterwards
	defer session.Close()

	//unmarshal json
	var js map[string]interface{}
	json.Unmarshal([]byte(data), &js)

	//add to collection sync in DB wellnomics
	c := session.DB("wellnomics").C("sync")
	err = c.Insert(&js)
	if err != nil {
		//Raise error
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/sync", handleRequest)
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

/*
git add .
git commit -m "MESSAGE"
git push -u origin master
*/
