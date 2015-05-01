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
	rxW  = regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(.*)$`)
	rxPs = regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(.*)$`)
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
	processors, err := strconv.Atoi(trim(out))
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
	result.ModelName = trim(out)

	// cat /proc/cpuinfo | grep '^cpu MHz' | head -n 1 | awk -F":" '{print $2;}'
	out, err = pipes(
		exec.Command("cat", "/proc/cpuinfo"),
		exec.Command("grep", "^cpu MHz"),
		exec.Command("head", "-n", "1"),
		exec.Command("awk", `-F:`, "{print $2;}"),
	)
	speed, err := strconv.ParseFloat(trim(out), 64)
	if err != nil {
		return nil, err
	}
	result.Speed = speed

	// cat /proc/loadavg | awk '{print $1";"$2";"$3";"$4;}'
	out, err = pipes(
		exec.Command("cat", "/proc/loadavg"),
		exec.Command("awk", `{print $1";"$2";"$3";"$4;}`),
	)
	fields := strings.SplitN(trim(out), ";", 43)
	var loads []float64
	for i := 0; i < 3; i++ {
		number, err := strconv.ParseFloat(trim(fields[i]), 64)
		if err != nil {
			return nil, err
		}
		loads = append(loads, number)
	}
	result.Load1 = loads[0]
	result.Load5 = loads[1]
	result.Load15 = loads[2]
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
	lines := strings.Split(trim(out), "\n")
	for _, line := range lines[1:] {
		values := strings.SplitN(line, ";", 4)

		total, err := strconv.Atoi(trim(values[1]))
		if err != nil {
			return nil, err
		}
		used, err := strconv.Atoi(trim(values[2]))
		if err != nil {
			return nil, err
		}
		free, err := strconv.Atoi(trim(values[3]))
		if err != nil {
			return nil, err
		}

		data := MemoryData{
			TotalM: total,
			UsedM:  used,
			FreeM:  free,
		}

		switch trim(values[0]) {
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
	lines = strings.Split(trim(out), "\n")
	for _, line := range lines[1:] {
		values := strings.SplitN(line, ";", 4)

		var data *MemoryData
		switch trim(values[0]) {
		case "Mem:":
			data = &memory.RAM
		case "Swap:":
			data = &memory.Swap
		case "Total:":
			data = &memory.Total
		}
		data.TotalH = trim(values[1])
		data.UsedH = trim(values[2])
		data.FreeH = trim(values[3])
	}

	fixMemory(memory)

	return memory, err
}

func fixMemory(mem *Memory) {
	fixMemoryData(&mem.RAM)
	fixMemoryData(&mem.Swap)
	fixMemoryData(&mem.Total)
}

func fixMemoryData(data *MemoryData) {
	if len(data.TotalH) == 0 {
		data.TotalH = fmt.Sprintf("%dM", data.TotalM)
	}
	if len(data.UsedH) == 0 {
		data.UsedH = fmt.Sprintf("%dM", data.UsedM)
	}
	if len(data.FreeH) == 0 {
		data.FreeH = fmt.Sprintf("%dM", data.FreeM)
	}
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

	// df -hP | awk '{print $1";"$2";"$3";"$4";"$5";"$6;}'
	out, err := pipes(
		exec.Command("df", "-hP"),
		exec.Command("awk", `{print $1";"$2";"$3";"$4";"$5";"$6;}`),
	)

	lines := strings.Split(trim(out), "\n")
	for _, line := range lines[1:] {
		if strings.HasPrefix(line, "Filesystem;") {
			continue
		}

		values := strings.SplitN(line, ";", 6)

		percentage, err := strconv.Atoi(strings.Trim(trim(values[4]), "%"))
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

type Top struct {
	Header    []string
	Processes []*Process
}

type Process struct {
	User    string
	Pid     float64
	Cpu     float64
	Mem     float64
	Vsz     float64
	Rss     float64
	Tty     string
	Stat    string
	Start   string
	Time    string
	Command string
}

func top() (data *Top, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()
	data = &Top{}

	// top -b -n 1 2>/dev/null | head -n 5
	out, err := pipes(
		exec.Command("top", "-b", "-n", "1"),
		exec.Command("head", "-n", "5"),
	)
	data.Header = strings.Split(trim(out), "\n")

	// ps -aux | tail -n +2 | grep -v 'ps -aux' | sort -nr -k6
	out, err = pipes(
		exec.Command("ps", "-aux"),
		exec.Command("tail", "-n", "+2"),
		exec.Command("grep", "-v", "ps -aux"),
		exec.Command("sort", "-nr", "-k6"),
	)
	lines := strings.Split(trim(out), "\n")
	for _, line := range lines {

		result := rxPs.FindAllStringSubmatch(line, 11)
		if len(result) > 0 && !(result[0][5] == "0" && result[0][6] == "0") {

			pid, err := strconv.ParseFloat(trim(result[0][2]), 64)
			if err != nil {
				return nil, err
			}
			cpu, err := strconv.ParseFloat(trim(result[0][3]), 64)
			if err != nil {
				return nil, err
			}
			mem, err := strconv.ParseFloat(trim(result[0][4]), 64)
			if err != nil {
				return nil, err
			}
			vsz, err := strconv.ParseFloat(trim(result[0][5]), 64)
			if err != nil {
				return nil, err
			}
			rss, err := strconv.ParseFloat(trim(result[0][6]), 64)
			if err != nil {
				return nil, err
			}

			data.Processes = append(data.Processes,
				&Process{
					User:    result[0][1],
					Pid:     pid,
					Cpu:     cpu,
					Mem:     mem,
					Vsz:     vsz,
					Rss:     rss,
					Tty:     result[0][7],
					Stat:    result[0][8],
					Start:   result[0][9],
					Time:    result[0][10],
					Command: trim(result[0][11]),
				})
		}
	}

	return data, err
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

	// PROCPS_USERLEN=24 PROCPS_FROMLEN=64 w -ih | grep -v 'w -ih'
	os.Setenv("PROCPS_USERLEN", "24")
	os.Setenv("PROCPS_FROMLEN", "64")
	out, err := pipes(
		exec.Command("w", "-ih"),
		exec.Command("grep", "-v", "w -ih"),
	)
	lines := strings.Split(trim(out), "\n")
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
					What:  trim(result[0][8]),
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
	lines := strings.Split(trim(out), "\n")
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

type If struct {
	Name  string
	Type  string
	Value string
}

func network() (network []*If, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()

	// ip -o addr | awk '{print $2";"$3";"$4;}'
	out, err := pipes(
		exec.Command("ip", "-o", "addr"),
		exec.Command("awk", `{print $2";"$3";"$4;}`),
	)
	lines := strings.Split(trim(out), "\n")
	for _, line := range lines {
		values := strings.SplitN(line, ";", 3)
		network = append(network,
			&If{
				Name:  values[0],
				Type:  values[1],
				Value: values[2],
			})
	}

	return network, err
}

type Env struct {
	Key   string
	Value string
}

func env() (env []*Env) {
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		env = append(env, &Env{pair[0], pair[1]})
	}
	return env
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

func trim(input string) string {
	return strings.Trim(input, "\t\n\f\r ")
}
