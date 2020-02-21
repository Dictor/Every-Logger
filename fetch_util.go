package main

import (
	"strconv"
	"strings"
)

func FetchStringStdCb(val string) (float64, bool) {
	ival, err := strconv.Atoi(strings.Replace(val, ",", "", 1))
	if err != nil {
		return 0.0, false
	} else {
		return float64(ival), true
	}
}
