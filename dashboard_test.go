/* This Source Code Form is subject to the terms of the Mozilla Public
* License, v. 2.0. If a copy of the MPL was not distributed with this
* file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-martini/martini"
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
	Contain(t, body, `<div ng-view>`)
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
	Contain(t, body, `@media (max-width: 1016px) {`)
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
	r := martini.NewRouter()
	m.Action(r.Handle)

	hostname := currentHostname
	defer func() {
		currentHostname = hostname
	}()
	currentHostname = "will_cause_error"
	r.Get("/api/ip", DataHandler("ip"))

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost:4005/api/ip", nil)
	if err != nil {
		t.Fatal(err)
	}

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

	host, err := hostname()
	if err != nil {
		t.Fatal(err)
	}
	if host.Hostname == "" {
		t.Fatal("could not figure out hostname")
	}
	body := response.Body.String()
	Contain(t, body, `"Hostname": "`+host.Hostname)

	var data Host
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		t.Fatal(err)
	}
	Expect(t, data, Host{host.Hostname})
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

	host, err := hostname()
	if err != nil {
		t.Fatal(err)
	}
	ips, err := ip(host.Hostname)
	if err != nil {
		t.Fatal(err)
	}
	if len(ips) == 0 {
		t.Fatal("no IPs found")
	}
	body := response.Body.String()
	Contain(t, body, `"`+ips[0]+`"`)

	var data []string
	if err := json.Unmarshal([]byte(body), &data); err != nil {
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

	cpuData, err := cpu()
	if err != nil {
		t.Fatal(err)
	}
	body := response.Body.String()
	Contain(t, body, fmt.Sprintf(`"Processors": %d`, cpuData.Processors))
	Contain(t, body, fmt.Sprintf(`"ModelName": "%s"`, cpuData.ModelName))
	Contain(t, body, `"Speed": `)
	Contain(t, body, `"Load1": `)
	Contain(t, body, `"Load5": `)
	Contain(t, body, `"Load15": `)
	Contain(t, body, `"Processes": `)
	Expect(t, cpuData.Load1 > 0, true)
	Contain(t, cpuData.Processes, `/`)
	Expect(t, cpuData.Processors >= 1, true)
	Expect(t, cpuData.Speed > 1000, true)

	var data *CPU
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		t.Fatal(err)
	}
	Expect(t, data.ModelName, cpuData.ModelName)
	Expect(t, data.Processors, cpuData.Processors)
	Expect(t, data.Speed, cpuData.Speed)
}

func Test_todoapp_api_GetMemory(t *testing.T) {
	m := setupMartini()

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost:4005/api/mem", nil)
	if err != nil {
		t.Fatal(err)
	}

	m.ServeHTTP(response, req)
	Expect(t, response.Code, http.StatusOK)

	memory, err := mem()
	if err != nil {
		t.Fatal(err)
	}
	body := response.Body.String()
	Contain(t, body, `"RAM": {`)
	Contain(t, body, `"Swap": {`)
	Contain(t, body, `"Total": {`)
	Contain(t, body, `"TotalM": `)
	Contain(t, body, `"FreeH": `)
	Contain(t, body, `"UsedM": `)
	NotExpect(t, memory.RAM.TotalM, 0)
	NotExpect(t, memory.RAM.FreeM, 0)
	NotExpect(t, memory.Total.TotalM, 0)
	NotExpect(t, memory.Total.FreeM, 0)

	var data *Memory
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		t.Fatal(err)
	}
	Expect(t, data.Total.TotalM, memory.Total.TotalM)
	Expect(t, data.Swap, memory.Swap)
	Expect(t, data.RAM.TotalH, memory.RAM.TotalH)
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
	body := response.Body.String()
	Contain(t, body, `"MountedOn": "/"`)
	Contain(t, body, `"Filesystem": "tmpfs",`)
	NotExpect(t, len(disks), 1)

	var data []*DiskUsage
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		t.Fatal(err)
	}
	Expect(t, data, disks)
}

func Test_todoapp_api_GetProcesses(t *testing.T) {
	m := setupMartini()

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost:4005/api/processes", nil)
	if err != nil {
		t.Fatal(err)
	}

	m.ServeHTTP(response, req)
	Expect(t, response.Code, http.StatusOK)

	processes, err := top()
	if err != nil {
		t.Fatal(err)
	}
	body := response.Body.String()
	Contain(t, body, `"Header": [`)
	Contain(t, body, `"Tasks: `)
	Contain(t, body, `"Processes": [`)
	Contain(t, body, `"User": "`)
	Contain(t, body, `"Command": "`)
	Contain(t, body, `"Tty": "`)
	Expect(t, len(processes.Processes) > 5, true)

	var data *Top
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		t.Fatal(err)
	}
	Expect(t, len(data.Processes), len(processes.Processes))
}

func Test_todoapp_api_GetLoggedOn(t *testing.T) {
	m := setupMartini()

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost:4005/api/logged_on", nil)
	if err != nil {
		t.Fatal(err)
	}

	m.ServeHTTP(response, req)
	Expect(t, response.Code, http.StatusOK)

	loggedOn, err := w()
	if err != nil {
		t.Fatal(err)
	}
	body := response.Body.String()
	Contain(t, body, `"TTY": `)
	Contain(t, body, `"Login": `)
	NotExpect(t, len(loggedOn), 0)

	var data []*LoggedOn
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		t.Fatal(err)
	}
	if len(data) > 0 {
		Expect(t, data[0].User, loggedOn[0].User)
		Expect(t, data[0].From, loggedOn[0].From)
		Expect(t, data[0].TTY, loggedOn[0].TTY)
	}
}

func Test_todoapp_api_GetUsers(t *testing.T) {
	m := setupMartini()

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost:4005/api/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	m.ServeHTTP(response, req)
	Expect(t, response.Code, http.StatusOK)

	users, err := passwd()
	if err != nil {
		t.Fatal(err)
	}
	body := response.Body.String()
	Contain(t, body, `"Type": "system",`)
	Contain(t, body, `"Type": "user",`)
	Contain(t, body, `"Name": "root"`)
	Contain(t, body, `"Description": `)
	Contain(t, body, `"Home": "/root"`)
	Contain(t, body, `"Shell": "/bin/bash"`)
	Contain(t, body, `"Shell": "/bin/false"`)
	Expect(t, len(users) >= 5, true)

	var data []*User
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		t.Fatal(err)
	}
	Expect(t, data, users)
}

func Test_todoapp_api_GetNetwork(t *testing.T) {
	m := setupMartini()

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost:4005/api/network", nil)
	if err != nil {
		t.Fatal(err)
	}

	m.ServeHTTP(response, req)
	Expect(t, response.Code, http.StatusOK)

	network, err := network()
	if err != nil {
		t.Fatal(err)
	}
	body := response.Body.String()
	Contain(t, body, `"Name": "lo",`)
	Contain(t, body, `"Type": "inet",`)
	Contain(t, body, `"Value": "`)
	Expect(t, len(network) >= 1, true)

	var data []*If
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		t.Fatal(err)
	}
	Expect(t, data, network)
}
