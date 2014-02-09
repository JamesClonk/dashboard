/* This Source Code Form is subject to the terms of the Mozilla Public
* License, v. 2.0. If a copy of the MPL was not distributed with this
* file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func init() {
	port := "4005"
	os.Setenv("PORT", port)
}

func Test_todoapp_index(t *testing.T) {
	m := setupMartini()

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost:4005/", nil)
	if err != nil {
		t.Fatal(err)
	}

	m.ServeHTTP(response, req)
	Expect(t, response.Code, http.StatusOK)

	body := response.Body.String()
	Contain(t, body, `<html lang="en" ng-app="dashboard" ng-controller="dashboardCtrl">`)
	Contain(t, body, `<title>Dashboard - {{Hostname}}</title>`)
	Contain(t, body, `<div ng-view></div>`)
}

func Test_todoapp_assets(t *testing.T) {
	m := setupMartini()

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost:4005/js/dashboard.js", nil)
	if err != nil {
		t.Fatal(err)
	}

	m.ServeHTTP(response, req)
	Expect(t, response.Code, http.StatusOK)

	body := response.Body.String()
	Contain(t, body, `var dashboard = angular.module('dashboard', [`)
	Contain(t, body, `dashboardControllers.controller('dashboardCtrl', ['$scope', '$http', '$location',`)

	response = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "http://localhost:4005/css/dashboard.css", nil)
	if err != nil {
		t.Fatal(err)
	}

	m.ServeHTTP(response, req)
	Expect(t, response.Code, http.StatusOK)

	body = response.Body.String()
	Contain(t, body, `.fork-me {`)
	Contain(t, body, `@media (max-width: 767px) {`)
}

func Test_todoapp_404(t *testing.T) {
	m := setupMartini()

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost:4005/unknown", nil)
	if err != nil {
		t.Fatal(err)
	}

	m.ServeHTTP(response, req)
	Expect(t, response.Code, http.StatusNotFound)

	body := response.Body.String()
	Contain(t, body, `<h1>404 - Not Found</h1>`)
	Contain(t, body, `<h4>This is not the page you are looking for..</h4>`)
}

func Test_todoapp_500(t *testing.T) {
	m := setupMartini()

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost:4005/api/ip", nil)
	if err != nil {
		t.Fatal(err)
	}

	hostname := currentHostname
	defer func() {
		currentHostname = hostname
	}()
	currentHostname = "will_cause_error"
	m.ServeHTTP(response, req)
	Expect(t, response.Code, http.StatusInternalServerError)

	body := response.Body.String()
	Contain(t, body, `<h1>500 - Internal Server Error</h1>`)
	Contain(t, body, `<h4>lookup will_cause_error: no such host</h4>`)
}

func Test_todoapp_api_GetHostname(t *testing.T) {
	m := setupMartini()

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost:4005/api/hostname", nil)
	if err != nil {
		t.Fatal(err)
	}

	m.ServeHTTP(response, req)
	Expect(t, response.Code, http.StatusOK)

	hostname, err := hostname()
	if err != nil {
		t.Fatal(err)
	}
	body := response.Body.String()
	Contain(t, body, `"Hostname": "`+hostname)

	var data struct{ Hostname string }
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		t.Fatal(err)
	}
	Expect(t, data, struct{ Hostname string }{hostname})
}

func Test_todoapp_api_GetIP(t *testing.T) {
	m := setupMartini()

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost:4005/api/ip", nil)
	if err != nil {
		t.Fatal(err)
	}

	m.ServeHTTP(response, req)
	Expect(t, response.Code, http.StatusOK)

	hostname, err := hostname()
	if err != nil {
		t.Fatal(err)
	}
	ips, err := ip(hostname)
	if err != nil {
		t.Fatal(err)
	}
	var data []string
	if err := json.Unmarshal([]byte(response.Body.String()), &data); err != nil {
		t.Fatal(err)
	}
	Expect(t, data, ips)
}

func Test_todoapp_api_GetCPU(t *testing.T) {
	m := setupMartini()

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost:4005/api/cpu", nil)
	if err != nil {
		t.Fatal(err)
	}

	m.ServeHTTP(response, req)
	Expect(t, response.Code, http.StatusOK)

	t.Fail()
}

func Test_todoapp_api_GetDisk(t *testing.T) {
	m := setupMartini()

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost:4005/api/disk", nil)
	if err != nil {
		t.Fatal(err)
	}

	m.ServeHTTP(response, req)
	Expect(t, response.Code, http.StatusOK)

	disks, err := df()
	if err != nil {
		t.Fatal(err)
	}
	var data []*DiskUsage
	if err := json.Unmarshal([]byte(response.Body.String()), &data); err != nil {
		t.Fatal(err)
	}
	Expect(t, data, disks)
}
