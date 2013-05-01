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
	"path/filepath"
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
 	Carousel []map[string]string// Images to put in the carousel
}

// Bean holds variables common to all handlers
type Bean struct {
	Config *Config     // Config holds the configuration of Bean
	Router *mux.Router // Router to all requests	
 	Carousel []map[string]string// Images to put in the carousel
}

// PageParams holds parameters common to all pages
type PageParams struct {
	Name    string // Name of the page (home.html -> "home")
	Title   string // Title of the page
}

// HomePageParams holds parameters of the home page
type HomePageParams struct {
	PageParams
 	Carousel []map[string]string// Images to put in the carousel
}

// GalleriesPageParams holds parameters of the galleries page
type GalleriesPageParams struct {
	PageParams
	Galleries []Gallery	// Galleries to display
 	Carousel []map[string]string// Images to put in the carousel
}

// GalleryPageParams holds parameters of the gallery page
type GalleryPageParams struct {
	PageParams
 	Carousel []map[string]string// Images to put in the carousel
	Gallery string	// Gallery to display
	Images []string // Images in the gallery
}

// ImagePageParams holds parameters of the image page
type ImagePageParams struct {
	PageParams
 	Carousel []map[string]string// Images to put in the carousel
	Gallery string // Gallery of the image
	Path string // Path of the image (relative to Gallery)
}

// Gallery holds parameters related to a gallery
type Gallery struct {
	Directory string // name of the directory
	Title string // title of the gallery
	Cover string // path of the cover
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
	err := p.Template.Execute(w, HomePageParams{PageParams{"home", p.Config.Title}, p.Carousel})
	if err != nil {
		log.Printf("error while serving template home: %s", err)
	}
}

// Galleries is a DynamicHandlerFunc that serves the galleries page
func GalleriesHandler(p *DynamicHandlerFuncParams, w http.ResponseWriter, r *http.Request) {
	var galleries []Gallery
	file, err := ioutil.ReadFile("www/static/galleries/galleries.json")
	if err != nil {
		log.Printf("galleries error: %v", err)
	} else {
		err = json.Unmarshal(file, &galleries)
		if err != nil {
			log.Printf("galleries error: %v", err)
		}
	}

	err = p.Template.Execute(w, GalleriesPageParams{PageParams{"galleries", p.Config.Title}, galleries, p.Carousel})
	if err != nil {
		log.Printf("error while serving template galleries: %s", err)
	}
}

// ImageHandler is a DynamicHandlerFunc that serves the image page
func ImageHandler(p *DynamicHandlerFuncParams, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	gallery := vars["gallery"]
	image := vars["image"]

	err := p.Template.Execute(w, ImagePageParams{PageParams{"image", p.Config.Title}, p.Carousel, gallery, image})
	if err != nil {
		log.Printf("error while serving template image: %s", err)
	}
}

// GalleryHandler is a DynamicHandlerFunc that serves the gallery page
func GalleryHandler(p *DynamicHandlerFuncParams, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gallery := vars["gallery"]
	dir := fmt.Sprintf("www/static/galleries/%s", gallery);
	var images []string
	filepath.Walk(dir, func (path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			name := filepath.Base(path)
			images = append(images, name)
		}
		return nil
	})
	err := p.Template.Execute(w, GalleryPageParams{PageParams{"gallery", p.Config.Title}, p.Carousel, gallery, images})
	if err != nil {
		log.Printf("error while serving template gallery: %s", err)
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
		handler(&DynamicHandlerFuncParams{b.Router, tpl, b.Config, b.Carousel}, w, r)
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

	// All pages have access to the carousel but they can discard it
	var carousel_items []map[string]string
	file, err = ioutil.ReadFile("www/static/carousel/carousel.json")
	if err != nil {
		log.Printf("carousel error: %v", err)
	} else {
		err = json.Unmarshal(file, &carousel_items)
		if err != nil {
			log.Printf("carousel error: %v", err)
		}
	}

	result := &Bean{
		&cfg,
		r,
		carousel_items,
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
	bean.Router.HandleFunc("/galleries.html", bean.BuildDynamicdHandler(GalleriesHandler, "galleries"))
	bean.Router.HandleFunc("/gallery/{gallery}.html", bean.BuildDynamicdHandler(GalleryHandler, "gallery"))
	bean.Router.HandleFunc("/image/{gallery}/{image}.html", bean.BuildDynamicdHandler(ImageHandler, "image"))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("www/static"))))
	http.Handle("/", bean.Router)

	log.Printf("bean started, serving pages")
	log.Fatal(http.ListenAndServe(bean.Config.ListenOn, nil))
}
