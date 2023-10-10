package space

import (
	"testing"
)

func TestLineOfSight(t *testing.T) {

}

func TestDistance(t *testing.T) {

	P1 := newVector(0, 0, 1)
	P2 := newVector(0, 0, 0)
	D1 := P1.Distance(P2)
	if D1 != 1 {
		t.Errorf("Distance should be 1, got %f", D1)
	}
	D2 := P2.Distance(P1)
	if D2 != 1 {
		t.Errorf("Distance should be 1, got %f", D2)
	}
	P1 = newVector(0, 1, 0)
	P2 = newVector(0, 0, 0)
	D3 := P1.Distance(P2)
	if D3 != 1 {
		t.Errorf("Distance should be 1, got %f", D3)
	}
}

func TestReachable(t *testing.T) {
	P1 := newVector(0, 0, 1)
	P2 := newVector(0, 0, 0)

	if !Reachable(P2, P1, 1.1) {
		t.Errorf("Reachable should be true, got false")
	}
}

// func TestStrinConcat(t *testing.T) {
// 	// var i int = 1
// 	// log.Log("Sat" + i)
// }

func TestVector11Difference(t *testing.T) {
	A := Vector3{
		1, 1, 1,
	}
	B := Vector3{
		1, 1, 1,
	}
	E := Vector3{
		0, 0, 0,
	}
	C := A.Sub(B)
	if C != E {
		t.Log(C, E)
		t.Fail()
	}
}

func TestVectorIdentity(t *testing.T) {
	A := Vector3{
		0, 0, 0,
	}
	B := Vector3{
		1, 1, 1,
	}
	E := Vector3{
		-1, -1, -1,
	}
	C := A.Sub(B)
	if C != E {
		t.Log(C, E)
		t.Fail()
	}
}

func TestVectorOrder(t *testing.T) {
	A := Vector3{
		1, 1, 1,
	}
	B := Vector3{
		0, 0, 0,
	}
	E := Vector3{
		1, 1, 1,
	}
	C := A.Sub(B)
	if C != E {
		t.Log(C, E)
		t.Fail()
	}
}

func TestDopplerShift(t *testing.T) {
	A := Vector3{
		0, 12, 0,
	}
	B := Vector3{
		0, 0, 0,
	}
	shift := A.DopplerShift(B)
	t.Log(shift)
	if shift > 40e3 {
		t.Fail()
	}
}
