package common

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"fmt"
	"github.com/zond/stockholm-ai/common"
	"net/http"
	"regexp"
	"strings"
)

var prefPattern = regexp.MustCompile("^([^\\s;]+)(;q=([\\d.]+))?$")

func Transaction(outerContext Context, f func(Context) error) error {
	return datastore.RunInTransaction(outerContext, func(innerContext appengine.Context) error {
		cpy := outerContext
		cpy.Context = innerContext
		return f(cpy)
	}, &datastore.TransactionOptions{XG: true})
}

type GAELogger struct {
	appengine.Context
}

func (self GAELogger) Printf(f string, i ...interface{}) {
	self.Infof(f, i...)
}

func GAELoggerFactory(r *http.Request) common.Logger {
	return GAELogger{appengine.NewContext(r)}
}

func MustMarshal(i interface{}) (b []byte) {
	var err error
	if b, err = MemCodec.Marshal(i); err != nil {
		panic(err)
	}
	return
}

func MustUnmarshal(b []byte, i interface{}) {
	if err := MemCodec.Unmarshal(b, i); err != nil {
		panic(err)
	}
}

type Context struct {
	appengine.Context
	Req     *http.Request
	Resp    http.ResponseWriter
	Version string
	User    *user.User
	Vars    map[string]string
}

func (self Context) Authenticated() bool {
	if self.User != nil {
		return true
	}
	fmt.Fprintln(self.Resp, "Unauthorized")
	self.Resp.WriteHeader(401)
	return false
}

func SetContentType(w http.ResponseWriter, t string, cache bool) {
	w.Header().Set("Content-Type", t)
	w.Header().Set("Vary", "Accept")
	if cache {
		if !appengine.IsDevAppServer() {
			w.Header().Set("Cache-Control", "public, max-age=864000")
		}
	} else {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
	}
}

func (self Context) SetContentType(t string, cache bool) {
	SetContentType(self.Resp, t, cache)
}

func (self Context) RenderJSON(i interface{}) {
	self.SetContentType("application/json; charset=UTF-8", false)
	common.MustEncodeJSON(self.Resp, i)
}

func MustDecodeKey(s string) (result *datastore.Key) {
	var err error
	if result, err = datastore.DecodeKey(s); err != nil {
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
				score = common.MustParseFloat64(match[3])
			}
			if score > bestScore {
				bestScore = score
				bestValue = match[1]
			}
		}
	}
	return bestValue
}

func AssertOkError(err error) {
	if !IsOkError(err) {
		panic(err)
	}
}

func IsOkError(err error) bool {
	if err != nil {
		if merr, ok := err.(appengine.MultiError); ok {
			for _, serr := range merr {
				if serr != nil {
					if _, ok := serr.(*datastore.ErrFieldMismatch); !ok {
						return false
					}
				}
			}
		} else if _, ok := err.(*datastore.ErrFieldMismatch); !ok {
			return false
		}
	}
	return true
}
