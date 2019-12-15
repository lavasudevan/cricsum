package db

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// func setup() {
// 	type data struct {
// 		id       int
// 		name     string
// 		teamname string
// 		active   int
// 	}
// 	dt := []data{
// 		{1, "p1", "t1", 1},
// 		{2, "p1", "t1", 1},
// 	}

// 	var ar []interface{} = make([]interface{}, len(dt))
// 	for ii, v := range dt {
// 		ar[ii] = v
// 		//fmt.Printf("index %d %v ", ii, v)
// 	}
// }

func compare(t *testing.T, want interface{}, got interface{}) {
	s1 := fmt.Sprintf("%+v", want)
	s2 := fmt.Sprintf("%+v", got)
	fmt.Printf("Want : %s \n got : %s\n", s1, s2)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("returned map is wrong")
	}
}
func TestGetPlayer(t *testing.T) {
	dbm, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer dbm.Close()

	cols := []string{"id", "name", "team", "active"}

	byid := map[int]string{
		1: "player1/pht",
		2: "player2/pht",
	}
	byname := map[string]int{
		"player1/pht": 1,
		"player2/pht": 2,
	}

	sdb := &SDB{dbm}

	mock.ExpectQuery("SELECT id,name,team,active FROM player").WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, "player1", "pht", 1).
			AddRow(2, "player2", "pht", 1))
	actbyname, actbyid := sdb.GetPlayer()

	compare(t, byname, actbyname)
	compare(t, byid, actbyid)
}
