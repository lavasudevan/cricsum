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

type Details struct {
	Name         string
	Id           int
	Date         string
	RunsScored   int
	HowOut       string
	OversBowled  float32
	RunsConceded int
	Maiden       int
	Wickets      int
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
	key := ""
	for rows.Next() {
		err = rows.Scan(&uid, &playername, &teamname)
		checkErr(err)
		key = playername + "/" + teamname
		nameMap[key] = uid
		idMap[uid] = key

		//fmt.Printf("%d %s", uid, playername)
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
func insertGame(tx *sql.Tx, gamedate string, inn1id int, inn2id int) {
	stmt := fmt.Sprintf("insert into game (date, innings1_id, innings2_id) values ('%s',%d,%d) ", gamedate, inn1id, inn2id)
	_, err := tx.Exec(stmt)
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

func verifyDuplicateBowlers(bowl map[int]parser.Bowling) []string {
	nameMap := make(map[string]string)
	var dupNames []string

	for _, v := range bowl {
		_, ok := nameMap[v.Name]
		if ok == true {
			dupNames = append(dupNames, v.Name)
		} else {
			nameMap[v.Name] = v.Name

		}
	}
	return dupNames
}
func verifyPlayerExist(inn parser.Innings, fieldingTeam string) []string {
	var missingPlayers []string
	var key string
	for _, v := range inn.Bat {
		key = v.Name + "/" + inn.TeamName
		_, e := mPlayerByName[key]
		if e == false && len(v.Name) > 0 {
			missingPlayers = append(missingPlayers, key)
		}
		key = v.FielerName + "/" + fieldingTeam
		_, e = mPlayerByName[key]
		if e == false && len(v.FielerName) > 0 {
			missingPlayers = append(missingPlayers, key)
		}
		key = v.BowlerName + "/" + fieldingTeam
		_, e = mPlayerByName[key]
		if e == false && len(v.BowlerName) > 0 {
			missingPlayers = append(missingPlayers, key)
		}
	}

	for _, v := range inn.Bowl {
		key = v.Name + "/" + fieldingTeam
		_, e := mPlayerByName[key]
		if e == false && len(v.Name) > 0 {
			missingPlayers = append(missingPlayers, key)
		}
	}
	return missingPlayers
}
func insertInnings(tx *sql.Tx, innid int, inn parser.Innings, bowlingTeamName string) {
	for _, v := range inn.Bat {
		plid := mPlayerByName[v.Name+"/"+inn.TeamName]
		flid := mPlayerByName[v.FielerName+"/"+bowlingTeamName]
		blid := mPlayerByName[v.BowlerName+"/"+bowlingTeamName]
		stmt := fmt.Sprintf("insert into innings (id,player_id, runs_scored,fielder_id, how_out,bowler_id) values (%d,%d,%d,%d,'%s',%d)",
			innid, plid, v.RunsScored, flid, v.HowOut, blid)
		fmt.Println(stmt)
		_, err := tx.Exec(stmt)
		checkErr(err)
		//fmt.Printf("key %d batsnmen name %s id %d \n", id, v.Name, plid)
	}
	for _, v := range inn.Bowl {
		plid := mPlayerByName[v.Name+"/"+bowlingTeamName]
		stmt := fmt.Sprintf("insert into bowl_innings (id,player_id, overs_bowled, maiden, runs, wickets ) values (%d,%d,%f,%d,%d,%d)",
			innid, plid, v.OversBowled, v.Maiden, v.RunsConceded, v.Wickets)
		fmt.Println(stmt)
		_, err := tx.Exec(stmt)
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
func makeKey(playerid int, date string) string {
	return fmt.Sprintf("%d/%s", playerid, date)
}
func GetDetails() map[string]Details {
	openDb()
	refreshPlayer()
	var inn1id, inn2id int
	var date string
	rows, err := dba.Query("select date,innings1_id,innings2_id from game ")
	checkErr(err)
	dateMap := make(map[int]string)
	for rows.Next() {
		err = rows.Scan(&date, &inn1id, &inn2id)
		checkErr(err)
		dateMap[inn1id] = date
		dateMap[inn2id] = date
	}

	var id, pid, runs_scored int
	var how_out string
	rows, err = dba.Query("select id,player_id,runs_scored,how_out from innings ")
	checkErr(err)
	detmap := make(map[string]Details)
	var ky string
	for rows.Next() {
		err = rows.Scan(&id, &pid, &runs_scored, &how_out)
		checkErr(err)
		var det Details
		ky = makeKey(pid, dateMap[id])
		det.Id = id
		det.RunsScored = runs_scored
		det.HowOut = how_out
		det.Id = pid
		det.Date = dateMap[id]
		det.Name = mPlayerById[pid]
		detmap[ky] = det
	}

	var overs_bowled float32
	var maiden, runs, wickets int
	rows, err = dba.Query("select id,player_id,overs_bowled,maiden,runs,wickets from bowl_innings")
	checkErr(err)
	for rows.Next() {
		err = rows.Scan(&id, &pid, &overs_bowled, &maiden, &runs, &wickets)
		checkErr(err)
		var det Details
		ky = makeKey(pid, dateMap[id])
		det = detmap[ky]
		det.OversBowled = overs_bowled
		det.RunsConceded = runs
		det.Maiden = maiden
		det.Wickets = wickets

		if len(det.Name) == 0 {
			det.Name = mPlayerById[pid]
			det.Date = dateMap[id]
			det.HowOut = "dnb"
		}
		detmap[ky] = det
	}
	return detmap
}

//
func UpdateGame(game parser.Game) int {
	openDb()

	if checkGameExist(game.GameDate) != 0 {
		fmt.Println("*** game already exist")
		return 0
	}
	refreshPlayer()
	mp := verifyPlayerExist(game.Team1, game.Team2.TeamName)
	er := 0
	if len(mp) > 0 {
		fmt.Printf("*** missing players %s\n", mp)
		er = er + 1
	}
	mp = verifyPlayerExist(game.Team2, game.Team1.TeamName)
	if len(mp) > 0 {
		fmt.Printf("*** missing players %s\n", mp)
		er = er + 1
	}
	dupNames := verifyDuplicateBowlers(game.Team2.Bowl)
	if len(dupNames) > 0 {
		fmt.Printf("*** duplicate bowlers in %s %s\n", game.Team1.TeamName, dupNames)
		er = er + 1
	}
	dupNames = verifyDuplicateBowlers(game.Team1.Bowl)
	if len(dupNames) > 0 {
		fmt.Printf("*** duplicate bowlers in %s %s\n", game.Team2.TeamName, dupNames)
		er = er + 1
	}

	if er > 0 {
		return 0
	}
	//vln
	//return 0

	tx, err := dba.Begin()
	defer tx.Rollback()
	checkErr(err)

	inn1id := getId("bat") + 1
	//fmt.Printf(" bat id %d\n", innid)
	inn2id := inn1id + 1
	//bowlInnId := getId("bowl") + 1
	//fmt.Printf(" bowl id %d\n", innid)
	fmt.Printf(" inn1 id %d inn2 id %d ", inn1id, inn2id)
	insertInnings(tx, inn1id, game.Team1, game.Team2.TeamName)
	insertInnings(tx, inn2id, game.Team2, game.Team1.TeamName)
	insertGame(tx, game.GameDate, inn1id, inn2id)

	tx.Commit()

	return 0
}
