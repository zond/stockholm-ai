package ai

import (
	"github.com/zond/stockholm-ai/common"
	"github.com/zond/stockholm-ai/state"
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

/*
nearestEmpty returns the path to the nearest node to src in s that has no units belonging to me.
*/
func (self Simpleton) nearestEmpty(me state.PlayerId, src state.NodeId, s *state.State) (result []state.NodeId) {
	// For each node in s
	for _, node := range s.Nodes {
		// If I have no units
		if node.Units[me] == 0 {
			// Calculate the path to the node, and if we have no result or this result is better
			if thisDist := s.Path(src, node.Id, nil); result == nil || len(thisDist) < len(result) {
				// Set the result to the path
				result = thisDist
			}
		}
	}
	return
}

/*
Orders will return orders for all nodes in s where me has more than 2 units, that moves half the units to the nearest
node in s where me has no units.
*/
func (self Simpleton) Orders(logger common.Logger, me state.PlayerId, s *state.State) (result state.Orders) {
	// For each node in s
	for _, node := range s.Nodes {
		// If me has more than 2 units there
		if units := node.Units[me]; units > 2 {
			// If there is a path to a another node (the nearest one) without units belonging to me
			if nearestEmpty := self.nearestEmpty(me, node.Id, s); len(nearestEmpty) > 0 {
				// Add an order moving half the units along the path to the other node
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
