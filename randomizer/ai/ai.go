package ai

import (
	"github.com/zond/stockholm-ai/common"
	"github.com/zond/stockholm-ai/state"
	"math/rand"
	"sort"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

/*
Randomizer is a randomly moving AI implementation.
*/
type Randomizer struct{}

/*
Orders will return orders randomly moving up to 20% of all troops in each occupied node along randomly chosen edges of that node.
*/
func (self Randomizer) Orders(logger common.Logger, me state.PlayerId, s *state.State) (result state.Orders) {
	// Go through all nodes
	for _, node := range s.Nodes {
		// If I have units here
		if node.Units[me] > 0 {
			// Create a slice with room for one float per edge in this node
			breakpoints := make(sort.Float64Slice, len(node.Edges))
			// Fill it with numbers between 0 and 0.2
			for index, _ := range breakpoints {
				breakpoints[index] = rand.Float64() / 5
			}
			// Sort the numbers
			sort.Sort(breakpoints)
			lastRate := 0.0
			rate := 1.0
			// For each edge in this node
			for _, edge := range node.Edges {
				// Take the fraction between the first float in the slice and the last picked float (or zero)
				rate = breakpoints[0] - lastRate
				lastRate = rate
				// Remove the first float from the slice
				breakpoints = breakpoints[1:]
				// Calculate how many units the fraction corresponds to
				units := common.Min(node.Units[me], int(float64(node.Units[me])*rate))
				// If any, create a move order for those units
				if units > 0 {
					result = append(result, state.Order{
						Src:   edge.Src,
						Dst:   edge.Dst,
						Units: units,
					})
				}
			}
		}
	}
	// Return the created orders
	return
}
