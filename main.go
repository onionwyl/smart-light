package main

import (
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/onionwyl/smart-light/db"
	"github.com/onionwyl/smart-light/light"
	"github.com/onionwyl/smart-light/sensor"
	"github.com/onionwyl/smart-light/service"
	"log"
	"net/http"
	"os"
)

var DB *gorm.DB

func main() {
	if ok := light.Init(); !ok{
		log.Fatal("light init error")
		os.Exit(-1)
	}
	if ok := sensor.Init(); !ok{
		log.Fatal("sensor init error")
		os.Exit(-1)
	}
	if ok := db.Init(); !ok{
		log.Fatal("db init error")
		os.Exit(-1)
	}
	go detectUserAction()
	r := mux.NewRouter()
	r.HandleFunc("/", service.IndexHandler)
	r.HandleFunc("/search", service.SearchHandler)
	r.HandleFunc("/lights/{id:[0-9]+}", service.LightsInfoHandler).Methods("GET")
	r.HandleFunc("/lights/{id:[0-9]+}", service.LightsControlHandler).Methods("POST")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	err := http.ListenAndServe(":9090", r)
	if err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}
