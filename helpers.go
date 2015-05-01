/* This Source Code Form is subject to the terms of the Mozilla Public
* License, v. 2.0. If a copy of the MPL was not distributed with this
* file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"reflect"
	"strings"
	"testing"
)

func Fail(t *testing.T, expected interface{}) {
	t.Errorf("Expected [%v] (%v)", expected, reflect.TypeOf(expected))
}

func Expect(t *testing.T, got interface{}, expected interface{}) {
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Expected [%v], but got [%v]", expected, got)
	}
}

func NotExpect(t *testing.T, got interface{}, expected interface{}) {
	if reflect.DeepEqual(expected, got) {
		t.Errorf("Expected [%v] not to be [%v]", expected, got)
	}
}

func Contain(t *testing.T, body string, expected string) {
	if !strings.Contains(body, expected) {
		t.Errorf("Expected body to contain [%v]", expected)
	}
}

func NotContain(t *testing.T, body string, expected string) {
	if strings.Contains(body, expected) {
		t.Errorf("Expected body not to contain [%v]", expected)
	}
}

func Contains(t *testing.T, got []interface{}, expected interface{}) {
	for i := range got {
		if reflect.DeepEqual(got[i], expected) {
			return
		}
	}
	t.Errorf("Expected slice to contain [%v]", expected)
}

func Trim(input string) string {
	return strings.Trim(input, "\t\n\f\r ")
}
