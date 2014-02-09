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
	"strconv"
	"strings"
)

func hostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return hostname, nil
}

func ip(hostname string) ([]string, error) {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, err
	}

	if len(ips) > 0 {
		result := make([]string, len(ips))
		for _, ip := range ips {
			result = append(result, ip.String())
		}
		return result, nil
	}
	return nil, errors.New("could not figure out IP address")
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

	out, err := pipes(
		exec.Command("df", "-h"),
		exec.Command("awk", `{print $1";"$2";"$3";"$4";"$5";"$6}`),
	)

	lines := strings.Split(strings.Trim(out, " \t\n"), "\n")
	for _, line := range lines[1:] {
		values := strings.Split(line, ";")

		percentage, err := strconv.Atoi(strings.Trim(values[4], "% \t\n"))
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
