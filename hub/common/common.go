package common

import (
	"appengine"
	"appengine/user"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var prefPattern = regexp.MustCompile("^([^\\s;]+)(;q=([\\d.]+))?$")

type Context struct {
	appengine.Context
	Req     *http.Request
	Resp    http.ResponseWriter
	Version string
	User    *user.User
}

func (self Context) RenderJSON(i interface{}) {
	self.Resp.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(self.Resp).Encode(i); err != nil {
		panic(err)
	}
}

func MustParseFloat64(s string) (result float64) {
	var err error
	if result, err = strconv.ParseFloat(s, 64); err != nil {
		panic(err)
	}
	return
}

func MostAccepted(r *http.Request, def, name string) string {
	bestValue := def
	var bestScore float64 = -1
	var score float64
	for _, pref := range strings.Split(r.Header.Get(name), ",") {
		if match := prefPattern.FindStringSubmatch(pref); match != nil {
			score = 1
			if match[3] != "" {
				score = MustParseFloat64(match[3])
			}
			if score > bestScore {
				bestScore = score
				bestValue = match[1]
			}
		}
	}
	return bestValue
}
