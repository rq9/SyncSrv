package main

import (
	"bytes"
	"log"
	"net/http"
)

type testStruct struct {
	Test string
}

func test(rw http.ResponseWriter, req *http.Request) {

	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	s := buf.String()

	log.Printf(s)

}

func main() {
	http.HandleFunc("/test", test)
	log.Fatal(http.ListenAndServe(":8082", nil))
}
