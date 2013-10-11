package common

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type LoggerFactory func(r *http.Request) Logger

func MustEncodeJSON(w io.Writer, i interface{}) {
	if err := json.NewEncoder(w).Encode(i); err != nil {
		panic(err)
	}
}

func RandomString(n int) string {
	b := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		b = append(b, byte(rand.Int()))
	}
	return base64.URLEncoding.EncodeToString(b)
}

type Logger interface {
	Infof(f string, o ...interface{})
}

func MustDecodeJSON(r io.Reader, result interface{}) {
	if err := json.NewDecoder(r).Decode(result); err != nil {
		panic(err)
	}
}

func Prettify(obj interface{}) string {
	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
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

func MustParseFloat64(s string) (result float64) {
	var err error
	if result, err = strconv.ParseFloat(s, 64); err != nil {
		panic(err)
	}
	return
}

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

/*
Norm returns a random int in the normal distribution with average avg and standard deviance dev, strictly limited by min and max (inclusive).
*/
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
