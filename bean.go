package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"flag"
)

// Flag settings
var (
	configPath = flag.String("c", "", "path to the configuration file (e.g, bean.json)")
	logPath = flag.String("l", "", "path to the log file (e.g, bean.log)")
)

// Config contains the various configurations of our project
type Config struct {
	ListenOn string // Interface to listen on
	ProjectName string // Name of the project
}

// DynamicHandlerFuncParams holds parameters dedicated to a page handler
type DynamicHandlerFuncParams struct {
	Router   *mux.Router        // Router of the request
	Template *template.Template // Template of the dynamic page
}

// Bean holds variables common to all handlers
type Bean struct {
	Config    *Config                       // Config holds the configuration of Bean
	Templates map[string]*template.Template // Templates is a cache of parsed templates
	Router    *mux.Router                   // Router to all requests	
}

// Parameters common to all pages
type PageParams struct {
	Name	string // Name of the page (home.html -> "home")
	Title	string // Title of the page
	Project string // Name of the project
}

type HomePageParams struct {
	PageParams
}

// DynamicHandlerFunc serves a http request
type DynamicHandlerFunc func(*DynamicHandlerFuncParams, http.ResponseWriter, *http.Request)

// HomeHandler is a DynamicHandlerFunc that serves the home page
func HomeHandler(p *DynamicHandlerFuncParams, w http.ResponseWriter, r *http.Request) {
	p.Template.Execute(w, &HomePageParams{PageParams{"home", "Title", "Project Name"}})
}

// BuildHandler maps DynamicHandlerFunc to http.HandlerFunc
func (g *Bean) BuildDynamicdHandler(handler DynamicHandlerFunc, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl, in_cache := g.Templates[path]
		if !in_cache {
			new_tpl, err := template.ParseFiles(fmt.Sprintf("www/dynamic/%s", path))
			if err != nil {
				log.Printf("unable to parse template %s: %s", path, err)
			}
			g.Templates[path] = new_tpl
			tpl = new_tpl
		}
		if tpl != nil {
			log.Printf("serving template %s", path)
			handler(&DynamicHandlerFuncParams{g.Router, tpl}, w, r)
		} else {
			log.Printf("template %s wasn't created", path)
		}
	}
}

// NewBean creates a new Bean from a configuration path
func NewBean(cfg_path, log_path string) *Bean {
	file, err := ioutil.ReadFile(cfg_path)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	var cfg Config
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	writer, err := os.OpenFile(log_path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Unable to open file log: %v", err)
	}
	log.SetOutput(writer)
	r := mux.NewRouter()
	return &Bean{
		&cfg,
		make(map[string]*template.Template),
		r,
	}
}

func main() {
	flag.Parse()
	bean := NewBean(*configPath, *logPath)
	bean.Router.HandleFunc("/", bean.BuildDynamicdHandler(HomeHandler, "home.html"))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("www/static"))))
	http.Handle("/", bean.Router)
	log.Fatal(http.ListenAndServe(bean.Config.ListenOn, nil))
}
