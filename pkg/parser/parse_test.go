package parser

import "testing"

func TestReadLine(t *testing.T) {
	game := ReadLine("20180320.csv")
	if len(game.Team1.Bat) == 0 {
		t.Error("Expected non 0")
	}
}
