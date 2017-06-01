package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"gopkg.in/mgo.v2"
)

var successes = 0
var failures = 0
var requests = 0

func handleRequest(rw http.ResponseWriter, req *http.Request) {
	requests++

	//try and fill buffer from request body
	buffer := new(bytes.Buffer)
	_, err := buffer.ReadFrom(req.Body)

	if err == nil {
		//if successful, convert to JSON string
		var data string
		data = buffer.String()
		err = isJSON(data)
		if err == nil {
			//if JSON is valid, store in DB
			err = storeInDB(data)
			if err == nil {
				rw.WriteHeader(http.StatusOK)
				successes++
				return
			} else {
				//if persisting to DB failed, send appropriate status
				rw.WriteHeader(http.StatusTooManyRequests)
			}
		} else {
			//if JSON is invalid, send appropriate status
			rw.WriteHeader(http.StatusExpectationFailed)
		}
	} else {
		//if Request.Body is invalid, send appropriate status
		rw.WriteHeader(http.StatusBadRequest)
	}

	//if this is reached, the request was unsuccessful; print error
	failures++
	//log.Printf("error: %v (%v)", err.Error(), failures)

}

func storeInDB(data string) error {
	//dial new db session
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		//Raise error
		return err
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
		return err
	}
	return nil
}

func main() {
	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for t := range ticker.C {
			log.Printf("Time: %v", t)
			log.Printf("requests: %v", requests)
			log.Printf("successes: %v", successes)
			log.Printf("errors: %v", failures)
		}
	}()

	http.HandleFunc("/sync", handleRequest)
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func isJSON(s string) error {
	var js map[string]interface{}
	if json.Unmarshal([]byte(s), &js) == nil {
		return nil
	} else {
		return errors.New("Invalid JSON")
	}
}

/*
git add .
git commit -m "MESSAGE"
git push -u origin master
*/
