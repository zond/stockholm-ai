package models

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

func (self *Node) reachable(c common.Logger, node *Node, nodeMap map[NodeId]*Node) bool {
	visited := map[NodeId]bool{}
	queue := make([]NodeId, 0, len(self.Edges))
	for nodeId, _ := range self.Edges {
		if nodeId == node.Id {
			return true
		}
		visited[nodeId] = true
		queue = append(queue, nodeId)
	}
	for len(queue) > 0 {
		newQueue := []NodeId{}
		for _, nodeId := range queue {
			for edgeNodeId, _ := range nodeMap[nodeId].Edges {
				if edgeNodeId == node.Id {
					return true
				}
				if !visited[edgeNodeId] {
					newQueue = append(newQueue, edgeNodeId)
					visited[edgeNodeId] = true
				}
			}
		}
		queue = newQueue
	}
	return false
}

func (self *Node) allReachable(c common.Logger, nodeMap map[NodeId]*Node) bool {
	for _, node := range nodeMap {
		if !self.reachable(c, node, nodeMap) {
			return false
		}
	}
	return true
}

func (self *Node) connectNode(node *Node) {
	edgeLength := common.Norm(3, 1, 1, 5)
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

func (self *Node) connect(c common.Logger, allNodes []*Node, nodeMap map[NodeId]*Node) {
	minEdges := common.Norm(4, 2, 2, len(allNodes)-1)
	for len(self.Edges) < minEdges || !self.allReachable(c, nodeMap) {
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
		self.connectNode(randomNode)
		minEdges--
	}
}

/*
RandomNode returns a random node without connections.
*/
func RandomNode() (result *Node) {
	result = &Node{
		Size:  common.Norm(50, 25, 10, 100),
		Id:    NodeId(common.RandomString(16)),
		Units: make(map[PlayerId]int),
		Edges: make(map[NodeId]Edge),
	}
	return
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

/*
State completely describes a single turn of the game.
*/
type State struct {
	// Nodes are the nodes in the game.
	Nodes map[NodeId]*Node
}

func (self *State) executeTransits(logger common.Logger) {
	execution := []func(){}
	for _, node := range self.Nodes {
		for _, edge := range node.Edges {
			for index, units := range edge.Units {
				for playerId, num := range units {
					numCpy := num
					edgeCpy := edge
					playerIdCpy := playerId
					indexCpy := index
					units[playerId] = 0
					if index == len(edge.Units)-1 {
						execution = append(execution, func() {
							self.Nodes[edgeCpy.Dst].Units[playerIdCpy] += numCpy
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
	for _, exec := range execution {
		exec()
	}
}

func (self *State) executeOrders(orderMap map[PlayerId]Orders) {
	execution := []func(){}
	for playerId, orders := range orderMap {
		for _, order := range orders {
			if src, found := self.Nodes[order.Src]; found {
				if src.Units[playerId] >= order.Units {
					if edge, found := src.Edges[order.Dst]; found {
						src.Units[playerId] -= order.Units
						edgeCpy := edge
						orderCpy := order
						playerIdCpy := playerId
						execution = append(execution, func() {
							edgeCpy.Units[0][playerIdCpy] += orderCpy.Units
						})
					}
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
			if total < node.Size {
				nodeCpy := node
				newSum := common.Min(node.Size, int(1+float64(units)*(1.0+(growthFactor*(float64(node.Size-total)/float64(node.Size))))))
				execution = append(execution, func() {
					nodeCpy.Units[playerId] = newSum
				})
			}
		} else if len(players) > 1 {
			for playerId, units := range node.Units {
				if total > node.Size {
					playerIdCpy := playerId
					nodeCpy := node
					newSum := common.Max(0, int(float64(units)/(1.0+(starvationFactor*(float64(units)/float64(node.Size))))-1))
					execution = append(execution, func() {
						nodeCpy.Units[playerIdCpy] = newSum
					})
				}
			}
		}
	}
	for _, exec := range execution {
		exec()
	}
}

func (self *State) executeConflicts() {
	execution := []func(){}
	for _, node := range self.Nodes {
		total := 0
		for _, units := range node.Units {
			total += units
		}
		for playerId, units := range node.Units {
			enemies := total - units
			if enemies > 0 {
				newSum := common.Max(0, common.Min(units-1, int(float64(units)-(float64(enemies)/5.0))))
				playerIdCpy := playerId
				nodeCpy := node
				execution = append(execution, func() {
					nodeCpy.Units[playerIdCpy] = newSum
				})
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
	self.executeTransits(c)
	self.executeOrders(orderMap)
	self.executeGrowth(c)
	self.executeConflicts()
	winner = self.onlyPlayerLeft(c)
	return
}

/*
RandomState creates a random state for the provided players.
*/
func RandomState(c common.Logger, players []PlayerId) (result State) {
	result.Nodes = map[NodeId]*Node{}
	size := common.Norm(len(players)*6, len(players), len(players)*4, len(players)*10)
	allNodes := make([]*Node, 0, size)
	for i := 0; i < size; i++ {
		node := RandomNode()
		result.Nodes[node.Id] = node
		allNodes = append(allNodes, node)
	}
	for _, node := range allNodes {
		node.connect(c, allNodes, result.Nodes)
	}
	perm := rand.Perm(len(allNodes))
	for index, playerId := range players {
		allNodes[perm[index]].Units[playerId] = 10
	}
	return
}
