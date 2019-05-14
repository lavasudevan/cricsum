package db

import (
	"fmt"
	"testing"
)

func TestGetPlayer(t *testing.T) {
	rvn, _ := GetPlayer()
	fmt.Println(len(rvn))
	if len(rvn) == 0 {
		t.Error("Expected 2, got ")
	}
}
