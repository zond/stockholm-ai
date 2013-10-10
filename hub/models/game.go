package models

import (
	"appengine"
	"appengine/datastore"
	"appengine/delay"
	"common"
	"fmt"
	"sort"
	"time"
)

const (
	GameKind        = "Game"
	AllGamesKey     = "Games{All}"
	maxGameDuration = 100
)

var nextTurnFunc *delay.Function

func init() {
	nextTurnFunc = delay.Func("models/game.nextTurnFunc", nextTurn)
}

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

func nextTurn(cont appengine.Context, id *datastore.Key, playerNames []string) {
	c := common.Context{Context: cont}
	self := getGameById(c, id)
	self.PlayerNames = playerNames
	if CountTurnsByParent(c, self.Id) > maxGameDuration {
		self.State = StateFinished
		self.Save(c)
		c.Infof("Ended %v due to timeout", self.Id)
		return
	}
	if err := common.Transaction(c, func(c common.Context) (err error) {
		lastTurn := GetLatestTurnByParent(c, self.Id)
		newTurn := lastTurn.Next()
		newTurn.Save(c, self.Id)
		self.State = StatePlaying
		self.Save(c)
		nextTurnFunc.Call(c, self.Id, playerNames)
		return nil
	}); err != nil {
		panic(err)
	}
}

func (self *Game) setPlayerNames(c common.Context) {
	self.PlayerNames = make([]string, len(self.Players))
	for index, id := range self.Players {
		if ai := GetAIById(c, id); ai != nil {
			self.PlayerNames[index] = ai.Name
		} else {
			self.PlayerNames[index] = "[redacted]"
		}
	}
}

func (self *Game) process(c common.Context) *Game {
	self.Turns = GetTurnsByParent(c, self.Id)
	self.setPlayerNames(c)
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

func getGameById(c common.Context, id *datastore.Key) *Game {
	var game Game
	if common.Memoize(c, gameKeyForId(id), &game, func() interface{} {
		return findGameById(c, id)
	}) {
		return &game
	}
	return nil
}

func GetGameById(c common.Context, id *datastore.Key) (result *Game) {
	result = getGameById(c, id)
	if result != nil {
		result.process(c)
	}
	return
}

func (self *Game) Save(c common.Context) *Game {
	var err error
	if self.Id == nil {
		err = common.Transaction(c, func(c common.Context) (err error) {
			self.CreatedAt = time.Now()
			self.State = StateCreated
			self.Id, err = datastore.Put(c, datastore.NewKey(c, GameKind, "", 0, nil), self)
			if err != nil {
				return
			}
			playerIds := make([]PlayerId, 0, len(self.Players))
			for _, id := range self.Players {
				playerIds = append(playerIds, PlayerId(id.Encode()))
			}
			turn := &Turn{
				State: RandomState(c, playerIds),
			}
			turn.Save(c, self.Id)
			self.Turns = Turns{*turn}
			self.setPlayerNames(c)
			nextTurnFunc.Call(c, self.Id, self.PlayerNames)
			return nil
		})
	} else {
		_, err = datastore.Put(c, self.Id, self)
	}
	common.AssertOkError(err)
	common.MemDel(c, AllGamesKey, gameKeyForId(self.Id))
	return self.process(c)
}
