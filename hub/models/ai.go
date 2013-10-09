package models

import (
	"appengine/datastore"
	"common"
)

const (
	AIKind    = "AI"
	AllAIsKey = "AIs{All}"
)

type AIs []AI

type AI struct {
	Id    *datastore.Key
	URL   string
	Name  string
	Owner string
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
	return
}

func (self *AI) Save(c common.Context) *AI {
	var err error
	if self.Id == nil {
		self.Id, err = datastore.Put(c, datastore.NewKey(c, AIKind, "", 0, nil), self)
	} else {
		_, err = datastore.Put(c, self.Id, self)
	}
	common.AssertOkError(err)
	common.MemDel(c, AllAIsKey)
	return self
}
