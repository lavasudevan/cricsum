package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/cricsum/pkg/parser"
)

//SDB wrapper for sql.DB
type SDB struct {
	*sql.DB
}

var mPlayerByName map[string]int
var mPlayerByID map[int]string
var mPlayer map[int]Player

func (dba *SDB) refreshPlayer() {
	if len(mPlayerByName) == 0 {
		mPlayerByName, mPlayerByID = dba.GetPlayer()
	}
}
func (dba *SDB) getID(inningsType string) int {
	//openDb()

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

//Player data structure
type Player struct {
	Name   string
	Active int
	Team   string
}

//Summary with stats
type Summary struct {
	Name            string
	PlayerID        int
	InningsPlayed   int
	NotOut          int
	RunsScored      int
	Average         float32
	Highest         int
	Dismissal       int
	DroppedCatches  int
	OversBowled     float32
	RunsConceded    int
	Maiden          int
	Wickets         int
	RunsPerWicket   float32
	RunsPerOver     float32
	BestWickets     int
	BestWicketsRuns int
	BallsFaced      int
	FoursHit        int
	SixesHit        int
}

//Details about each game
type Details struct {
	Name           string
	ID             int
	Date           string
	RunsScored     int
	Dismissal      int
	DroppedCatches int
	HowOut         string
	OversBowled    float32
	RunsConceded   int
	Maiden         int
	Wickets        int
}

//GetPlayer retrieves all the players, creates 2 map. one with id and another with name
func (dba *SDB) GetPlayer() (map[string]int, map[int]string) {
	//openDb()
	//	db, err := sql.Open("sqlite3", "./"+DBNAME)//
	//	checkErr(err)
	rows, err := dba.Query("SELECT id,name,team,active FROM player")
	checkErr(err)
	var uid int
	var playername string
	var teamname string
	var active int

	nameMap := make(map[string]int)
	idMap := make(map[int]string)
	mPlayer = make(map[int]Player)
	key := ""
	for rows.Next() {
		err = rows.Scan(&uid, &playername, &teamname, &active)
		checkErr(err)
		key = playername + "/" + teamname
		nameMap[key] = uid
		idMap[uid] = key
		var p Player
		p.Name = playername
		p.Team = teamname
		p.Active = active
		mPlayer[uid] = p
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

func (dba *SDB) updateIndex(itype string, id int) {
	//select id FROM innings_index where type like '$1'"
	stmt := fmt.Sprintf("update innings_index set id = %d where type like '%s'", id, itype)
	_, err := dba.Exec(stmt)
	fmt.Println(stmt)
	checkErr(err)
}
func insertGame(tx *sql.Tx, gamedate string, inn1id int, inn2id int, wonby string,
	inn1total int, inn1overs float32,
	inn2total int, inn2overs float32,
	team1name string, team2name string) {
	stmt := fmt.Sprintf("insert into game (date, innings1_id, innings2_id,wonby,team1_score,team1_overs,team2_score,team2_overs,team1_name,team2_name) "+
		"values ('%s',%d,%d,'%s',%d,%f,%d,%f,'%s','%s') ",
		gamedate, inn1id, inn2id, wonby, inn1total, inn1overs, inn2total, inn2overs, team1name, team2name)
	_, err := tx.Exec(stmt)
	fmt.Println(stmt)
	checkErr(err)
}
func (dba *SDB) checkGameExist(gameDate string) int {
	stmt := fmt.Sprintf("select count(*) from game where date like '%s'", gameDate)
	//openDb()
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
func verifyDuplicateBatsman(bat map[int]parser.Batting) []string {
	nameMap := make(map[string]string)
	var dupNames []string

	for _, v := range bat {
		_, ok := nameMap[v.Name]
		if ok == true {
			dupNames = append(dupNames, v.Name)
		} else {
			nameMap[v.Name] = v.Name
		}
	}
	return dupNames
}

func verifyRunsScored(inn parser.Innings) int {
	nzc := 0
	for _, v := range inn.Bat {
		if v.RunsScored > 0 {
			nzc = nzc + 1
		}
	}
	return nzc
}
func verifyDismissal(inn parser.Innings) []string {
	var howout []string
	for _, v := range inn.Bat {
		if v.HowOut == "no" || v.HowOut == "b" || v.HowOut == "c" || v.HowOut == "dnb" || v.HowOut == "st" || v.HowOut == "ro" {

		} else {
			howout = append(howout, v.HowOut)
		}
	}
	return howout
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
func verifyFielders(droppedCatches []string, teamname string) []string {
	var missingPlayers []string
	var key string
	for _, plname := range droppedCatches {
		key = plname + "/" + teamname
		_, e := mPlayerByName[key]
		if e == false && len(plname) > 0 {
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
		stmt := fmt.Sprintf("insert into innings (id,player_id, runs_scored,fielder_id, how_out,bowler_id,balls_faced,fours_count,sixes_count) values (%d,%d,%d,%d,'%s',%d,%d,%d,%d)",
			innid, plid, v.RunsScored, flid, v.HowOut, blid, v.BallsFaced, v.FoursHit, v.SixesHit)
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

func insertDroppedCatches(tx *sql.Tx, innid int, teamName string, droppedCatches []string) {
	var key string
	for _, plname := range droppedCatches {
		key = plname + "/" + teamName
		plid := mPlayerByName[key]
		stmt := fmt.Sprintf("insert into dropped_catches (innings_id,player_id ) values (%d,%d)",
			innid, plid)
		fmt.Println(stmt)
		_, err := tx.Exec(stmt)
		checkErr(err)
	}
}

//GetPlayerDetails returns player details with a player id
func GetPlayerDetails(pid int) Player {
	return mPlayer[pid]
}

//IsPlayerActive returns if the player is active or inactive
func IsPlayerActive(pid int) bool {
	p := mPlayer[pid]
	if p.Active == 1 {
		return true
	}
	return false
}

const noOfMatches int = 3

//DisablePlayers disable all inactive players
func (dba *SDB) DisablePlayers(year string) {
	//1 active, 0 - inactive
	rs := dba.GetSummary(year)
	fmt.Println("disabling...")

	tx, err := dba.Begin()
	defer tx.Rollback()
	checkErr(err)

	for name, v := range rs {
		if v.InningsPlayed > noOfMatches {
			continue
		}
		p := GetPlayerDetails(v.PlayerID)
		if p.Active == 1 {
			fmt.Println(name + "/" + v.Name)
			stmt := fmt.Sprintf("update player set active = %d where id = %d ", 0, v.PlayerID)
			_, err = tx.Exec(stmt)
			checkErr(err)
		}
	}
	tx.Commit()
}

func embedIDS(qry, year string) string {
	id1 := "select innings1_id from game where date like '%YEAR%'"
	id2 := "select innings2_id from game where date like '%YEAR%'"
	id1 = strings.Replace(id1, "%YEAR%", year+"%", 1)
	id2 = strings.Replace(id2, "%YEAR%", year+"%", 1)
	ids := "(" + id1 + " union " + id2 + ")"

	//fmt.Println(ids)
	qry = strings.Replace(qry, "%IDS%", ids, 1)
	//fmt.Println(qry)
	return qry
}

//GetSummary returns stats for all the games
func (dba *SDB) GetSummary(year string) map[string]Summary {
	dba.refreshPlayer()
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

	//openDb()

	qry := "select player_id,count() as cnt from innings where id in %IDS% and how_out not in ('dnb') group by player_id"
	rows, err := dba.Query(embedIDS(qry, year))
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

	qry = "select player_id,sum(runs_scored) from innings where id in %IDS% and how_out not in ('dnb') group by player_id"
	rows, err = dba.Query(embedIDS(qry, year))
	checkErr(err)
	var sumruns int
	for rows.Next() {
		err = rows.Scan(&pid, &sumruns)
		checkErr(err)
		sm := idMap[pid]
		sm.RunsScored = sumruns
		idMap[pid] = sm
	}

	qry = "select player_id,count(*) from innings where id in %IDS% and how_out like 'no' group by player_id"
	rows, err = dba.Query(embedIDS(qry, year))
	checkErr(err)
	var cntno int
	for rows.Next() {
		err = rows.Scan(&pid, &cntno)
		checkErr(err)
		sm := idMap[pid]
		sm.NotOut = cntno
		idMap[pid] = sm
	}

	qry = "select player_id,max(runs_scored) from innings where id in %IDS% group by player_id"
	rows, err = dba.Query(embedIDS(qry, year))
	checkErr(err)
	var maxruns int
	for rows.Next() {
		err = rows.Scan(&pid, &maxruns)
		checkErr(err)
		sm := idMap[pid]
		sm.Highest = maxruns
		idMap[pid] = sm
	}

	qry = "select player_id, sum(overs_bowled) from bowl_innings where id in %IDS% group by player_id"
	rows, err = dba.Query(embedIDS(qry, year))
	checkErr(err)
	var sumovers float32
	for rows.Next() {
		err = rows.Scan(&pid, &sumovers)
		checkErr(err)
		sm := idMap[pid]
		sm.OversBowled = sumovers
		idMap[pid] = sm
	}

	qry = "select player_id, sum(maiden) from bowl_innings where id in %IDS% group by player_id"
	rows, err = dba.Query(embedIDS(qry, year))
	checkErr(err)
	var summaiden int
	for rows.Next() {
		err = rows.Scan(&pid, &summaiden)
		checkErr(err)
		sm := idMap[pid]
		sm.Maiden = summaiden
		idMap[pid] = sm
	}

	qry = "select player_id, sum(runs) from bowl_innings where id in %IDS% group by player_id"
	rows, err = dba.Query(embedIDS(qry, year))
	checkErr(err)
	for rows.Next() {
		err = rows.Scan(&pid, &sumruns)
		checkErr(err)
		sm := idMap[pid]
		sm.RunsConceded = sumruns
		idMap[pid] = sm
	}

	qry = "select player_id, sum(wickets) from bowl_innings where id in %IDS% group by player_id"
	rows, err = dba.Query(embedIDS(qry, year))
	checkErr(err)
	var sumwickets int
	for rows.Next() {
		err = rows.Scan(&pid, &sumwickets)
		checkErr(err)
		sm := idMap[pid]
		sm.Wickets = sumwickets
		idMap[pid] = sm
	}

	qry = "select fielder_id, count(fielder_id) from innings where fielder_id > 0 and id in %IDS% group by fielder_id"
	rows, err = dba.Query(embedIDS(qry, year))
	checkErr(err)
	var cntfeilder int
	for rows.Next() {
		err = rows.Scan(&pid, &cntfeilder)
		checkErr(err)
		sm := idMap[pid]
		sm.Dismissal = cntfeilder
		idMap[pid] = sm
	}

	qry = "select player_id, count(*) from dropped_catches where innings_id in %IDS% group by player_id"
	rows, err = dba.Query(embedIDS(qry, year))
	checkErr(err)
	for rows.Next() {
		err = rows.Scan(&pid, &cntfeilder)
		checkErr(err)
		sm := idMap[pid]
		sm.DroppedCatches = cntfeilder
		idMap[pid] = sm
	}

	qry = "select player_id, max(wickets) from bowl_innings where id in %IDS% group by player_id order by max(wickets) desc"
	rows, err = dba.Query(embedIDS(qry, year))
	var maxwickets int
	for rows.Next() {
		err = rows.Scan(&pid, &maxwickets)
		checkErr(err)
		sm := idMap[pid]
		sm.BestWickets = maxwickets
		s1 := fmt.Sprintf("and player_id = %d and wickets = %d", pid, maxwickets)
		stmt := "select min(runs) from bowl_innings where id in %IDS% " + s1
		row := dba.QueryRow(embedIDS(stmt, year))
		var minruns int
		row.Scan(&minruns)
		sm.BestWicketsRuns = minruns

		idMap[pid] = sm
	}

	var ii int
	qry = "select player_id,sum(balls_faced) from innings where id in %IDS% and how_out not in ('dnb') group by player_id"
	rows, err = dba.Query(embedIDS(qry, year))
	checkErr(err)
	for rows.Next() {
		err = rows.Scan(&pid, &ii)
		checkErr(err)
		sm := idMap[pid]
		sm.BallsFaced = ii
		idMap[pid] = sm
	}

	qry = "select player_id,sum(fours_count) from innings where id in %IDS% and how_out not in ('dnb') group by player_id"
	rows, err = dba.Query(embedIDS(qry, year))
	checkErr(err)
	for rows.Next() {
		err = rows.Scan(&pid, &ii)
		checkErr(err)
		sm := idMap[pid]
		sm.FoursHit = ii
		idMap[pid] = sm
	}
	qry = "select player_id,sum(sixes_count) from innings where id in %IDS% and how_out not in ('dnb') group by player_id"
	rows, err = dba.Query(embedIDS(qry, year))
	checkErr(err)
	for rows.Next() {
		err = rows.Scan(&pid, &ii)
		checkErr(err)
		sm := idMap[pid]
		sm.SixesHit = ii
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
		if v.Wickets > 0 {
			v.RunsPerWicket = (float32)(v.RunsConceded) / (float32)(v.Wickets)
		}
		v.PlayerID = id
		rv[mPlayerByID[id]] = v
	}
	return rv
}
func makeKey(playername string, date string) string {
	return fmt.Sprintf("%s/%s", playername, date)
}

//GetDetails returs details of each of the game
func (dba *SDB) GetDetails() map[string]Details {
	//openDb()
	dba.refreshPlayer()
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

	var id, pid, runsScored int
	var howOut string
	rows, err = dba.Query("select id,player_id,runs_scored,how_out from innings ")
	checkErr(err)
	detmap := make(map[string]Details)
	var ky string
	for rows.Next() {
		err = rows.Scan(&id, &pid, &runsScored, &howOut)
		checkErr(err)
		var det Details
		ky = makeKey(mPlayerByID[pid], dateMap[id])
		det.ID = id
		det.RunsScored = runsScored
		det.HowOut = howOut
		det.ID = pid
		det.Date = dateMap[id]
		det.Name = mPlayerByID[pid]
		detmap[ky] = det
	}

	var overBowled float32
	var maiden, runs, wickets int
	rows, err = dba.Query("select id,player_id,overs_bowled,maiden,runs,wickets from bowl_innings")
	checkErr(err)
	for rows.Next() {
		err = rows.Scan(&id, &pid, &overBowled, &maiden, &runs, &wickets)
		checkErr(err)
		var det Details
		ky = makeKey(mPlayerByID[pid], dateMap[id])
		det = detmap[ky]
		det.OversBowled = overBowled
		det.RunsConceded = runs
		det.Maiden = maiden
		det.Wickets = wickets

		if len(det.Name) == 0 {
			det.Name = mPlayerByID[pid]
			det.Date = dateMap[id]
			det.HowOut = "dnb"
		}
		detmap[ky] = det
	}

	var cntDismisal int
	rows, err = dba.Query("select id,fielder_id, count(fielder_id) from innings where fielder_id > 0 group by id,fielder_id")
	checkErr(err)
	for rows.Next() {
		err = rows.Scan(&id, &pid, &cntDismisal)
		checkErr(err)
		var det Details
		ky = makeKey(mPlayerByID[pid], dateMap[id])
		det = detmap[ky]
		//accumulating becase we may have common fielder some times when playing within phantoms
		det.Dismissal += cntDismisal
		if len(det.Name) == 0 {
			det.Name = mPlayerByID[pid]
			det.Date = dateMap[id]
			det.HowOut = "dnb"
		}
		detmap[ky] = det
	}

	var cntCatchDropped int
	rows, err = dba.Query("select innings_id,player_id, count(player_id) from dropped_catches group by innings_id,player_id")
	checkErr(err)
	for rows.Next() {
		err = rows.Scan(&id, &pid, &cntCatchDropped)
		checkErr(err)
		var det Details
		ky = makeKey(mPlayerByID[pid], dateMap[id])
		det = detmap[ky]
		//accumulating becase we may have common fielder some times when playing within phantoms
		det.DroppedCatches += cntCatchDropped
		if len(det.Name) == 0 {
			det.Name = mPlayerByID[pid]
			det.Date = dateMap[id]
			det.HowOut = "dnb"
		}
		detmap[ky] = det
	}

	return detmap
}

func (dba *SDB) getInningsID(gameDate string) (int, int, int) {
	var stmt string
	stmt = fmt.Sprintf("select id,innings1_id,innings2_id FROM game where date like '%s'", gameDate)

	row := dba.QueryRow(stmt)
	var gameid, inn1id, inn2id int

	switch err := row.Scan(&gameid, &inn1id, &inn2id); err {
	case sql.ErrNoRows:
		//		fmt.Println("No rows were returned!")
		return 0, 0, 0
	case nil:
		return gameid, inn1id, inn2id
	default:
		return 0, 0, 0
	}
}

//RemoveGame removes game from DB
func (dba *SDB) RemoveGame(gameDate string) int {
	if dba.checkGameExist(gameDate) <= 0 {
		fmt.Println("*** game doesnt exist")
		return 0
	}
	tx, err := dba.Begin()
	gameid, inn1id, inn2id := dba.getInningsID(gameDate)
	stmt := fmt.Sprintf("delete from dropped_catches where innings_id in (%d,%d)",
		inn1id, inn2id)
	fmt.Println(stmt)
	_, err = tx.Exec(stmt)
	checkErr(err)

	stmt = fmt.Sprintf("delete from innings where id in (%d,%d)",
		inn1id, inn2id)
	fmt.Println(stmt)
	_, err = tx.Exec(stmt)
	checkErr(err)

	stmt = fmt.Sprintf("delete from bowl_innings where id in (%d,%d)",
		inn1id, inn2id)
	fmt.Println(stmt)
	_, err = tx.Exec(stmt)
	checkErr(err)

	stmt = fmt.Sprintf("delete from game where id = %d", gameid)
	fmt.Println(stmt)
	_, err = tx.Exec(stmt)
	checkErr(err)

	defer tx.Rollback()
	tx.Commit()

	return 0
}

//UpdateGame inserts game data in the DB
func (dba *SDB) UpdateGame(game parser.Game) int {
	//openDb()

	if dba.checkGameExist(game.GameDate) != 0 {
		fmt.Println("*** game already exist")
		return 0
	}
	dba.refreshPlayer()
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
	dupNames := verifyDuplicateBatsman(game.Team1.Bat)
	if len(dupNames) > 0 {
		fmt.Printf("*** duplicate batsman in %s %s\n", game.Team1.TeamName, dupNames)
		er = er + 1
	}
	dupNames = verifyDuplicateBatsman(game.Team2.Bat)
	if len(dupNames) > 0 {
		fmt.Printf("*** duplicate batsman in %s %s\n", game.Team1.TeamName, dupNames)
		er = er + 1
	}

	dupNames = verifyDuplicateBowlers(game.Team2.Bowl)
	if len(dupNames) > 0 {
		fmt.Printf("*** duplicate bowlers in %s %s\n", game.Team1.TeamName, dupNames)
		er = er + 1
	}
	dupNames = verifyDuplicateBowlers(game.Team1.Bowl)
	if len(dupNames) > 0 {
		fmt.Printf("*** duplicate bowlers in %s %s\n", game.Team2.TeamName, dupNames)
		er = er + 1
	}
	ho := verifyDismissal(game.Team1)
	if len(ho) > 0 {
		fmt.Printf("*** unknown dismissal type  %s %s\n", game.Team1.TeamName, ho)
		er = er + 1
	}
	if verifyRunsScored(game.Team1) == 0 {
		fmt.Printf("*** all bastman have 0 score team name %s\n", game.Team1.TeamName)
		er = er + 1
	}
	ho = verifyDismissal(game.Team2)
	if len(ho) > 0 {
		fmt.Printf("*** unknown dismissal type  %s %s\n", game.Team2.TeamName, ho)
		er = er + 1
	}
	if verifyRunsScored(game.Team2) == 0 {
		fmt.Printf("*** all bastman have 0 score team name %s\n", game.Team2.TeamName)
		er = er + 1
	}
	if game.WonBy == "team1" || game.WonBy == "team2" || game.WonBy == "phantom" || game.WonBy == "tesla" {
	} else {
		fmt.Printf("*** team wonby is not right %s \n", game.WonBy)
		er = er + 1
	}
	if game.Team1.Total == 0 || game.Team2.Total == 0 {
		fmt.Printf("*** missing total \n")
		er = er + 1
	}
	if game.Team1.OversPlayed == 0 || game.Team2.OversPlayed == 0 {
		fmt.Printf("*** missing overs \n")
		er = er + 1
	}
	mp = verifyFielders(game.Team1.DroppedCatches, game.Team2.TeamName)
	if len(mp) > 0 {
		fmt.Printf("*** missing players %s\n", mp)
		er = er + 1
	}
	mp = verifyFielders(game.Team2.DroppedCatches, game.Team1.TeamName)
	if len(mp) > 0 {
		fmt.Printf("*** missing players %s\n", mp)
		er = er + 1
	}

	if er > 0 {
		return 0
	}

	tx, err := dba.Begin()
	defer tx.Rollback()
	checkErr(err)

	inn1id := dba.getID("bat") + 1
	//fmt.Printf(" bat id %d\n", innid)
	inn2id := inn1id + 1
	//bowlInnId := getId("bowl") + 1
	//fmt.Printf(" bowl id %d\n", innid)
	fmt.Printf(" inn1 id %d inn2 id %d ", inn1id, inn2id)
	insertInnings(tx, inn1id, game.Team1, game.Team2.TeamName)
	insertInnings(tx, inn2id, game.Team2, game.Team1.TeamName)
	insertDroppedCatches(tx, inn1id, game.Team2.TeamName, game.Team1.DroppedCatches)
	insertDroppedCatches(tx, inn2id, game.Team1.TeamName, game.Team2.DroppedCatches)
	insertGame(tx, game.GameDate, inn1id, inn2id, game.WonBy,
		game.Team1.Total, game.Team1.OversPlayed,
		game.Team2.Total, game.Team2.OversPlayed,
		game.Team1Name, game.Team2Name)

	tx.Commit()

	return 0
}
