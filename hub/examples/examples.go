package examples

import (
	common "github.com/zond/stockholm-ai/common"
	state "github.com/zond/stockholm-ai/state"
	"math/rand"
	"sort"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Randomizer struct{}

func (self Randomizer) Orders(logger common.Logger, me state.PlayerId, s state.State) (result state.Orders) {
	for _, node := range s.Nodes {
		if node.Units[me] > 0 {
			breakpoints := make(sort.Float64Slice, len(node.Edges))
			for index, _ := range breakpoints {
				breakpoints[index] = rand.Float64()
			}
			sort.Sort(breakpoints)
			lastRate := 0.0
			rate := 1.0
			for _, edge := range node.Edges {
				rate = breakpoints[0] - lastRate
				lastRate = rate
				breakpoints = breakpoints[1:]
				units := common.Min(node.Units[me], int(float64(node.Units[me])*rate))
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
	return
}
