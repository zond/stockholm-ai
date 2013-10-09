package models

import (
	"appengine/datastore"
	"common"
)

const (
	GameKind    = "Game"
	AllGamesKey = "Games{All}"
)

type Games []Game

type Game struct {
	Id      *datastore.Key
	Players []*datastore.Key
}

func findAllGames(c common.Context) (result Games) {
	ids, err := datastore.NewQuery(GameKind).GetAll(c, &result)
	common.AssertOkError(err)
	for index, id := range ids {
		result[index].Id = id
	}
	if result == nil {
		result = Games{}
	}
	return
}

func GetAllGames(c common.Context) (result Games) {
	common.Memoize(c, AllGamesKey, &result, func() interface{} {
		return findAllGames(c)
	})
	return
}

func (self *Game) Save(c common.Context) *Game {
	var err error
	if self.Id == nil {
		self.Id, err = datastore.Put(c, datastore.NewKey(c, GameKind, "", 0, nil), self)
	} else {
		_, err = datastore.Put(c, self.Id, self)
	}
	common.AssertOkError(err)
	common.MemDel(c, AllGamesKey)
	return self
}
