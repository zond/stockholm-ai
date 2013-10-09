package models

import (
	"appengine/datastore"
	"common"
	"fmt"
)

const (
	TurnKind = "Turn"
)

func turnsKeyForParent(k interface{}) string {
	return fmt.Sprintf("Turns{Parent:%v}", k)
}

type Turns []Turn

func (self Turns) process(c common.Context) Turns {
	for index, _ := range self {
		(&self[index]).process(c)
	}
	return self
}

type Turn struct {
	Id              *datastore.Key
	Ordinal         int
	SerializedState []byte `json:"-"`
	State           State  `datastore:"-"`
}

func (self *Turn) process(c common.Context) *Turn {
	if len(self.SerializedState) > 0 {
		common.MustUnmarshalJSON(self.SerializedState, &self.State)
	}
	return self
}

func findTurnsByParent(c common.Context, parent *datastore.Key) (result Turns) {
	ids, err := datastore.NewQuery(TurnKind).Ancestor(parent).GetAll(c, &result)
	common.AssertOkError(err)
	for index, id := range ids {
		result[index].Id = id
	}
	if result == nil {
		result = Turns{}
	}
	return
}

func GetTurnsByParent(c common.Context, parent *datastore.Key) (result Turns) {
	common.Memoize(c, turnsKeyForParent(parent), &result, func() interface{} {
		return findTurnsByParent(c, parent)
	})
	return result.process(c)
}

func (self *Turn) Save(c common.Context, parent *datastore.Key) *Turn {
	self.SerializedState = common.MustMarshalJSON(self.State)
	var err error
	if self.Id == nil {
		self.Id, err = datastore.Put(c, datastore.NewKey(c, TurnKind, "", 0, parent), self)
	} else {
		_, err = datastore.Put(c, self.Id, self)
	}
	common.AssertOkError(err)
	common.MemDel(c, turnsKeyForParent(parent))
	return self
}
