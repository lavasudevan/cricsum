package db

import (
	"database/sql"
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
func getMockdB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	dbm, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

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
	dbm, mock := getMockdB(t)
	defer dbm.Close()

	cols := []string{"pid", "name", "team", "active_col4"}

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
	sdb := new(SDB)
	sdb.DB = dbm
	cols := []string{"player_id", "cnt"}
	eq := []string{
		`select player_id,count() as cnt from innings where id in 
		(select innings1_id from game where date like '2019%' union select innings2_id from game where date like '2019%') 
		and how_out not in ('dnb') group by player_id`,
		`select player_id,sum(runs_scored) from innings where id in
		(select innings1_id from game where date like '2019%' union select innings2_id from game where date like '2019%') 
		and how_out not in ('dnb') group by player_id`,
		`select player_id,count(*) from innings where id in
		(select innings1_id from game where date like '2019%' union select innings2_id from game where date like '2019%') 
		and how_out like 'no' group by player_id`,
		`select player_id,max(runs_scored) from innings where id in
		(select innings1_id from game where date like '2019%' union select innings2_id from game where date like '2019%') 
		group by player_id`,
		`select player_id, sum(overs_bowled) from bowl_innings where id in
		(select innings1_id from game where date like '2019%' union select innings2_id from game where date like '2019%') 
		group by player_id`,
		`select player_id, sum(maiden) from bowl_innings where id in
		(select innings1_id from game where date like '2019%' union select innings2_id from game where date like '2019%') 
		group by player_id`,
		`select player_id, sum(runs) from bowl_innings where id in
		(select innings1_id from game where date like '2019%' union select innings2_id from game where date like '2019%') 
		group by player_id`,
		`select player_id, sum(wickets) from bowl_innings where id in
		(select innings1_id from game where date like '2019%' union select innings2_id from game where date like '2019%') 
		group by player_id`,
		`select fielder_id, count(fielder_id) from innings where fielder_id > 0 and id in
		(select innings1_id from game where date like '2019%' union select innings2_id from game where date like '2019%') 
		group by fielder_id`,
		`select player_id, count(*) from dropped_catches where innings_id in
		(select innings1_id from game where date like '2019%' union select innings2_id from game where date like '2019%') 
		group by player_id`,
		`select player_id, max(wickets) from bowl_innings where id in
		(select innings1_id from game where date like '2019%' union select innings2_id from game where date like '2019%') 
		group by player_id order by max(wickets) desc`,
		`select min(runs) from bowl_innings where id in 
		(select innings1_id from game where date like '2019%' union select innings2_id from game where date like '2019%') 
		and player_id = 1 and wickets = 2`,
		`select min(runs) from bowl_innings where id in
		(select innings1_id from game where date like '2019%' union select innings2_id from game where date like '2019%') 
		and player_id = 2 and wickets = 3`,

		`select player_id,sum(balls_faced) from innings where id in
		(select innings1_id from game where date like '2019%' union select innings2_id from game where date like '2019%')
		and how_out not in ('dnb') group by player_id`,

		`select player_id,sum(fours_count) from innings where id in
		(select innings1_id from game where date like '2019%' union select innings2_id from game where date like '2019%')
		and how_out not in ('dnb') group by player_id`,

		`select player_id,sum(sixes_count) from innings where id in
		(select innings1_id from game where date like '2019%' union select innings2_id from game where date like '2019%')
		and how_out not in ('dnb') group by player_id`,
	}

	setupPlayer(mock)
	mock.ExpectQuery(eq[0]).
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 5).
			AddRow(2, 1))
	mock.ExpectQuery(eq[1]).
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 25).
			AddRow(2, 10))
	mock.ExpectQuery(eq[2]).
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 0).
			AddRow(2, 1))

	mock.ExpectQuery(eq[3]).
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 15).
			AddRow(2, 8))

	mock.ExpectQuery(eq[4]).
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 5).
			AddRow(2, 28.1))
	mock.ExpectQuery(eq[5]).
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 0).
			AddRow(2, 2))
	mock.ExpectQuery(eq[6]).
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 117).
			AddRow(2, 402))
	mock.ExpectQuery(eq[7]).
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 0).
			AddRow(2, 23))
	mock.ExpectQuery(eq[8]).
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 4).
			AddRow(2, 3))
	mock.ExpectQuery(eq[9]).
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 0).
			AddRow(2, 2))
	mock.ExpectQuery(eq[10]).
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 2).
			AddRow(2, 3))
	mock.ExpectQuery(eq[11]).
		WithArgs().
		WillReturnRows(mock.NewRows([]string{"min"}).
			AddRow(8))
	mock.ExpectQuery(eq[12]).
		WithArgs().
		WillReturnRows(mock.NewRows([]string{"min"}).
			AddRow(8))
	mock.ExpectQuery(eq[13]).
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 50).
			AddRow(2, 5))
	mock.ExpectQuery(eq[14]).
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 1).
			AddRow(2, 1))
	mock.ExpectQuery(eq[15]).
		WithArgs().
		WillReturnRows(mock.NewRows(cols).
			AddRow(1, 0).
			AddRow(2, 1))

	rs := sdb.GetSummary("2019")
	v := rs["player2/pht"]
	compare(t, v.InningsPlayed, 1)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
