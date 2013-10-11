package models

import (
	"appengine/datastore"
	"common"
	"fmt"
	"sort"
	"time"
)

const (
	AIKind    = "AI"
	AllAIsKey = "AIs{All}"
)

func AIByIdKey(k interface{}) string {
	return fmt.Sprintf("AI{Id:%v}", k)
}

type AIs []AI

func (self AIs) Len() int {
	return len(self)
}

func (self AIs) Less(i, j int) bool {
	return self[j].CreatedAt.Before(self[i].CreatedAt)
}

func (self AIs) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self AIs) process(c common.Context) AIs {
	for index, _ := range self {
		(&self[index]).process(c)
	}
	return self
}

type AI struct {
	Id        *datastore.Key
	URL       string
	Name      string
	Owner     string `json:"-"`
	IsOwner   bool   `datastore:"-"`
	CreatedAt time.Time
}

func (self *AI) process(c common.Context) *AI {
	if c.User != nil {
		self.IsOwner = self.Owner == c.User.Email
	}
	if !self.IsOwner {
		self.URL = ""
	}
	return self
}

func findAIById(c common.Context, id *datastore.Key) *AI {
	var ai AI
	err := datastore.Get(c, id, &ai)
	if err == datastore.ErrNoSuchEntity {
		return nil
	}
	common.AssertOkError(err)
	ai.Id = id
	return &ai
}

func GetAIById(c common.Context, id *datastore.Key) *AI {
	var ai AI
	if common.Memoize(c, AIByIdKey(id), &ai, func() interface{} {
		return findAIById(c, id)
	}) {
		return &ai
	}
	return nil
}

func findAllAIs(c common.Context) (result AIs) {
	ids, err := datastore.NewQuery(AIKind).GetAll(c, &result)
	common.AssertOkError(err)
	for index, id := range ids {
		result[index].Id = id
	}
	if result == nil {
		result = AIs{}
	}
	return
}

func GetAllAIs(c common.Context) (result AIs) {
	common.Memoize(c, AllAIsKey, &result, func() interface{} {
		return findAllAIs(c)
	})
	sort.Sort(result)
	return result.process(c)
}

func (self *AI) Delete(c common.Context) {
	datastore.Delete(c, self.Id)
	common.MemDel(c, AllAIsKey, AIByIdKey(self.Id))
}

func (self *AI) Save(c common.Context) *AI {
	var err error
	if self.Id == nil {
		self.CreatedAt = time.Now()
		self.Id, err = datastore.Put(c, datastore.NewKey(c, AIKind, "", 0, nil), self)
	} else {
		_, err = datastore.Put(c, self.Id, self)
	}
	common.AssertOkError(err)
	common.MemDel(c, AllAIsKey, AIByIdKey(self.Id))
	return self
}
