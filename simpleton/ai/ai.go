package ai

import (
	common "github.com/zond/stockholm-ai/common"
	state "github.com/zond/stockholm-ai/state"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

/*
Simpleton will just try to colonize as much as possible by always sending half the units on each colonized node to the nearest uncolonized node.
*/
type Simpleton struct{}

func (self Simpleton) nearestEmpty(me state.PlayerId, src state.NodeId, s *state.State) (result []state.NodeId) {
	for _, node := range s.Nodes {
		if node.Units[me] == 0 {
			if thisDist := s.Path(src, node.Id, nil); result == nil || len(thisDist) < len(result) {
				result = thisDist
			}
		}
	}
	return
}

func (self Simpleton) Orders(logger common.Logger, me state.PlayerId, s *state.State) (result state.Orders) {
	for _, node := range s.Nodes {
		if units := node.Units[me]; units > 2 {
			if nearestEmpty := self.nearestEmpty(me, node.Id, s); len(nearestEmpty) > 0 {
				result = append(result, state.Order{
					Src:   node.Id,
					Dst:   nearestEmpty[0],
					Units: units / 2,
				})
			}
		}
	}
	return
}
