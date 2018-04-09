package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", MainHandler)
	log.Println(http.ListenAndServe(":8080", nil))
}
