package linkset

import "testing"

func TestAnd(t *testing.T) {
	var next []string = []string{"a", "d", "c"}
	var current []string = []string{"a", "b", "c"}
	var expected []string = []string{"a", "c"}
	res := And(current, next)
	if !Equal(res, expected) {
		t.Fail()
	}
}
func TestSub(t *testing.T) {
	var current []string = []string{"a", "b", "c"}
	var next []string = []string{"a", "b"}
	var expected []string = []string{"c"}
	res := Sub(current, next)
	if !Equal(res, expected) {
		t.Fail()
	}
}

func TestDuplicate(t *testing.T) {
	var current []string = []string{"a", "b", "c", "c"}
	var next []string = []string{"a", "b"}
	var expected []string = []string{"c"}
	res := Sub(current, next)
	if !Equal(res, expected) {
		t.Log(res)
		t.Fail()
	}

}
