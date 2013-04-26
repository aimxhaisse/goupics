package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"log"
)

// Config contains the various configurations of our project
type Config struct {
	ListenOn string // Interface to listen on
}

// MyHandlerFunc serves a http request
type MyHandlerFunc func(*mux.Router, http.ResponseWriter, *http.Request)

// HomeHandler is a MyHandlerFunc that serves the home page
func HomeHandler(*mux.Router, http.ResponseWriter, *http.Request) {

}

// BuildHandler maps MyHandlerFunc to http.HandlerFunc
func BuildHandler(router *mux.Router, handler MyHandlerFunc) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		handler(router, w, r)
	}
}

// NewConfig creates a new configuration from the input path
func NewConfig(path string) *Config {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	var cfg Config
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	return &cfg
}

func main() {
	cfg := NewConfig("gow.json")
	r := mux.NewRouter()
	r.HandleFunc("/", BuildHandler(r, HomeHandler))
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(cfg.ListenOn, nil))
}
