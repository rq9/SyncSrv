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
	//requests++

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
			for i := 0; i < 5; i++ {
				err = storeInDB(data)
				if err == nil {
					return
				} else {
					//retry n times if failed
					err = storeInDB(data)
				}
			}
			if err == nil {
				rw.WriteHeader(http.StatusOK)
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
	log.Printf("error: %v", err.Error())

}

func storeInDB(data string) error {
	//dial new db session
	var err error
	//var session *session
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
