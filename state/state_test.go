package state

import (
	"reflect"
	"testing"
)

var a = NodeId("a")
var b = NodeId("b")
var c = NodeId("c")
var d = NodeId("d")
var e = NodeId("e")
var f = NodeId("f")
var g = NodeId("g")
var h = NodeId("h")

/*
         f
        /|
 ----g / |
 |   |/  |
 a - b - c
  \     /
	 \   / |
	  \ /
		 d---e h

*/
func testState() (result *State) {
	result = NewState()
	result.Add(NewNode(a, 100)).Add(NewNode(b, 100)).Add(NewNode(c, 100)).Add(NewNode(d, 100)).Add(NewNode(e, 100)).Add(NewNode(f, 100)).Add(NewNode(g, 100)).Add(NewNode(h, 100))
	result.Nodes[a].Connect(result.Nodes[b], 1)
	result.Nodes[a].Connect(result.Nodes[d], 3)
	result.Nodes[b].Connect(result.Nodes[f], 3)
	result.Nodes[b].Connect(result.Nodes[c], 1)
	result.Nodes[c].Connect(result.Nodes[f], 3)
	result.Nodes[c].Connect(result.Nodes[d], 3)
	result.Nodes[c].Connect(result.Nodes[e], 1)
	result.Nodes[d].Connect(result.Nodes[e], 3)
	result.Nodes[a].Connect(result.Nodes[g], 5)
	result.Nodes[b].Connect(result.Nodes[g], 1)
	return
}

func assertPath(t *testing.T, s *State, src, dst NodeId, exp ...NodeId) {
	if found := s.Path(src, dst, nil); !reflect.DeepEqual(found, exp) {
		t.Fatalf("Wanted path from %v to %v to be %+v, but got %+v", src, dst, exp, found)
	}
}

func TestPath(t *testing.T) {
	s := testState()
	assertPath(t, s, a, h)
	assertPath(t, s, a, e, b, b, c, c, e, e)
	assertPath(t, s, a, g, b, b, g, g)
	assertPath(t, s, a, b, b, b)
	assertPath(t, s, a, c, b, b, c, c)
	assertPath(t, s, a, d, d, d, d, d)
	assertPath(t, s, a, d, d, d, d, d)
	assertPath(t, s, a, f, b, b, f, f, f, f)
	assertPath(t, s, f, g, b, b, b, b, g, g)
}
