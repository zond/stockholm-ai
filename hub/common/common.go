package common

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var prefPattern = regexp.MustCompile("^([^\\s;]+)(;q=([\\d.]+))?$")

func Max(is ...int) (result int) {
	result = is[0]
	for _, i := range is[1:] {
		if i > result {
			result = i
		}
	}
	return
}

func Min(is ...int) (result int) {
	result = is[0]
	for _, i := range is[1:] {
		if i < result {
			result = i
		}
	}
	return
}

func Norm(avg, dev, min, max int) (result int) {
	result = int(rand.NormFloat64()*float64(dev) + float64(avg))
	if result < min {
		result = min
	}
	if result > max {
		result = max
	}
	return
}

func Transaction(outerContext Context, f func(Context) error) error {
	return datastore.RunInTransaction(outerContext, func(innerContext appengine.Context) error {
		cpy := outerContext
		cpy.Context = innerContext
		return f(cpy)
	}, &datastore.TransactionOptions{XG: true})
}

func Prettify(obj interface{}) string {
	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

func RandomString(n int) string {
	b := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		b = append(b, byte(rand.Int()))
	}
	return base64.URLEncoding.EncodeToString(b)
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

func (self Context) RenderJSON(i interface{}) {
	self.Resp.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(self.Resp).Encode(i); err != nil {
		panic(err)
	}
}

func MustDecodeJSON(r io.Reader, result interface{}) {
	if err := json.NewDecoder(r).Decode(result); err != nil {
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

func MustDecodeKey(s string) (result *datastore.Key) {
	var err error
	if result, err = datastore.DecodeKey(s); err != nil {
		panic(err)
	}
	return
}

func MustMarshalJSON(i interface{}) (b []byte) {
	var err error
	if b, err = json.Marshal(i); err != nil {
		panic(err)
	}
	return
}

func MustUnmarshalJSON(b []byte, i interface{}) {
	var err error
	if err = json.Unmarshal(b, i); err != nil {
		panic(err)
	}
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
