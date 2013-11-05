package models

import (
	"appengine"
	"appengine/datastore"
	"appengine/delay"
	"appengine/urlfetch"
	"bytes"
	"common"
	"encoding/json"
	"fmt"
	ai "github.com/zond/stockholm-ai/ai"
	aiCommon "github.com/zond/stockholm-ai/common"
	"github.com/zond/stockholm-ai/state"
	"io"
	"net/http"
	"time"
)

const (
	GameKind        = "Game"
	allGamesKey     = "Games{All}"
	maxGameDuration = 100
)

var nextTurnFunc *delay.Function

func init() {
	nextTurnFunc = delay.Func("models/game.nextTurnFunc", nextTurn)
}

func gameKeyForId(k interface{}) string {
	return fmt.Sprintf("Game{Id:%v}", k)
}

func gamePageKey(limit, offset int) string {
	return fmt.Sprintf("GamePage{limit:%v,offset:%v}", limit, offset)
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

func (self Games) Less(j, i int) bool {
	return self[i].CreatedAt.Before(self[j].CreatedAt)
}

func (self Games) Swap(i, j int) {
	self[j], self[i] = self[i], self[j]
}

func (self Games) process(c common.Context) Games {
	for index, _ := range self {
		(&self[index]).fastProcess(c)
	}
	return self
}

type Game struct {
	Id          *datastore.Key
	Players     []*datastore.Key
	Winner      *datastore.Key
	State       GameState
	PlayerNames []string `datastore:"-"`
	WinnerName  string   `datastore:"-"`
	Turns       Turns    `datastore:"-"`
	Length      int
	CreatedAt   time.Time
}

type orderResponse struct {
	DatastorePlayerId *datastore.Key
	StatePlayerId     state.PlayerId
	Orders            state.Orders
	Error             error
}

type orderError struct {
	Request      *http.Request
	RequestBody  string
	Response     *http.Response
	ResponseBody string
}

func (self orderError) Error() string {
	return fmt.Sprintf("Got %v from %v", self.Response.StatusCode, self.Request.URL)
}

func nextTurn(cont appengine.Context, id *datastore.Key, playerNames []string) {
	con := common.Context{Context: cont}
	self := getGameById(con, id)
	self.PlayerNames = playerNames
	if self.Length > maxGameDuration {
		self.State = StateFinished
		self.Save(con)
		con.Infof("Ended %v due to timeout", self.Id)
		return
	}
	errorSavers := []func(){}
	if err := common.Transaction(con, func(c common.Context) (err error) {
		lastTurn := GetLatestTurnByParent(c, self.Id)
		responses := make(chan orderResponse, len(self.Players))
		for _, playerId := range self.Players {
			orderResp := orderResponse{
				DatastorePlayerId: playerId,
				StatePlayerId:     state.PlayerId(playerId.Encode()),
			}
			if foundAi := GetAIById(c, playerId); foundAi != nil {
				go func() {
					// Always deliver the order response
					defer func() {
						responses <- orderResp
					}()

					// create a request
					orderRequest := ai.OrderRequest{
						Me:          orderResp.StatePlayerId,
						State:       lastTurn.State,
						GameId:      state.GameId(self.Id.Encode()),
						TurnOrdinal: lastTurn.Ordinal,
					}

					// encode it into a body, and remember its string representation
					sendBody := &bytes.Buffer{}
					aiCommon.MustEncodeJSON(sendBody, orderRequest)
					sendBodyString := sendBody.String()

					// get a client
					client := urlfetch.Client(c)

					// send the request to the ai
					req, err := http.NewRequest("POST", foundAi.URL, sendBody)
					var resp *http.Response
					if err == nil {
						req.Header.Set("Content-Type", "application/json; charset=UTF-8")
						resp, err = client.Do(req)
					}

					recvBody := &bytes.Buffer{}
					recvBodyString := ""
					if err == nil {
						// check what we received
						_, err = io.Copy(recvBody, resp.Body)
						recvBodyString = recvBody.String()
					}
					// if we have no other errors, but we got a non-200
					if err == nil && resp.StatusCode != 200 {
						err = orderError{
							Request:      req,
							RequestBody:  sendBodyString,
							Response:     resp,
							ResponseBody: recvBodyString,
						}
					}

					// lets try to unserialize
					if err == nil {
						err = json.Unmarshal(recvBody.Bytes(), &orderResp.Orders)
					}

					// store the error, if any
					if err != nil {
						orderResp.Error = err
					}
				}()
			} else {
				responses <- orderResp
			}
		}
		orderMap := map[state.PlayerId]state.Orders{}
		for _, _ = range self.Players {
			// wait for the responses
			orderResp := <-responses
			// store it
			orderMap[orderResp.StatePlayerId] = orderResp.Orders
			// if we got an error
			if orderResp.Error != nil {
				// make sure to save it later
				errorSavers = append(errorSavers, func() {
					if ai := GetAIById(con, orderResp.DatastorePlayerId); ai != nil {
						ai.AddError(con, lastTurn.Id, orderResp.Error)
					}
				})
			}
		}
		// execute the orders
		newTurn, winner := lastTurn.Next(c, orderMap)
		// save the new turn
		newTurn.Save(c, self.Id)
		// if we got a winner, end the game and store the winner
		if winner == nil {
			self.State = StatePlaying
		} else {
			self.Winner = common.MustDecodeKey(string(*winner))
			self.State = StateFinished
		}
		// increase our length with the new turn
		self.Length += 1
		// save us
		self.Save(c)
		// if we didn't end, queue the next turn
		if winner == nil {
			nextTurnFunc.Call(c, self.Id, playerNames)
		}
		return nil
	}); err != nil {
		panic(err)
	}
	// run any error savers we got
	for _, saver := range errorSavers {
		saver()
	}
	// store the new stats in the players if we ended
	if self.State == StateFinished {
		for _, playerId := range self.Players {
			common.Transaction(con, func(c common.Context) error {
				if ai := GetAIById(c, playerId); ai != nil {
					if playerId.Equal(self.Winner) {
						ai.Wins += 1
					} else {
						ai.Losses += 1
					}
					ai.Save(c)
				}
				return nil
			})
		}
	}
}

func (self *Game) setPlayerNames(c common.Context) {
	self.PlayerNames = make([]string, len(self.Players))
	for index, id := range self.Players {
		if ai := GetAIById(c, id); ai != nil {
			self.PlayerNames[index] = ai.Name
			if ai.Id.Equal(self.Winner) {
				self.WinnerName = ai.Name
			}
		} else {
			self.PlayerNames[index] = "[redacted]"
		}
	}
}

func (self *Game) fastProcess(c common.Context) *Game {
	self.setPlayerNames(c)
	return self
}

func (self *Game) process(c common.Context) *Game {
	self.Turns = GetTurnsByParent(c, self.Id)
	self.setPlayerNames(c)
	return self
}

type GamePage struct {
	Content Games
	Total   int
}

func findGamePage(c common.Context, offset, limit int) (result GamePage) {
	query := datastore.NewQuery(GameKind)
	var err error
	result.Total, err = query.Count(c)
	common.AssertOkError(err)
	var ids []*datastore.Key
	ids, err = query.Limit(limit).Offset(offset).Order("-CreatedAt").GetAll(c, &result.Content)
	common.AssertOkError(err)
	for index, id := range ids {
		result.Content[index].Id = id
	}
	if result.Content == nil {
		result.Content = Games{}
	}
	return
}

func GetGamePage(c common.Context, offset, limit int) (result GamePage) {
	common.Memoize2(c, allGamesKey, gamePageKey(limit, offset), &result, func() interface{} {
		return findGamePage(c, offset, limit)
	})
	result.Content.process(c)
	return result
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
		self.setPlayerNames(c)
		err = common.Transaction(c, func(c common.Context) (err error) {
			self.CreatedAt = time.Now()
			self.State = StateCreated
			self.Length = 1
			self.Id, err = datastore.Put(c, datastore.NewKey(c, GameKind, "", 0, nil), self)
			if err != nil {
				return
			}
			playerIds := make([]state.PlayerId, 0, len(self.Players))
			for _, id := range self.Players {
				playerIds = append(playerIds, state.PlayerId(id.Encode()))
			}
			turn := &Turn{
				State: state.RandomState(common.GAELogger{c}, playerIds),
			}
			turn.Save(c, self.Id)
			self.Turns = Turns{*turn}
			nextTurnFunc.Call(c, self.Id, self.PlayerNames)
			return nil
		})
		if err == nil {
			for _, playerId := range self.Players {
				common.Transaction(c, func(common.Context) error {
					if ai := GetAIById(c, playerId); ai != nil {
						ai.Games += 1
						ai.Save(c)
					}
					return nil
				})
			}
		}
	} else {
		_, err = datastore.Put(c, self.Id, self)
	}
	common.AssertOkError(err)
	common.MemDel(c, allGamesKey, gameKeyForId(self.Id))
	return self.process(c)
}
