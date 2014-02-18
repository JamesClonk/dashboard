/* This Source Code Form is subject to the terms of the Mozilla Public
* License, v. 2.0. If a copy of the MPL was not distributed with this
* file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var (
	rxW = regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(.*)$`)
)

type Host struct {
	Hostname string
}

func hostname() (*Host, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return &Host{hostname}, nil
}

func ip(hostname string) (result []string, err error) {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, err
	}

	if len(ips) > 0 {
		for _, ip := range ips {
			if ip.String() != "" {
				result = append(result, ip.String())
			}
		}
		return result, nil
	}
	return nil, errors.New("could not figure out IP address")
}

type CPU struct {
	Processors int
	ModelName  string
	Speed      float64
	Load1      float64
	Load5      float64
	Load15     float64
	Processes  string
}

func cpu() (result *CPU, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()
	result = &CPU{}

	// cat /proc/cpuinfo | grep -c '^processor'
	out, err := pipes(
		exec.Command("cat", "/proc/cpuinfo"),
		exec.Command("grep", "-c", "^processor"),
	)
	processors, err := strconv.Atoi(Trim(out))
	if err != nil {
		return nil, err
	}
	result.Processors = processors

	// cat /proc/cpuinfo | grep '^model name' | head -n 1 | awk -F":" '{print $2;}'
	out, err = pipes(
		exec.Command("cat", "/proc/cpuinfo"),
		exec.Command("grep", "^model name"),
		exec.Command("head", "-n", "1"),
		exec.Command("awk", `-F:`, "{print $2;}"),
	)
	result.ModelName = Trim(out)

	// cat /proc/cpuinfo | grep '^cpu MHz' | head -n 1 | awk -F":" '{print $2;}'
	out, err = pipes(
		exec.Command("cat", "/proc/cpuinfo"),
		exec.Command("grep", "^cpu MHz"),
		exec.Command("head", "-n", "1"),
		exec.Command("awk", `-F:`, "{print $2;}"),
	)
	speed, err := strconv.ParseFloat(Trim(out), 64)
	if err != nil {
		return nil, err
	}
	result.Speed = speed

	// cat /proc/loadavg | awk '{print $1";"$2";"$3";"$4;}'
	out, err = pipes(
		exec.Command("cat", "/proc/loadavg"),
		exec.Command("awk", `{print $1";"$2";"$3";"$4;}`),
	)
	fields := strings.SplitN(Trim(out), ";", 43)
	var loads []float64
	for i := 0; i < 3; i++ {
		number, err := strconv.ParseFloat(Trim(fields[i]), 64)
		if err != nil {
			return nil, err
		}
		loads = append(loads, number)
	}
	result.Load1 = loads[0]
	result.Load5 = loads[0]
	result.Load15 = loads[0]
	result.Processes = fields[3]

	return result, err
}

type MemoryData struct {
	TotalM int
	TotalH string
	UsedM  int
	UsedH  string
	FreeM  int
	FreeH  string
}

type Memory struct {
	RAM   MemoryData
	Swap  MemoryData
	Total MemoryData
}

func mem() (memory *Memory, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()
	memory = &Memory{}

	// free -otm | awk '{print $1";"$2";"$3";"$4;}'
	out, err := pipes(
		exec.Command("free", "-otm"),
		exec.Command("awk", `{print $1";"$2";"$3";"$4;}`),
	)
	lines := strings.Split(Trim(out), "\n")
	for _, line := range lines[1:] {
		values := strings.SplitN(line, ";", 4)

		total, err := strconv.Atoi(Trim(values[1]))
		if err != nil {
			return nil, err
		}
		used, err := strconv.Atoi(Trim(values[2]))
		if err != nil {
			return nil, err
		}
		free, err := strconv.Atoi(Trim(values[3]))
		if err != nil {
			return nil, err
		}

		data := MemoryData{
			TotalM: total,
			UsedM:  used,
			FreeM:  free,
		}

		switch Trim(values[0]) {
		case "Mem:":
			memory.RAM = data
		case "Swap:":
			memory.Swap = data
		case "Total:":
			memory.Total = data
		}
	}

	// free -oth | awk '{print $1";"$2";"$3";"$4;}'
	out, err = pipes(
		exec.Command("free", "-oth"),
		exec.Command("awk", `{print $1";"$2";"$3";"$4;}`),
	)
	lines = strings.Split(Trim(out), "\n")
	for _, line := range lines[1:] {
		values := strings.SplitN(line, ";", 4)

		var data *MemoryData
		switch Trim(values[0]) {
		case "Mem:":
			data = &memory.RAM
		case "Swap:":
			data = &memory.Swap
		case "Total:":
			data = &memory.Total
		}
		data.TotalH = Trim(values[1])
		data.UsedH = Trim(values[2])
		data.FreeH = Trim(values[3])
	}

	return memory, err
}

type DiskUsage struct {
	Filesystem      string
	Size            string
	Used            string
	Available       string
	UsagePercentage int
	MountedOn       string
}

func df() (diskUsage []*DiskUsage, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()

	// df -h | awk '{print $1";"$2";"$3";"$4";"$5";"$6;}'
	out, err := pipes(
		exec.Command("df", "-h"),
		exec.Command("awk", `{print $1";"$2";"$3";"$4";"$5";"$6;}`),
	)

	lines := strings.Split(Trim(out), "\n")
	for _, line := range lines[1:] {
		values := strings.SplitN(line, ";", 6)

		percentage, err := strconv.Atoi(strings.Trim(Trim(values[4]), "%"))
		if err != nil {
			return nil, err
		}

		diskUsage = append(diskUsage,
			&DiskUsage{
				Filesystem:      values[0],
				Size:            values[1],
				Used:            values[2],
				Available:       values[3],
				UsagePercentage: percentage,
				MountedOn:       values[5],
			})
	}

	return diskUsage, err
}

type LoggedOn struct {
	User  string
	TTY   string
	From  string
	Login string
	Idle  string
	JCPU  string
	PCPU  string
	What  string
}

func w() (loggedOn []*LoggedOn, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()

	// PROCPS_USERLEN=24 PROCPS_FROMLEN=64 w -ih | grep -v 'w -ih
	os.Setenv("PROCPS_USERLEN", "24")
	os.Setenv("PROCPS_FROMLEN", "64")
	out, err := pipes(
		exec.Command("w", "-ih"),
		exec.Command("grep", "-v", "w -ih"),
	)
	lines := strings.Split(Trim(out), "\n")
	for _, line := range lines {
		result := rxW.FindAllStringSubmatch(line, 8)
		if len(result) > 0 {
			loggedOn = append(loggedOn,
				&LoggedOn{
					User:  result[0][1],
					TTY:   result[0][2],
					From:  result[0][3],
					Login: result[0][4],
					Idle:  result[0][5],
					JCPU:  result[0][6],
					PCPU:  result[0][7],
					What:  Trim(result[0][8]),
				})
		}
	}

	return loggedOn, err
}

type User struct {
	Type        string
	Name        string
	Description string
	Home        string
	Shell       string
}

func passwd() (users []*User, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()

	// awk -F: '{ if ($3<=499) print "system;"$1";"$5";"$6";"$7; else print "user;"$1";"$5";"$6";"$7; }' /etc/passwd
	out, err := pipes(
		exec.Command("awk", "-F:", `{ if ($3<=499) print "system;"$1";"$5";"$6";"$7; else print "user;"$1";"$5";"$6";"$7; }`, "/etc/passwd"),
	)
	lines := strings.Split(Trim(out), "\n")
	for _, line := range lines {
		values := strings.SplitN(line, ";", 5)
		users = append(users,
			&User{
				Type:        values[0],
				Name:        values[1],
				Description: values[2],
				Home:        values[3],
				Shell:       values[4],
			})
	}

	return users, err
}

func pipes(commands ...*exec.Cmd) (string, error) {
	if len(commands) < 1 {
		return "", errors.New("not enough commands passed to pipes()")
	}
	var stdout bytes.Buffer

	for c := range commands[:len(commands)-1] {
		if pipe, err := commands[c].StdoutPipe(); err != nil {
			return "", err
		} else {
			commands[c+1].Stdin = pipe
		}
	}
	commands[len(commands)-1].Stdout = &stdout

	for _, command := range commands {
		if err := command.Start(); err != nil {
			return stdout.String(), err
		}
	}

	for _, command := range commands {
		if err := command.Wait(); err != nil {
			return stdout.String(), err
		}
	}

	return stdout.String(), nil
}
