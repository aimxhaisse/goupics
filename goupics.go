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
	"reflect"
)

// Flag settings
var (
	configPath = flag.String("c", "", "path to the configuration file (e.g, bean.json)")
	logPath    = flag.String("l", "", "path to the log file (e.g, bean.log)")
)

// Config contains the various configurations of our project
type Config struct {
	ListenOn    string // Interface to listen on
	Title	    string // Title of all pages
}

// DynamicHandlerFuncParams holds parameters dedicated to a page handler
type DynamicHandlerFuncParams struct {
	Router   *mux.Router        // Router of the request
	Template *template.Template // Template of the dynamic page
	Config   *Config            // Goupics' configuration
}

// Bean holds variables common to all handlers
type Bean struct {
	Config *Config     // Config holds the configuration of Bean
	Router *mux.Router // Router to all requests	
}

// PageParams holds parameters common to all pages
type PageParams struct {
	Name    string // Name of the page (home.html -> "home")
	Title   string // Title of the page
}

// HomePageParams holds parameters of the home page
type HomePageParams struct {
	PageParams
	Carousel []map[string]string	// Images to put in the carousel
}

// DynamicHandlerFunc serves a http request
type DynamicHandlerFunc func(*DynamicHandlerFuncParams, http.ResponseWriter, *http.Request)

// eq reports whether the first argument is equal to any of the remaining arguments.
// borrowed from a post from rsc
func eq(args ...interface{}) bool {
        if len(args) == 0 {
                return false
        }
        x := args[0]
        switch x := x.(type) {
        case string, int, int64, byte, float32, float64:
                for _, y := range args[1:] {
                        if x == y {
                                return true
                        }
                }
                return false
        }

        for _, y := range args[1:] {
                if reflect.DeepEqual(x, y) {
                        return true
                }
        }
        return false
}

// HomeHandler is a DynamicHandlerFunc that serves the home page
func HomeHandler(p *DynamicHandlerFuncParams, w http.ResponseWriter, r *http.Request) {
	var carousel_items []map[string]string

	file, err := ioutil.ReadFile("www/static/carousel/carousel.json")
	if err != nil {
		log.Printf("carousel error: %v", err)
	} else {
		err = json.Unmarshal(file, &carousel_items)
		if err != nil {
			log.Printf("carousel error: %v", err)
		}
	}

	err = p.Template.Execute(w, HomePageParams{PageParams{"home", p.Config.Title}, carousel_items})
	if err != nil {
		log.Printf("error while serving template home: %s", err)
	}
}

// BuildHandler maps DynamicHandlerFunc to http.HandlerFunc
func (b *Bean) BuildDynamicdHandler(handler DynamicHandlerFunc, name string) http.HandlerFunc {
	tpl := template.New(name)
	tpl.Funcs(template.FuncMap{"eq": eq})
	_, err := tpl.ParseFiles("www/dynamic/common.html", fmt.Sprintf("www/dynamic/%s.html", name))
	if err != nil {
		log.Fatalf("unable to parse template %s: %s", name, err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("serving template %s", name)
		handler(&DynamicHandlerFuncParams{b.Router, tpl, b.Config}, w, r)
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
	home := bean.BuildDynamicdHandler(HomeHandler, "home")
	bean.Router.HandleFunc("/", home)
	bean.Router.HandleFunc("/home.html", home)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("www/static"))))
	http.Handle("/", bean.Router)

	log.Printf("bean started, serving pages")
	log.Fatal(http.ListenAndServe(bean.Config.ListenOn, nil))
}
