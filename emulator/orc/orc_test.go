package orc

import "testing"

func TestCreateFile(t *testing.T) {
	Name := []string{"Rose", "Smith", "William", "James", "Rolf"}
	Age := []int{28, 24, 29, 31, 21}
	Country := []string{"U.K.", "U.S.", "France", "Norway", "Denmark"}
	writeFile("test", Name, Age, Country)
}
