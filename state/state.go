package state

import (
	common "github.com/zond/stockholm-ai/common"
	"math/rand"
)

const (
	growthFactor     = 0.2
	starvationFactor = 0.2
)

type NodeId string

type PlayerId string

type GameId string

/*
Edge goes from one node to another.
*/
type Edge struct {
	// Src is the id of the source node.
	Src NodeId
	// Dst is the id of the destination node.
	Dst NodeId
	// Units contains the units in transit along the edge.
	Units []map[PlayerId]int
}

func (self *Edge) init() {
	for index, _ := range self.Units {
		self.Units[index] = make(map[PlayerId]int)
	}
}

/*
Nodes contain units, and are connected by edges.
*/
type Node struct {
	// Id is the id of the node
	Id NodeId
	// Size is the number of units supported by the node. If less units, from a single player, occupy the node they will procreate. If more units occupy the node they will starve.
	Size int
	// Units contain the number of units for each player occupying this node.
	Units map[PlayerId]int
	// Edges go from this node to others.
	Edges map[NodeId]Edge
}

func (self *Node) allReachable(c common.Logger, state *State) bool {
	for nodeId, _ := range state.Nodes {
		if nodeId != self.Id {
			if state.Path(self.Id, nodeId, nil) == nil {
				return false
			}
		}
	}
	return true
}

func (self *Node) Connect(node *Node, edgeLength int) {
	away := &Edge{
		Src:   self.Id,
		Dst:   node.Id,
		Units: make([]map[PlayerId]int, edgeLength),
	}
	away.init()
	here := &Edge{
		Src:   node.Id,
		Dst:   self.Id,
		Units: make([]map[PlayerId]int, edgeLength),
	}
	here.init()
	self.Edges[node.Id] = *away
	node.Edges[self.Id] = *here
}

func (self *Node) connectRandomly(c common.Logger, allNodes []*Node, state *State) {
	minEdges := common.Norm(4, 2, 2, len(allNodes)-1)
	for len(self.Edges) < minEdges || !self.allReachable(c, state) {
		perm := rand.Perm(len(allNodes))
		var randomNode *Node
		for _, index := range perm {
			suggested := allNodes[index]
			if suggested.Id != self.Id {
				if _, found := self.Edges[suggested.Id]; !found {
					randomNode = suggested
					break
				}
			}
		}
		self.Connect(randomNode, common.Norm(3, 1, 1, 5))
		minEdges--
	}
}

/*
NewNode returns a node named id with size.
*/
func NewNode(id NodeId, size int) *Node {
	return &Node{
		Size:  size,
		Id:    id,
		Units: make(map[PlayerId]int),
		Edges: make(map[NodeId]Edge),
	}
}

func (self *State) Add(n *Node) *State {
	self.Nodes[n.Id] = n
	return self
}

/*
RandomNode returns a random node without connections.
*/
func RandomNode() (result *Node) {
	return NewNode(NodeId(common.RandomString(16)), common.Norm(50, 25, 10, 100))
}

/*
Order contains a single order from an AI.

Orders will start units moving along edges of nodes, and they will be unorderable until they arrive.
*/
type Order struct {
	// Src is from where the units should move.
	Src NodeId
	// Dst is where the units should move.
	Dst NodeId
	// Units is the number of units to move.
	Units int
}

/*
Orders are given by AIs to move around their units.
*/
type Orders []Order

type ChangeReason string

type Change struct {
	Units    int
	PlayerId PlayerId
	Reason   ChangeReason
}

type Changes []Change

/*
State completely describes a single turn of the game.
*/
type State struct {
	// Nodes are the nodes in the game.
	Nodes map[NodeId]*Node
	// Changes are the changes and reasons since last turn.
	Changes map[NodeId]Changes
}

func (self *State) executeTransits(logger common.Logger) {
	execution := []func(){}
	for _, node := range self.Nodes {
		for _, edge := range node.Edges {
			for index, units := range edge.Units {
				for playerId, num := range units {
					if num > 0 {
						numCpy := num
						edgeCpy := edge
						playerIdCpy := playerId
						indexCpy := index
						units[playerId] = 0
						if index == len(edge.Units)-1 {
							execution = append(execution, func() {
								self.Nodes[edgeCpy.Dst].Units[playerIdCpy] += numCpy
								self.Changes[edgeCpy.Dst] = append(self.Changes[edgeCpy.Dst], Change{
									Units:    numCpy,
									PlayerId: playerIdCpy,
									Reason:   ChangeReason("Incoming"),
								})
							})
						} else {
							execution = append(execution, func() {
								edgeCpy.Units[indexCpy+1][playerIdCpy] += numCpy
							})
						}
					}
				}
			}
		}
	}
	for _, exec := range execution {
		exec()
	}
}

func (self *State) executeOrders(orderMap map[PlayerId]Orders) {
	execution := []func(){}
	for playerId, orders := range orderMap {
		for _, order := range orders {
			if src, found := self.Nodes[order.Src]; found {
				if edge, found := src.Edges[order.Dst]; found {
					toMove := common.Min(src.Units[playerId], order.Units)
					src.Units[playerId] -= toMove
					edgeCpy := edge
					playerIdCpy := playerId
					execution = append(execution, func() {
						edgeCpy.Units[0][playerIdCpy] += toMove
						self.Changes[edgeCpy.Src] = append(self.Changes[edgeCpy.Src], Change{
							Units:    -toMove,
							PlayerId: playerIdCpy,
							Reason:   ChangeReason("Orders"),
						})
					})
				}
			}
		}
	}
	for _, exec := range execution {
		exec()
	}
}

func (self *State) executeGrowth(c common.Logger) {
	execution := []func(){}
	for _, node := range self.Nodes {
		total := 0
		for _, units := range node.Units {
			total += units
		}
		players := make([]PlayerId, 0, len(node.Units))
		for playerId, _ := range node.Units {
			players = append(players, playerId)
		}
		if len(players) == 1 {
			playerId := players[0]
			units := node.Units[playerId]
			if total > 0 && total < node.Size {
				nodeCpy := node
				newSum := common.Min(node.Size, int(1+float64(units)*(1.0+(growthFactor*(float64(node.Size-total)/float64(node.Size))))))
				if newSum > units {
					execution = append(execution, func() {
						nodeCpy.Units[playerId] = newSum
						self.Changes[nodeCpy.Id] = append(self.Changes[nodeCpy.Id], Change{
							Units:    newSum - units,
							PlayerId: playerId,
							Reason:   ChangeReason("Growth"),
						})
					})
				}
			}
		} else if len(players) > 1 {
			if total > node.Size {
				for playerId, units := range node.Units {
					if units > 0 {
						playerIdCpy := playerId
						nodeCpy := node
						newSum := common.Max(0, int(float64(units)/(1.0+(starvationFactor*(float64(units)/float64(node.Size)))))-1)
						if newSum < units {
							oldSum := units
							execution = append(execution, func() {
								nodeCpy.Units[playerIdCpy] = newSum
								self.Changes[nodeCpy.Id] = append(self.Changes[nodeCpy.Id], Change{
									Units:    newSum - oldSum,
									PlayerId: playerIdCpy,
									Reason:   ChangeReason("Starvation"),
								})
							})
						}
					}
				}
			}
		}
	}
	for _, exec := range execution {
		exec()
	}
}

func (self *State) executeConflicts(l common.Logger) {
	execution := []func(){}
	for _, node := range self.Nodes {
		total := 0
		for _, units := range node.Units {
			total += units
		}
		for playerId, units := range node.Units {
			enemies := total - units
			if units > 0 && enemies > 0 {
				newSum := common.Max(0, common.Min(units-1, int(float64(units)-(float64(enemies)/5.0))))
				playerIdCpy := playerId
				nodeCpy := node
				if newSum < units {
					oldSum := units
					execution = append(execution, func() {
						nodeCpy.Units[playerIdCpy] = newSum
						self.Changes[nodeCpy.Id] = append(self.Changes[nodeCpy.Id], Change{
							Units:    newSum - oldSum,
							PlayerId: playerIdCpy,
							Reason:   ChangeReason("Conflict"),
						})
					})
				}
			}
		}
	}
	for _, exec := range execution {
		exec()
	}
}

func (self *State) onlyPlayerLeft(c common.Logger) *PlayerId {
	players := map[PlayerId]bool{}
	for _, node := range self.Nodes {
		for playerId, units := range node.Units {
			if units > 0 {
				players[playerId] = true
			}
		}
		for _, edge := range node.Edges {
			for _, spot := range edge.Units {
				for playerId, units := range spot {
					if units > 0 {
						players[playerId] = true
					}
				}
			}
		}
	}
	playerSlice := []PlayerId{}
	for playerId, _ := range players {
		playerSlice = append(playerSlice, playerId)
	}
	if len(playerSlice) == 1 {
		return &playerSlice[0]
	}
	return nil
}

/*
Next changes this state into the next state, subject to the provided orders.
*/
func (self *State) Next(c common.Logger, orderMap map[PlayerId]Orders) (winner *PlayerId) {
	self.Changes = map[NodeId]Changes{}
	self.executeTransits(c)
	self.executeOrders(orderMap)
	self.executeGrowth(c)
	self.executeConflicts(c)
	winner = self.onlyPlayerLeft(c)
	return
}

/*
PathFilter will return if a node is accepted by the filter.
*/
type PathFilter func(node *Node) bool

type pathStep struct {
	path []NodeId
	pos  NodeId
}

/*
Path will return the shortest path between src and dst in self, discounting all paths that don't match the filter. A nil filter matches all nodes.
*/
func (self *State) Path(src, dst NodeId, filter PathFilter) (result []NodeId) {
	// queue of paths to try
	queue := []pathStep{
		pathStep{
			path: nil,
			pos:  src,
		},
	}
	// found shortest paths to the nodes
	paths := map[NodeId][]NodeId{
		src: nil,
	}
	// next step preallocated
	step := pathStep{}
	// best path to the dest so far
	var best []NodeId
	// as long as we have new paths to try
	for len(queue) > 0 {
		// pick first path to try
		step = queue[0]
		// pop the queue
		queue = queue[1:]
		// if the node actually exists
		if node, found := self.Nodes[step.pos]; found {
			// for each edge from the node
			for _, edge := range node.Edges {
				// if we either haven't been where this edge leads before, or we would get there along a shorter path this time (*1)
				if lastPathHere, found := paths[edge.Dst]; !found || len(step.path)+len(edge.Units)+1 < len(lastPathHere) {
					// if we either haven't found dst yet, or if following this path is shorter than where we found dst
					if best == nil || len(step.path)+len(edge.Units)+1 < len(best) {
						// if we aren't filtering nodes, or this node matches the filter
						if filter == nil || filter(node) {
							// make a new path that is the path here + this node + the edge we want to follow
							thisPath := make([]NodeId, len(step.path)+1+len(edge.Units))
							// copy the path to here to the new path
							copy(thisPath, step.path)
							// add this node
							thisPath[len(step.path)] = edge.Dst
							// remember that this is the best way so far (guaranteed by *1)
							paths[edge.Dst] = thisPath
							// if this path leads to dst
							if edge.Dst == dst {
								best = thisPath
							}
							// queue up following this path further
							queue = append(queue, pathStep{
								path: thisPath,
								pos:  edge.Dst,
							})
						}
					}
				}
			}
		}
	}
	return paths[dst]
}

func NewState() *State {
	return &State{
		Nodes:   map[NodeId]*Node{},
		Changes: map[NodeId]Changes{},
	}
}

/*
RandomState creates a random state for the provided players.
*/
func RandomState(c common.Logger, players []PlayerId) (result *State) {
	result = NewState()
	size := common.Norm(len(players)*6, len(players), len(players)*4, len(players)*10)
	allNodes := make([]*Node, 0, size)
	for i := 0; i < size; i++ {
		node := RandomNode()
		result.Nodes[node.Id] = node
		allNodes = append(allNodes, node)
	}
	for _, node := range allNodes {
		node.connectRandomly(c, allNodes, result)
	}
	perm := rand.Perm(len(allNodes))
	for index, playerId := range players {
		allNodes[perm[index]].Units[playerId] = 10
	}
	return
}
