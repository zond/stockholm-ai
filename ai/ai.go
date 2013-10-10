package ai

import (
	common "github.com/zond/stockholm-ai/common"
	state "github.com/zond/stockholm-ai/state"
	"net/http"
)

type OrderRequest struct {
	Me    state.PlayerId
	State state.State
}

type LoggerFactory func(r *http.Request) common.Logger

type AI interface {
	Orders(logger common.Logger, me state.PlayerId, s state.State) state.Orders
}

func HTTPHandlerFunc(lf LoggerFactory, ai AI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req OrderRequest
		common.MustDecodeJSON(r.Body, &req)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		common.MustEncodeJSON(w, ai.Orders(lf(r), req.Me, req.State))
	}
}
