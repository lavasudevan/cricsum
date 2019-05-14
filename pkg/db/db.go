package db

import (
	"database/sql"
	"fmt"

	"github.com/cricsum/pkg/parser"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DBNAME = "scores.db"
)

var dba *sql.DB
var mPlayerByName map[string]int
var mPlayerById map[int]string

func openDb() {
	var err error
	if dba == nil {
		dba, err = sql.Open("sqlite3", "./"+DBNAME)
		checkErr(err)
	}
}
func closeDB() {
	dba.Close()
}

func refreshPlayer() {
	if len(mPlayerByName) == 0 {
		mPlayerByName, mPlayerById = GetPlayer()
	}
}
func getId(inningsType string) int {
	openDb()

	//	stmt := "select id FROM innings_index where type like '$1'"
	var stmt string
	if inningsType == "bat" {
		stmt = "select max(id) FROM innings "
	}
	if inningsType == "bowl" {
		stmt = "select max(id) FROM bowl_innings"
	}

	row := dba.QueryRow(stmt)
	var id int
	switch err := row.Scan(&id); err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return 0
	case nil:
		return id
	default:
		return 0
	}
}

type Summary struct {
	Name          string
	InningsPlayed int
	NotOut        int
	RunsScored    int
	Average       float32
	Highest       int
	Dismissal     int
	OversBowled   float32
	RunsConceded  int
	Maiden        int
	Wickets       int
	RunsPerOver   float32
}

//
func GetPlayer() (map[string]int, map[int]string) {
	openDb()
	//	db, err := sql.Open("sqlite3", "./"+DBNAME)//
	//	checkErr(err)
	rows, err := dba.Query("SELECT id,name,team FROM player")
	checkErr(err)
	var uid int
	var playername string
	var teamname string
	nameMap := make(map[string]int)
	idMap := make(map[int]string)
	for rows.Next() {
		err = rows.Scan(&uid, &playername, &teamname)
		checkErr(err)
		nameMap[playername] = uid
		idMap[uid] = playername

		fmt.Printf("%d %s", uid, playername)
	}

	rows.Close() //good habit to close

	//db.Close()
	return nameMap, idMap
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func updateIndex(itype string, id int) {
	//select id FROM innings_index where type like '$1'"
	stmt := fmt.Sprintf("update innings_index set id = %d where type like '%s'", id, itype)
	_, err := dba.Exec(stmt)
	fmt.Println(stmt)
	checkErr(err)
}
func insertGame(gamedate string, inn1id int, inn2id int) {
	stmt := fmt.Sprintf("insert into game (date, innings1_id, innings2_id) values ('%s',%d,%d) ", gamedate, inn1id, inn2id)
	_, err := dba.Exec(stmt)
	fmt.Println(stmt)
	checkErr(err)
}
func checkGameExist(gameDate string) int {
	stmt := fmt.Sprintf("select count(*) from game where date like '%s'", gameDate)

	fmt.Println(stmt)
	row := dba.QueryRow(stmt)
	var cnt int
	switch err := row.Scan(&cnt); err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return 0
	case nil:
		return cnt
	default:
		return -1
	}
}

func verifyPlayerExist(inn parser.Innings) []string {
	var missingPlayers []string
	for _, v := range inn.Bat {
		_, e := mPlayerByName[v.Name]
		if e == false && len(v.Name) > 0 {
			missingPlayers = append(missingPlayers, v.Name)
		}
		_, e = mPlayerByName[v.FielerName]
		if e == false && len(v.FielerName) > 0 {
			missingPlayers = append(missingPlayers, v.FielerName)
		}
		_, e = mPlayerByName[v.BowlerName]
		if e == false && len(v.BowlerName) > 0 {
			missingPlayers = append(missingPlayers, v.BowlerName)
		}
	}

	for _, v := range inn.Bowl {
		_, e := mPlayerByName[v.Name]
		if e == false && len(v.Name) > 0 {
			missingPlayers = append(missingPlayers, v.Name)
		}
	}
	return missingPlayers
}
func insertInnings(innid int, inn parser.Innings) {
	for _, v := range inn.Bat {
		plid := mPlayerByName[v.Name]
		flid := mPlayerByName[v.FielerName]
		blid := mPlayerByName[v.BowlerName]
		stmt := fmt.Sprintf("insert into innings (id,player_id, runs_scored,fielder_id, how_out,bowler_id) values (%d,%d,%d,%d,'%s',%d)",
			innid, plid, v.RunsScored, flid, v.HowOut, blid)
		fmt.Println(stmt)
		_, err := dba.Exec(stmt)
		checkErr(err)
		//fmt.Printf("key %d batsnmen name %s id %d \n", id, v.Name, plid)
	}
	for _, v := range inn.Bowl {
		plid := mPlayerByName[v.Name]
		stmt := fmt.Sprintf("insert into bowl_innings (id,player_id, overs_bowled, maiden, runs, wickets ) values (%d,%d,%f,%d,%d,%d)",
			innid, plid, v.OversBowled, v.Maiden, v.RunsConceded, v.Wickets)
		fmt.Println(stmt)
		_, err := dba.Exec(stmt)
		checkErr(err)
		//fmt.Printf("key %d batsnmen name %s id %d \n", id, v.Name, plid)
	}
}

func GetSummary() map[string]Summary {
	refreshPlayer()
	/*
			type Summary struct {
			Name          string
			InningsPlayed int
			NotOut        int
			RunsScored    int
			Average       float32
			Highest       int
			OversBowled   float32
			RunsConceced  int
			Maiden        int
			Wickets       int
			RunsPerOver   float32
		}
		select player_id,count(*) from innings where how_out not in ('dnb') group by player_id
		select player_id,sum(runs_scored) from innings where how_out not in ('dnb') group by player_id
		select player_id,count(*) from innings where how_out like 'no' group by player_id
		select player_id,max(runs_scored) from innings group by player_id

		select player_id, sum(overs_bowled) from bowl_innings group by player_id
		select player_id, sum(maiden) from bowl_innings group by player_id
		select player_id, sum(runs) from bowl_innings group by player_id
		select player_id, sum(wickets) from bowl_innings group by player_id
		select fielder_id, count(fielder_id) from innings where fielder_id > 0 group by fielder_id;
	*/

	openDb()

	rows, err := dba.Query("select player_id,count(*) from innings where how_out not in ('dnb') group by player_id")
	checkErr(err)
	var pid, cnt int
	idMap := make(map[int]Summary)

	for rows.Next() {
		err = rows.Scan(&pid, &cnt)
		checkErr(err)
		var sm Summary
		sm.InningsPlayed = cnt
		idMap[pid] = sm
	}

	rows, err = dba.Query("select player_id,sum(runs_scored) from innings where how_out not in ('dnb') group by player_id")
	checkErr(err)
	var sumruns int
	for rows.Next() {
		err = rows.Scan(&pid, &sumruns)
		checkErr(err)
		sm := idMap[pid]
		sm.RunsScored = sumruns
		idMap[pid] = sm
	}

	rows, err = dba.Query("select player_id,count(*) from innings where how_out like 'no' group by player_id")
	checkErr(err)
	var cntno int
	for rows.Next() {
		err = rows.Scan(&pid, &cntno)
		checkErr(err)
		sm := idMap[pid]
		sm.NotOut = cntno
		idMap[pid] = sm
	}

	rows, err = dba.Query("select player_id,max(runs_scored) from innings group by player_id")
	checkErr(err)
	var maxruns int
	for rows.Next() {
		err = rows.Scan(&pid, &maxruns)
		checkErr(err)
		sm := idMap[pid]
		sm.Highest = maxruns
		idMap[pid] = sm
	}

	rows, err = dba.Query("select player_id, sum(overs_bowled) from bowl_innings group by player_id")
	checkErr(err)
	var sumovers float32
	for rows.Next() {
		err = rows.Scan(&pid, &sumovers)
		checkErr(err)
		sm := idMap[pid]
		sm.OversBowled = sumovers
		idMap[pid] = sm
	}

	rows, err = dba.Query("select player_id, sum(maiden) from bowl_innings group by player_id")
	checkErr(err)
	var summaiden int
	for rows.Next() {
		err = rows.Scan(&pid, &summaiden)
		checkErr(err)
		sm := idMap[pid]
		sm.Maiden = summaiden
		idMap[pid] = sm
	}

	rows, err = dba.Query("select player_id, sum(runs) from bowl_innings group by player_id")
	checkErr(err)
	for rows.Next() {
		err = rows.Scan(&pid, &sumruns)
		checkErr(err)
		sm := idMap[pid]
		sm.RunsConceded = sumruns
		idMap[pid] = sm
	}

	rows, err = dba.Query("select player_id, sum(wickets) from bowl_innings group by player_id")
	checkErr(err)
	var sumwickets int
	for rows.Next() {
		err = rows.Scan(&pid, &sumwickets)
		checkErr(err)
		sm := idMap[pid]
		sm.Wickets = sumwickets
		idMap[pid] = sm
	}

	rows, err = dba.Query("select fielder_id, count(fielder_id) from innings where fielder_id > 0 group by fielder_id;")
	checkErr(err)
	var cntfeilder int
	for rows.Next() {
		err = rows.Scan(&pid, &cntfeilder)
		checkErr(err)
		sm := idMap[pid]
		sm.Dismissal = cntfeilder
		idMap[pid] = sm
	}

	rv := make(map[string]Summary)
	for id, v := range idMap {
		var denom float32
		denom = (float32)(v.InningsPlayed - v.NotOut)
		if denom == 0 {
			denom = 1.0
		}
		v.Average = ((float32)(v.RunsScored)) / denom
		ico := ((int)(v.OversBowled * 10.0))
		compOvers := ((int)(v.OversBowled)) * 10
		remBalls := ico - compOvers
		noofballs := ((compOvers / 10) * 6) + remBalls

		if v.OversBowled > 0.0 {
			v.RunsPerOver = ((float32)(v.RunsConceded) / ((float32)(noofballs))) * (6.0)
		}

		rv[mPlayerById[id]] = v
	}
	return rv
}

//
func UpdateGame(game parser.Game) int {
	openDb()

	if checkGameExist(game.GameDate) != 0 {
		fmt.Println("*** game already exist")
		return 0
	}
	refreshPlayer()
	mp := verifyPlayerExist(game.Team1)
	mp1 := verifyPlayerExist(game.Team2)
	er := 0
	if len(mp) > 0 {
		fmt.Printf("*** missing players %s", mp)
		er = er + 1
	}
	if len(mp1) > 0 {
		fmt.Printf("*** missing players %s", mp1)
		er = er + 1
	}
	if er > 0 {
		return 0
	}

	tx, err := dba.Begin()
	checkErr(err)

	//refreshPlayer()
	innid := getId("bat") + 1
	fmt.Printf(" bat id %d\n", innid)
	bowlInnId := getId("bowl") + 1
	fmt.Printf(" bowl id %d\n", innid)
	fmt.Printf(" %d %d ", innid, bowlInnId)
	insertInnings(innid, game.Team1)
	insertInnings(bowlInnId, game.Team2)
	//updateIndex("bat", innid)
	//updateIndex("bowl", bowlInnId)
	insertGame(game.GameDate, innid, bowlInnId)

	tx.Commit()

	return 0
}
