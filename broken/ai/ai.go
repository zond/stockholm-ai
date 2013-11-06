package ai

import (
	"github.com/zond/stockholm-ai/ai"
	"github.com/zond/stockholm-ai/common"
	"github.com/zond/stockholm-ai/state"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

/*
Broken will just return 500s.
*/
type Broken struct{}

func (self Broken) Orders(logger common.Logger, req ai.OrderRequest) (result state.Orders) {
	panic("Oh noes")
}
