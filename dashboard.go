/* This Source Code Form is subject to the terms of the Mozilla Public
* License, v. 2.0. If a copy of the MPL was not distributed with this
* file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

var (
	logFlags           = log.Ldate | log.Ltime
	logPrefix          = "[dashboard] "
	productionHostname = "jamesclonk.io"
	currentHostname    = ""
)

type view struct {
	Title string
	Error error
	Data  interface{}
}

func init() {
	log.SetFlags(logFlags)
	log.SetPrefix(logPrefix)

	if hostname, err := hostname(); err != nil {
		log.Fatalf("Encountered a problem while trying to lookup current hostname: %v", err)
	} else {
		currentHostname = hostname.Hostname
	}

	if currentHostname == productionHostname {
		log.Printf("Running on %s, switch to production settings\n", productionHostname)
		martini.Env = martini.Prod
	}
}

func main() {
	m := setupMartini()
	m.Run()
}

func setupMartini() *martini.Martini {
	r := martini.NewRouter()
	m := martini.New()
	m.Use(martini.Recovery())
	m.Use(martini.Static("assets", martini.StaticOptions{SkipLogging: true})) // skip logging on static content
	m.Use(martini.Logger())
	m.Use(render.Renderer(render.Options{
		Directory:  "templates",
		Layout:     "layout",
		Extensions: []string{".html"},
		Delims:     render.Delims{"{[{", "}]}"},
		IndentJSON: true,
	}))
	m.Map(log.New(os.Stdout, logPrefix, logFlags))
	m.Action(r.Handle)

	setupRoutes(r)

	return m
}

func setupRoutes(r martini.Router) {
	// static
	r.Get("/", func(r render.Render) {
		r.HTML(http.StatusOK, "index", View("Dashboard"))
	})
	r.NotFound(func(r render.Render) {
		r.HTML(http.StatusNotFound, "404", View("404 - Not Found"))
	})

	// api
	r.Get("/api/hostname", DataHandler("hostname"))
	r.Get("/api/ip", DataHandler("ip"))
	r.Get("/api/cpu", DataHandler("cpu"))
	r.Get("/api/mem", DataHandler("mem"))
	r.Get("/api/disk", DataHandler("disk"))
	r.Get("/api/processes", DataHandler("top"))
	r.Get("/api/logged_on", DataHandler("logged_on"))
	r.Get("/api/users", DataHandler("passwd"))
	r.Get("/api/network", DataHandler("network"))

	r.Get("/api/debug/:method", DebugHandler)
}

func DebugHandler(params martini.Params, r render.Render) {
	method := params["method"]
	data, err := data(method)
	view := View("Debug")
	view.Error = err
	view.Data = data
	r.HTML(200, "debug", view)
}

func data(method string) (data interface{}, err error) {
	switch method {
	case "hostname":
		data, err = hostname()
	case "ip":
		data, err = ip(currentHostname)
	case "cpu":
		data, err = cpu()
	case "mem":
		data, err = mem()
	case "disk":
		data, err = df()
	case "top":
		data, err = top()
	case "logged_on":
		data, err = w()
	case "passwd":
		data, err = passwd()
	case "network":
		data, err = network()
	default:
		data, err = nil, errors.New("unknown method")
	}
	return
}

func DataHandler(method string) func(r render.Render) {
	return func(r render.Render) {
		data, err := data(method)
		if err != nil && method != "mem" { // exclude 'mem' for now.. there's some syntax problem on CF with it
			view := View("500 - Internal Server Error")
			view.Error = err
			r.HTML(http.StatusInternalServerError, "500", view)
			return
		}

		r.JSON(http.StatusOK, data)
	}
}

func View(title string) *view {
	return &view{
		Title: title,
	}
}
