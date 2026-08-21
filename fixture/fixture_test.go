package fixture

import "testing"

func TestReady(t *testing.T) {
	if !Ready() {
		t.Fatal("fixture should be ready")
	}
}
