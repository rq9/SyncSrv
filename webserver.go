package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/ant0ine/go-json-rest/rest" //authentication + ssl
	"gopkg.in/mgo.v2"                      //mongo db
	"gopkg.in/mgo.v2/bson"                 //bson
)

func handleGetRequest(rw rest.ResponseWriter, req *rest.Request) {
	rw.WriteJson(map[string]string{"body": "use POST https://localhost:433/sync, include authentication"})
}

func handlePostRequest(rw rest.ResponseWriter, req *rest.Request) {
	//try and fill buffer from request body
	buffer := new(bytes.Buffer)
	_, err := buffer.ReadFrom(req.Body)

	if err == nil {
		//if successful, convert to JSON string
		var data string
		data = buffer.String()
		err = isJSON(data)
		if err == nil {
			//find storageId for this user
			authHeader := req.Header.Get("Authorization")
			authToken := strings.Split(authHeader, "Basic ")[1]
			base64Text := make([]byte, base64.StdEncoding.DecodedLen(len(authToken)))
			base64.StdEncoding.Decode(base64Text, []byte(authToken))
			storageID := strings.Split(string(base64Text), ":")[0]
			//add storageId to JSON
			out := map[string]interface{}{}
			json.Unmarshal([]byte(data), &out)
			out["storageId"] = storageID
			outputJSON, err := json.Marshal(out)
			data = string(outputJSON)
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
	session, err := mgo.Dial("127.0.0.1") //TODO use connection pool
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

func authenticate(userID string, password string) bool {
	session, err := mgo.Dial("127.0.0.1") //TODO use connection pool
	if err != nil {
		//Raise error
		panic(err)
	}
	//close afterwards
	defer session.Close()

	c := session.DB("wellnomics").C("authentication")
	count, err := c.Find(bson.M{"name": userID, "password": password}).Count()
	if err != nil {
		panic(err)
	}

	return count > 0
}

func main() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.Use(&rest.AuthBasicMiddleware{
		Realm:         "test zone",
		Authenticator: authenticate,
	})

	router, err := rest.MakeRouter(
		rest.Post("/sync", handlePostRequest),
		rest.Get("/", handleGetRequest),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	go http.ListenAndServe(":8082", http.HandlerFunc(rejectHTTP))

	log.Fatal(http.ListenAndServeTLS(":443", "server.pem", "server.key", api.MakeHandler()))

}

func rejectHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusHTTPVersionNotSupported)
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
GIT
	git add .
	git commit -m "MESSAGE"
	git push -u origin master
GO
	go get
OPENSSL
	https://www.akadia.com/services/ssh_test_certificate.html
MONGODB
	db.authentication.insert({name:'till',password:'genesis',storageId:1})
	db.sync.count({"storageId":"till"})
*/
