package models

import (
	"fmt"
	"sort"
	"time"

	"github.com/zond/stockholm-ai/hub/common"
	"github.com/zond/stockholm-ai/state"
	"google.golang.org/appengine/datastore"
)

const (
	TurnKind = "Turn"
)

func countTurnsKeyForParent(k interface{}) string {
	return fmt.Sprintf("Turns{Count,Parent:%v}", k)
}

func turnsKeyForParent(k interface{}) string {
	return fmt.Sprintf("Turns{Parent:%v}", k)
}

func givenTurnKeyForParent(k interface{}, o interface{}) string {
	return fmt.Sprintf("Turns{Parent:%v,Ordinal:%v}", k, o)
}

func latestTurnKeyForParent(k interface{}) string {
	return fmt.Sprintf("Turns{Latest,Parent:%v}", k)
}

type Turns []Turn

func (self Turns) Len() int {
	return len(self)
}

func (self Turns) Less(i, j int) bool {
	return self[i].Ordinal < self[j].Ordinal
}

func (self Turns) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self Turns) process(c common.Context) Turns {
	for index, _ := range self {
		(&self[index]).process(c)
	}
	return self
}

type Turn struct {
	Id              *datastore.Key
	Ordinal         int
	SerializedState []byte       `json:"-"`
	State           *state.State `datastore:"-"`
	CreatedAt       time.Time
}

func (self *Turn) Next(c common.Context, orderMap map[state.PlayerId]state.Orders) (*Turn, *state.PlayerId) {
	cpy := *self
	cpy.Id = nil
	cpy.Ordinal += 1
	winner := cpy.State.Next(common.GAELogger{c}, orderMap)
	return &cpy, winner
}

func (self *Turn) process(c common.Context) *Turn {
	if len(self.SerializedState) > 0 {
		self.State = &state.State{}
		common.MustUnmarshal(self.SerializedState, self.State)
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
	sort.Sort(result)
	return result.process(c)
}

func findGivenTurnByParent(c common.Context, parent *datastore.Key, ordinal int) *Turn {
	var turns Turns
	ids, err := datastore.NewQuery(TurnKind).Ancestor(parent).Filter("Ordinal=", ordinal).Limit(1).GetAll(c, &turns)
	common.AssertOkError(err)
	for index, id := range ids {
		turns[index].Id = id
	}
	if len(turns) == 0 {
		return nil
	}
	return &turns[0]
}

func GetGivenTurnByParent(c common.Context, parent *datastore.Key, ordinal int) *Turn {
	var result Turn
	if common.Memoize(c, givenTurnKeyForParent(parent, ordinal), &result, func() interface{} {
		return findGivenTurnByParent(c, parent, ordinal)
	}) {
		return (&result).process(c)
	}
	return nil
}

func findLatestTurnByParent(c common.Context, parent *datastore.Key) *Turn {
	var turns Turns
	ids, err := datastore.NewQuery(TurnKind).Ancestor(parent).Order("-CreatedAt").Limit(1).GetAll(c, &turns)
	common.AssertOkError(err)
	for index, id := range ids {
		turns[index].Id = id
	}
	if len(turns) == 0 {
		return nil
	}
	return &turns[0]
}

func GetLatestTurnByParent(c common.Context, parent *datastore.Key) *Turn {
	var result Turn
	if common.Memoize(c, latestTurnKeyForParent(parent), &result, func() interface{} {
		return findLatestTurnByParent(c, parent)
	}) {
		return (&result).process(c)
	}
	return nil
}

func (self *Turn) Save(c common.Context, parent *datastore.Key) *Turn {
	self.SerializedState = common.MustMarshal(self.State)
	var err error
	if self.Id == nil {
		self.CreatedAt = time.Now()
		self.Id, err = datastore.Put(c, datastore.NewKey(c, TurnKind, "", 0, parent), self)
		common.AssertOkError(err)
		common.MemPut(c, latestTurnKeyForParent(parent), self)
	} else {
		_, err = datastore.Put(c, self.Id, self)
		common.AssertOkError(err)
	}
	common.MemDel(c, turnsKeyForParent(parent), latestTurnKeyForParent(parent), givenTurnKeyForParent(parent, self.Ordinal))
	return self
}
