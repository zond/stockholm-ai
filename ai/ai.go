package ai

import (
	"github.com/zond/stockholm-ai/common"
	"github.com/zond/stockholm-ai/state"
	"net/http"
	"runtime/debug"
)

/*
OrderRequest encapsulates the JSON body in the requests from the hub to the
AI players for each turn.
*/
type OrderRequest struct {
	// Me is the id used to represent the receiving AI in the state.
	Me state.PlayerId
	// GameId is the unique id of the game, if the AI wants to keep state between order requests.
	GameId state.GameId
	// State contains all the state of the game for the turn the order request refers to.
	State *state.State
	// The ordinal of the turn we want orders for
	TurnOrdinal int
	// The IDs of the AIs in the game, for those interested
	AIs map[state.PlayerId]string
}

/*
AI has to be implemented by any player of the game.
*/
type AI interface {
	// Orders returns the orders the AI playing me wants to issue at the turn described by s.
	Orders(logger common.Logger, req OrderRequest) state.Orders
}

/*
HTTPHandlerFunc returns an http.HandlerFunc to use when hosting an AI.
*/
func HTTPHandlerFunc(lf common.LoggerFactory, ai AI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := lf(r)
		defer func() {
			if e := recover(); e != nil {
				w.WriteHeader(500)
				logger.Printf("Error delivering orders: %v\n%v", e, string(debug.Stack()))
			}
		}()
		var req OrderRequest
		common.MustDecodeJSON(r.Body, &req)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		common.MustEncodeJSON(w, ai.Orders(lf(r), req))
	}
}
