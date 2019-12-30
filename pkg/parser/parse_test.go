package parser

import (
	"os"
	"reflect"
	"testing"
)

type testcase struct {
	name     string
	expected interface{}
	actual   interface{}
}

func compare(t *testing.T, ts testcase) {
	t.Logf("\nWant : %+v \n got : %+v\n", ts.expected, ts.actual)
	if !reflect.DeepEqual(ts.expected, ts.actual) {
		t.Errorf("[%s]Didn't get expected result", ts.name)
	}
}

func TestReadLine(t *testing.T) {

	gm := ReadLine("20110101.csv")
	gm1 := ReadLine("20110102.csv")

	cases := []testcase{
		{"team name1", "phantom", gm.Team1Name},
		{"team name2", "tesla", gm.Team2Name},
		{"team name1", "team1", gm1.Team1Name},
		{"team name2", "team2", gm1.Team2Name}}
	gm1.GenHTML(gm1.GameDate)
	_, err := os.Stat("20110102.html")
	tf := (err == nil)
	c := testcase{"score card html", true, tf}
	cases = append(cases, c)
	for _, c := range cases {
		compare(t, c)
	}
}
