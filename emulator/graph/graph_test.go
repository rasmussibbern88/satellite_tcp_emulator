package graph

import "testing"

func TestInstantiateGraph(t *testing.T) {
	InstantiateGraph(5)
}

func TestAddBothCost(t *testing.T) {
	g := InstantiateGraph(5)
	err := AddBothCost(g, 5, 1, 2, 4)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
}

// func TestAddBothCostNoGraph(t *testing.T) {
// 	err := AddBothCost(1, 7, 4)
// 	if err != nil {
// 		t.Log(err.Error())
// 	}
// }

func TestAddBothCostOutOfRange(t *testing.T) {
	g := InstantiateGraph(5)
	err := AddBothCost(g, 5, 1, 7, 4)
	if err != nil {
		t.Log(err.Error())
	}
}

func TestShortestPath(t *testing.T) {
	g := InstantiateGraph(5)
	_, _, err := GetShortestPath(g, 5, 1, 2)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
}

// func TestShortestPathNoGraph(t *testing.T) {
// 	_, _, err := GetShortestPath(1, 2)
// 	if err != nil {
// 		t.Log(err.Error())
// 	}
// }

func TestShortestPathOutOfRange(t *testing.T) {
	g := InstantiateGraph(5)
	_, _, err := GetShortestPath(g, 5, 1, 7)
	if err != nil {
		t.Log(err.Error())
	}
}

func TestLastinArray(t *testing.T) {
	var a []int = []int{1, 2, 3}
	if a[len(a)-1] != 3 {
		t.Fail()
	}
}
