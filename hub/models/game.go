package models

import (
	"appengine/datastore"
	"common"
	"fmt"
	"sort"
	"time"
)

const (
	GameKind    = "Game"
	AllGamesKey = "Games{All}"
)

func gameKeyForId(k interface{}) string {
	return fmt.Sprintf("Game{Id:%v}", k)
}

type GameState string

const (
	StateCreated  GameState = "Created"
	StatePlaying  GameState = "Playing"
	StateFinished GameState = "Finished"
)

type Games []Game

func (self Games) Len() int {
	return len(self)
}

func (self Games) Less(i, j int) bool {
	return self[i].CreatedAt.Before(self[j].CreatedAt)
}

func (self Games) Swap(i, j int) {
	self[j], self[i] = self[i], self[j]
}

func (self Games) process(c common.Context) Games {
	for index, _ := range self {
		(&self[index]).process(c)
	}
	return self
}

type Game struct {
	Id          *datastore.Key
	Players     []*datastore.Key
	State       GameState
	PlayerNames []string `datastore:"-"`
	Turns       Turns    `datastore:"-"`
	CreatedAt   time.Time
}

func (self *Game) process(c common.Context) *Game {
	self.PlayerNames = make([]string, len(self.Players))
	self.Turns = GetTurnsByParent(c, self.Id)
	for index, id := range self.Players {
		if ai := GetAIById(c, id); ai != nil {
			self.PlayerNames[index] = ai.Name
		} else {
			self.PlayerNames[index] = "[redacted]"
		}
	}
	return self
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
	sort.Sort(result)
	return result.process(c)
}

func findGameById(c common.Context, id *datastore.Key) *Game {
	var game Game
	err := datastore.Get(c, id, &game)
	if err == datastore.ErrNoSuchEntity {
		return nil
	}
	common.AssertOkError(err)
	game.Id = id
	return &game
}

func GetGameById(c common.Context, id *datastore.Key) *Game {
	var game Game
	if common.Memoize(c, gameKeyForId(id), &game, func() interface{} {
		return findGameById(c, id)
	}) {
		return (&game).process(c)
	}
	return nil
}

func (self *Game) Save(c common.Context) *Game {
	var err error
	if self.Id == nil {
		self.State = StateCreated
		self.Id, err = datastore.Put(c, datastore.NewKey(c, GameKind, "", 0, nil), self)
		playerIds := make([]PlayerId, 0, len(self.Players))
		for _, id := range self.Players {
			playerIds = append(playerIds, PlayerId(id.Encode()))
		}
		turn := &Turn{
			State: RandomState(c, playerIds),
		}
		turn.Save(c, self.Id)
	} else {
		_, err = datastore.Put(c, self.Id, self)
	}
	common.AssertOkError(err)
	common.MemDel(c, AllGamesKey, gameKeyForId(self.Id))
	return self.process(c)
}
