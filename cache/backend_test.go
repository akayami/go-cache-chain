package cache

import "testing"

func TestBacked(t *testing.T) {

	b := Backend{name: "test"}

	if b.GetName() != "test" {
		t.Errorf("Expect the name to be test")
	}

}
