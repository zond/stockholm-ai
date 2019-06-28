package models

import (
	"appengine/datastore"
	"fmt"
	"github.com/zond/stockholm-ai/hub/common"
	"sort"
	"time"
)

const (
	AIKind      = "AI"
	AIErrorKind = "AIError"
	AllAIsKey   = "AIs{All}"
)

func aIByIdKey(k interface{}) string {
	return fmt.Sprintf("AI{Id:%v}", k)
}

func aiErrorsKeyByParent(k interface{}) string {
	return fmt.Sprintf("AIErrors{Parent:%v}", k)
}

type AIError struct {
	Turn              *datastore.Key
	Game              *datastore.Key `datastore:"-"`
	Error             string         `datastore:"-"`
	ErrorDetail1      string         `datastore:"-"`
	ErrorDetail2      string         `datastore:"-"`
	ErrorBytes        []byte         `json:"-"`
	ErrorDetail1Bytes []byte         `json:"-"`
	ErrorDetail2Bytes []byte         `json:"-"`
	CreatedAt         time.Time
}

func (self *AIError) process(c common.Context) *AIError {
	self.Error = string(self.ErrorBytes)
	self.ErrorDetail1 = string(self.ErrorDetail1Bytes)
	self.ErrorDetail2 = string(self.ErrorDetail2Bytes)
	self.Game = self.Turn.Parent()
	return self
}

type AIErrors []AIError

func (self AIErrors) process(c common.Context) AIErrors {
	for index, _ := range self {
		(&self[index]).process(c)
	}
	return self
}

func (self AIErrors) Len() int {
	return len(self)
}

func (self AIErrors) Less(j, i int) bool {
	return self[i].CreatedAt.Before(self[j].CreatedAt)
}

func (self AIErrors) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
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
	Games     int
	Wins      int
	Losses    int
	Owner     string `json:"-"`
	IsOwner   bool   `datastore:"-"`
	CreatedAt time.Time
}

func (self *AI) AddError(c common.Context, turnId *datastore.Key, err error) {
	_, e := datastore.Put(c, datastore.NewKey(c, AIErrorKind, "", 0, self.Id), &AIError{
		CreatedAt:         time.Now(),
		Turn:              turnId,
		ErrorBytes:        []byte(err.Error()),
		ErrorDetail1Bytes: []byte(fmt.Sprintf("%+v", err)),
		ErrorDetail2Bytes: []byte(fmt.Sprintf("%#v", err)),
	})
	if e != nil {
		c.Errorf("Got %+v when trying to save a new error!", e)
	}
	common.MemDel(c, aiErrorsKeyByParent(self.Id))
}

func (self *AI) findErrors(c common.Context) (result AIErrors) {
	_, err := datastore.NewQuery(AIErrorKind).Ancestor(self.Id).Order("-CreatedAt").Limit(50).GetAll(c, &result)
	common.AssertOkError(err)
	if result == nil {
		result = AIErrors{}
	}
	return
}

func (self *AI) GetErrors(c common.Context) (result AIErrors) {
	common.Memoize(c, aiErrorsKeyByParent(self.Id), &result, func() interface{} {
		return self.findErrors(c)
	})
	sort.Sort(result)
	return result.process(c)
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
	if common.Memoize(c, aIByIdKey(id), &ai, func() interface{} {
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
	common.MemDel(c, AllAIsKey, aIByIdKey(self.Id))
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
	common.MemDel(c, AllAIsKey, aIByIdKey(self.Id))
	return self
}
