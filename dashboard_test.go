/* This Source Code Form is subject to the terms of the Mozilla Public
* License, v. 2.0. If a copy of the MPL was not distributed with this
* file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	//"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func init() {
	// Port during tests
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
	Contain(t, body, `<title>Dashboard - `+getCurrentHostname()+`</title>`)
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
	Contain(t, body, `dashboard.config(['$routeProvider',`)

	response = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "http://localhost:4005/css/dashboard.css", nil)
	if err != nil {
		t.Fatal(err)
	}

	m.ServeHTTP(response, req)
	Expect(t, response.Code, http.StatusOK)

	body = response.Body.String()
	Contain(t, body, `.completed, .completed a {`)
	Contain(t, body, `color: #333333;`)
	Contain(t, body, `background-color: #999999;`)
	Contain(t, body, `text-decoration: line-through;`)
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
	Contain(t, body, `<h5>This is not the page you are looking for..</h5>`)
}

func Test_todoapp_500(t *testing.T) {
	m := setupMartini()

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost:4005/", nil)
	if err != nil {
		t.Fatal(err)
	}

	m.ServeHTTP(response, req)
	Expect(t, response.Code, http.StatusInternalServerError)

	body := response.Body.String()
	Contain(t, body, `<h1>500 - Internal Server Error</h1>`)
	Contain(t, body, `<h5>...</h5>`)
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
	//body := response.Body.String()
	//Contain(t, body, `"abc": 123`)

	// var jsonData Struct
	// if err := json.Unmarshal([]byte(body), &jsonData); err != nil {
	// 	t.Fatal(err)
	// }
	// Expect(t, jsonData, expectedJsonData)
}
