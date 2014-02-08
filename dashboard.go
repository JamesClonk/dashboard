/* This Source Code Form is subject to the terms of the Mozilla Public
* License, v. 2.0. If a copy of the MPL was not distributed with this
* file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"log"
	"net"
	"net/http"
	"os"
)

var (
	logFlags           = log.Ldate | log.Ltime
	logPrefix          = "[dashboard] "
	productionHostname = "jamesclonk.io"
	currentHostname    = "localhost"
	currentIP          = "0.0.0.0"
)

type view struct {
	Title    string
	Hostname string
	IP       string
}

func init() {
	log.SetFlags(logFlags)
	log.SetPrefix(logPrefix)

	currentHostname = getCurrentHostname()
	currentIP = getIP(currentHostname)
}

func getCurrentHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Encountered a problem while trying to lookup current hostname: %v", err)
	}

	if hostname == productionHostname {
		log.Printf("Running on %s, switch to production settings\n", productionHostname)
		martini.Env = martini.Prod
	}

	return hostname
}

func getIP(hostname string) string {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		log.Fatalf("Encountered a problem while trying to lookup IP address of [%s]: %v", hostname, err)
	}

	if len(ips) > 0 {
		return ips[0].String()
	}

	return ""
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
	r.Get("/api/cpu", func(r render.Render) {
		r.JSON(http.StatusOK, nil)
	})
}

func View(title string) *view {
	return &view{
		Title:    title,
		Hostname: currentHostname,
		IP:       currentIP,
	}
}
