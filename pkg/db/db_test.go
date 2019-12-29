package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
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
func getMockdB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	dbm, mock, err := sqlmock.New()

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	return dbm, mock
}
func setupPlayer(mock sqlmock.Sqlmock) {
	cols := []string{"id", "name", "team", "active"}

	mock.ExpectQuery("SELECT id,name,team,active FROM player").WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, "player1", "pht", 1).
			AddRow(2, "player2", "pht", 1))
}
func TestGetPlayer(t *testing.T) {
	// dbm, mock, err := sqlmock.New()

	// if err != nil {
	// 	t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	// }
	dbm, mock := getMockdB(t)
	defer dbm.Close()

	//cols := []string{"pid", "name", "team", "active_col4"}
	//var cols [4]string
	cols := []string{"", "", "team", "active_col4"}

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

func TestGetSummary(t *testing.T) {
	dbm, mock := getMockdB(t)
	defer dbm.Close()
	sdb := &SDB{dbm}
	cols := []string{"player_id", "cnt"}

	actual := "select player_id,count() as cnt from innings where how_out not like 'dnb' group by player_id"
	expect := "select player_id,count() as cnt from innings where how_out not like 'dnb' group by player_id"
	re, err := regexp.Compile(expect)
	if err != nil {
		fmt.Printf("**** err %s \n", err)
	}
	if !re.MatchString(actual) {
		fmt.Printf(`**** could not match actual sql: "%s" with expected regexp "%s"`, actual, re.String())
	}

	setupPlayer(mock)
	mock.ExpectQuery("select player_id,count(1) as cnt from innings where how_out not in ('dnb') group by player_id").
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 5).
			AddRow(2, 1))
	mock.ExpectQuery("select player_id,sum(runs_scored) from innings where how_out not in ('dnb') group by player_id").
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 25).
			AddRow(2, 10))
	mock.ExpectQuery("select player_id,count(*) from innings where how_out like 'no' group by player_id").
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 0).
			AddRow(2, 1))

	mock.ExpectQuery("select player_id,max(runs_scored) from innings group by player_id").
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 15).
			AddRow(2, 8))

	mock.ExpectQuery("select player_id, sum(overs_bowled) from bowl_innings group by player_id").
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 5).
			AddRow(2, 28.1))
	mock.ExpectQuery("select player_id, sum(maiden) from bowl_innings group by player_id").
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 0).
			AddRow(2, 2))
	mock.ExpectQuery("select player_id, sum(runs) from bowl_innings group by player_id").
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 117).
			AddRow(2, 402))
	mock.ExpectQuery("select player_id, sum(wickets) from bowl_innings group by player_id").
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 0).
			AddRow(2, 23))
	mock.ExpectQuery("select fielder_id, count(fielder_id) from innings where fielder_id > 0 group by fielder_id").
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 4).
			AddRow(2, 3))
	//
	mock.ExpectQuery("select player_id, count(*) from dropped_catches group by player_id").
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 0).
			AddRow(2, 2))
	mock.ExpectQuery("select player_id, max(wickets) from bowl_innings group by player_id order by max(wickets) desc").
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 2).
			AddRow(2, 3))
	mock.ExpectQuery("select min(runs) from bowl_innings where player_id = 1 and wickets = 2").
		WithArgs().
		WillReturnRows(mock.NewRows([]string{"min"}).
			AddRow(8))
	mock.ExpectQuery("select min(runs) from bowl_innings where player_id = 2 and wickets = 3").
		WithArgs().
		WillReturnRows(mock.NewRows([]string{"min"}).
			AddRow(8))
	rs := sdb.GetSummary()
	v := rs[""]
	compare(t, v.InningsPlayed, 1)

}
