package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Flag settings
var (
	configPath = flag.String("c", "", "path to the configuration file (e.g, bean.json)")
	logPath    = flag.String("l", "", "path to the log file (e.g, bean.log)")
)

// Config contains the various configurations of our project
type Config struct {
	ListenOn    string // Interface to listen on
	ProjectName string // Name of the project
}

// DynamicHandlerFuncParams holds parameters dedicated to a page handler
type DynamicHandlerFuncParams struct {
	Router   *mux.Router        // Router of the request
	Template *template.Template // Template of the dynamic page
}

// Bean holds variables common to all handlers
type Bean struct {
	Config *Config     // Config holds the configuration of Bean
	Router *mux.Router // Router to all requests	
}

// Parameters common to all pages
type PageParams struct {
	Name    string // Name of the page (home.html -> "home")
	Title   string // Title of the page
	Project string // Name of the project
}

type HomePageParams struct {
	PageParams
}

// DynamicHandlerFunc serves a http request
type DynamicHandlerFunc func(*DynamicHandlerFuncParams, http.ResponseWriter, *http.Request)

// HomeHandler is a DynamicHandlerFunc that serves the home page
func HomeHandler(p *DynamicHandlerFuncParams, w http.ResponseWriter, r *http.Request) {
	err := p.Template.Execute(w, HomePageParams{PageParams{"home", "Title", "Project Name"}})
	if err != nil {
		log.Printf("error while serving template home: %s", err)
	}
}

// BuildHandler maps DynamicHandlerFunc to http.HandlerFunc
func (b *Bean) BuildDynamicdHandler(handler DynamicHandlerFunc, name string) http.HandlerFunc {
	tpl, err := template.ParseFiles("www/dynamic/common.html", fmt.Sprintf("www/dynamic/%s.html", name))
	if err != nil {
		log.Fatalf("unable to parse template %s: %s", name, err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("serving template %s", name)
		handler(&DynamicHandlerFuncParams{b.Router, tpl}, w, r)
	}
}

// NewBean creates a new Bean from a configuration path
func NewBean(cfg_path, log_path string) *Bean {

	// load config file
	file, err := ioutil.ReadFile(cfg_path)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	var cfg Config
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// redirect logging to a file
	writer, err := os.OpenFile(log_path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Unable to open file log: %v", err)
	}
	log.SetOutput(writer)
	r := mux.NewRouter()

	result := &Bean{
		&cfg,
		r,
	}

	return result
}

func main() {
	flag.Parse()
	bean := NewBean(*configPath, *logPath)

	// register your dynamic routes here
	bean.Router.HandleFunc("/home.html", bean.BuildDynamicdHandler(HomeHandler, "home"))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("www/static"))))
	http.Handle("/", bean.Router)

	log.Printf("bean started, serving pages")

	log.Fatal(http.ListenAndServe(bean.Config.ListenOn, nil))
}
