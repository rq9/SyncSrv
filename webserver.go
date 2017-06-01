package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest" //authentication
	"gopkg.in/mgo.v2"                      //mongo db
)

func handleRequest(rw rest.ResponseWriter, req *rest.Request) {
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
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.Use(&rest.AuthBasicMiddleware{
		Realm: "test zone",
		Authenticator: func(userId string, password string) bool {
			if userId == "till" && password == "genesis" {
				return true
			}
			return false
		},
	})

	router, err := rest.MakeRouter(
		rest.Post("/sync", handleRequest),
		rest.Get("/sync", handleRequest),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServeTLS(":443", "server.pem", "server.key", api.MakeHandler()))
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
