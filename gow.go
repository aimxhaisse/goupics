package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
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

// Gow holds variables common to all handlers
type Gow struct {
	Config    *Config                       // Config holds the configuration of Gow
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
func (g *Gow) BuildDynamicdHandler(handler DynamicHandlerFunc, path string) http.HandlerFunc {
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

// NewGow creates a new Gow from a configuration path
func NewGow(path string) *Gow {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	var cfg Config
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	r := mux.NewRouter()

	return &Gow{
		&cfg,
		make(map[string]*template.Template),
		r,
	}
}

func main() {
	gow := NewGow("gow.json")
	gow.Router.HandleFunc("/", gow.BuildDynamicdHandler(HomeHandler, "home.html"))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("www/static"))))
	http.Handle("/", gow.Router)
	log.Fatal(http.ListenAndServe(gow.Config.ListenOn, nil))
}
