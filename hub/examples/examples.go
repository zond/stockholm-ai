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
		breakpoints := make(sort.Float64Slice, 0, len(node.Edges)-1)
		for index, _ := range breakpoints {
			breakpoints[index] = rand.Float64()
		}
		sort.Sort(breakpoints)
		lastRate := 0.0
		for _, edge := range node.Edges {
			rate := 1.0
			if len(breakpoints) > 0 {
				rate = lastRate - breakpoints[0]
				lastRate = rate
				breakpoints = breakpoints[1:]
			}
			result = append(result, state.Order{
				Src:   edge.Src,
				Dst:   edge.Dst,
				Units: common.Min(node.Units[me], int(float64(node.Units[me])*rate)),
			})
		}
	}
	return
}
