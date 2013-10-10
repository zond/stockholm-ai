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

type Edge struct {
	Src   NodeId
	Dst   NodeId
	Units []map[PlayerId]int
}

func (self *Edge) init() {
	for index, _ := range self.Units {
		self.Units[index] = make(map[PlayerId]int)
	}
}

type Node struct {
	Id    NodeId
	Size  int
	Units map[PlayerId]int
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

func RandomNode() (result *Node) {
	result = &Node{
		Size:  common.Norm(50, 25, 10, 100),
		Id:    NodeId(common.RandomString(16)),
		Units: make(map[PlayerId]int),
		Edges: make(map[NodeId]Edge),
	}
	return
}

type Order struct {
	Src   NodeId
	Dst   NodeId
	Units int
}

type Orders []Order

type State struct {
	Nodes map[NodeId]*Node
}

func (self *State) executeTransits() {
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
	for _, node := range self.Nodes {
		if len(node.Units) == 1 {
			players := make([]PlayerId, 0, len(node.Units))
			for playerId, _ := range node.Units {
				players = append(players, playerId)
			}
			playerId := players[0]
			units := node.Units[playerId]
			if units < node.Size {
				node.Units[playerId] = common.Min(node.Size, int(1+float64(units)*(1.0+(growthFactor*(float64(node.Size-units)/float64(node.Size))))))
			} else if units > node.Size {
				node.Units[playerId] = common.Max(0, int(float64(units)/(1.0+(starvationFactor*(float64(units)/float64(node.Size))))-1))
			}
		}
	}
}

func (self *State) executeConflicts() {
}

func (self *State) Next(c common.Logger, orderMap map[PlayerId]Orders) {
	self.executeTransits()
	self.executeOrders(orderMap)
	self.executeGrowth(c)
	self.executeConflicts()
}

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
