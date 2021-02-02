package common

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"time"

	"github.com/golang/snappy"
	"golang.org/x/net/context"

	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

const (
	regularCache = iota
	nilCache
)

var snappyGobCodec = memcache.Codec{
	Marshal: func(i interface{}) (compressed []byte, err error) {
		uncompressed, err := memcache.Gob.Marshal(i)
		if err != nil {
			return
		}
		compressed = snappy.Encode(nil, uncompressed)
		return
	},
	Unmarshal: func(compressed []byte, i interface{}) (err error) {
		var uncompressed []byte
		if uncompressed, err = snappy.Decode(nil, compressed); err != nil {
			return
		}
		err = memcache.Gob.Unmarshal(uncompressed, i)
		return
	},
}

var MemCodec = snappyGobCodec

func isNil(v reflect.Value) bool {
	k := v.Kind()
	if k == reflect.Chan {
		return v.IsNil()
	}
	if k == reflect.Func {
		return v.IsNil()
	}
	if k == reflect.Interface {
		return v.IsNil()
	}
	if k == reflect.Map {
		return v.IsNil()
	}
	if k == reflect.Ptr {
		return v.IsNil()
	}
	if k == reflect.Slice {
		return v.IsNil()
	}
	return false
}

func keyify(k string) string {
	buf := new(bytes.Buffer)
	enc := base64.NewEncoder(base64.StdEncoding, buf)
	h := sha1.New()
	io.WriteString(h, k)
	sum := h.Sum(nil)
	if wrote, err := enc.Write(sum); err != nil {
		panic(err)
	} else if wrote != len(sum) {
		panic(fmt.Errorf("Tried to write %v bytes but wrote %v bytes", len(sum), wrote))
	}
	if err := enc.Close(); err != nil {
		panic(err)
	}
	return string(buf.Bytes())
}

func MemDel(c context.Context, keys ...string) {
	for index, key := range keys {
		keys[index] = keyify(key)
	}
	memcache.DeleteMulti(c, keys)
}

func MemGet(c context.Context, key string, val interface{}) {
	if _, err := MemCodec.Get(c, keyify(key), val); err != nil && err != memcache.ErrCacheMiss {
		log.Errorf(c, "When trying to load %v(%v) => %#v: %#v", key, keyify(key), val, err)
		panic(err)
	}
}

func MemPut(c context.Context, key string, val interface{}) {
	if err := MemCodec.Set(c, &memcache.Item{
		Key:    keyify(key),
		Object: val,
	}); err != nil {
		log.Errorf(c, "When trying to save %v(%v) => %#v: %#v", key, keyify(key), val, err)
		panic(err)
	}
}

func Memoize2(c context.Context, super, key string, destP interface{}, f func() interface{}) (exists bool) {
	superH := keyify(super)
	var seed string
	item, err := memcache.Get(c, superH)
	if err != nil && err != memcache.ErrCacheMiss {
		panic(err)
	}
	if err == memcache.ErrCacheMiss {
		seed = fmt.Sprint(rand.Int63())
		item = &memcache.Item{
			Key:   superH,
			Value: []byte(seed),
		}
		//log.Infof(c, "Didn't find %v in memcache, reseeding with %v", super, seed)
		if err = memcache.Set(c, item); err != nil {
			panic(err)
		}
	} else {
		seed = string(item.Value)
	}
	return Memoize(c, fmt.Sprintf("%v@%v", key, seed), destP, f)
}

func reflectCopy(srcValue reflect.Value, source, destP interface{}) {
	if reflect.PtrTo(reflect.TypeOf(source)) == reflect.TypeOf(destP) {
		reflect.ValueOf(destP).Elem().Set(srcValue)
	} else {
		reflect.ValueOf(destP).Elem().Set(reflect.Indirect(srcValue))
	}
}

func MemoizeDuring(c context.Context, key string, dur time.Duration, cacheNil bool, destP interface{}, f func() interface{}) (exists bool) {
	return memoizeMulti(c, []string{key}, dur, cacheNil, []interface{}{destP}, []func() interface{}{f})[0]
}

func Memoize(c context.Context, key string, destP interface{}, f func() interface{}) (exists bool) {
	return MemoizeMulti(c, []string{key}, []interface{}{destP}, []func() interface{}{f})[0]
}

func memGetMulti(c context.Context, keys []string, dests []interface{}) (items []*memcache.Item, errors []error) {
	items = make([]*memcache.Item, len(keys))
	errors = make([]error, len(keys))

	itemHash, err := memcache.GetMulti(c, keys)
	if err != nil {
		for index, _ := range errors {
			errors[index] = err
		}
		return
	}

	var item *memcache.Item
	var ok bool
	for index, keyHash := range keys {
		if item, ok = itemHash[keyHash]; ok {
			items[index] = item
			if err := MemCodec.Unmarshal(item.Value, dests[index]); err != nil {
				errors[index] = err
			}
		} else {
			errors[index] = memcache.ErrCacheMiss
		}
	}
	return
}

func MemoizeMulti(c context.Context, keys []string, destPs []interface{}, f []func() interface{}) (exists []bool) {
	return memoizeMulti(c, keys, 0, true, destPs, f)
}

func memoizeMulti(c context.Context, keys []string, dur time.Duration, cacheNil bool, destPs []interface{}, f []func() interface{}) (exists []bool) {
	exists = make([]bool, len(keys))
	keyHashes := make([]string, len(keys))
	for index, key := range keys {
		keyHashes[index] = keyify(key)
	}

	t := time.Now()
	items, errors := memGetMulti(c, keyHashes, destPs)
	if d := time.Now().Sub(t); d > time.Millisecond*10 {
		log.Debugf(c, "SLOW memGetMulti(%v): %v", keys, d)
	}

	done := make(chan bool, len(items))

	for i, it := range items {
		index := i
		item := it
		err := errors[index]
		keyH := keyHashes[index]
		key := keys[index]
		destP := destPs[index]
		if err == memcache.ErrCacheMiss {
			log.Infof(c, "Didn't find %v in memcache, regenerating", key)
			go func() {
				defer func() {
					done <- true
				}()
				result := f[index]()
				resultValue := reflect.ValueOf(result)
				if isNil(resultValue) {
					if cacheNil {
						nilObj := reflect.Indirect(reflect.ValueOf(destP)).Interface()
						if err = MemCodec.Set(c, &memcache.Item{
							Key:        keyH,
							Flags:      nilCache,
							Object:     nilObj,
							Expiration: dur,
						}); err != nil {
							log.Errorf(c, "When trying to save %v(%v) => %#v: %#v", key, keyH, nilObj, err)
							panic(err)
						}
					}
					exists[index] = false
				} else {
					if err = MemCodec.Set(c, &memcache.Item{
						Key:        keyH,
						Object:     result,
						Expiration: dur,
					}); err != nil {
						log.Errorf(c, "When trying to save %v(%v) => %#v: %#v", key, keyH, result, err)
						panic(err)
					} else {
						reflectCopy(resultValue, result, destP)
						exists[index] = true
					}
				}
			}()
		} else if err != nil {
			log.Errorf(c, "When trying to get %v(%v): %#v", key, keyH, err)
			panic(err)
		} else {
			if item.Flags&nilCache == nilCache {
				exists[index] = false
			} else {
				exists[index] = true
			}
			done <- true
		}
	}
	for i := 0; i < len(items); i++ {
		<-done
	}
	return
}
