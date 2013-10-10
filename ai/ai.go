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

type AI interface {
	Orders(me state.PlayerId, s state.State) state.Orders
}

func HTTPHandlerFunc(ai AI) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req OrderRequest
		common.MustDecodeJSON(r.Body, &req)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		common.MustEncodeJSON(w, ai.Orders(req.Me, req.State))
	}
}
