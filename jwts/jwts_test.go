package jwts

import "testing"

func TestJWT(t *testing.T) {
	j := &JWT{}

	if j == nil {
		t.Error("variable not initialized")
	}
}
