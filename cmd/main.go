package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cricsum/pkg/db"
	"github.com/cricsum/pkg/parser"
	_ "github.com/mattn/go-sqlite3"
)

const (
	dbname = "scores.db"
)

var mdb *db.SDB

func openDb() {
	var err error
	var dba *sql.DB
	if mdb == nil {
		dba, err = sql.Open("sqlite3", "./"+dbname)
		checkErr(err)
		mdb = new(db.SDB)
		mdb.DB = dba
	}
}
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

type statType int

const (
	byRuns               statType = 0
	byWickets            statType = 1
	byDismissal          statType = 2
	byDroppedCatches     statType = 3
	byBattingAvg         statType = 4
	byEconRate           statType = 5
	byNotout             statType = 6
	byNumberInnings      statType = 7
	byNumberofBallsFaced statType = 8
	byStrikeRate         statType = 9
	byBoundariesHit      statType = 10
	byOversPerWicket     statType = 11
	byOversBowled        statType = 12
	byRunsPerWicket      statType = 13
)

func (st statType) string() string {
	titles := [...]string{
		"By Runs",
		"By Wickets",
		"By Dismissal",
		"By Dropped catches",
		"By Batting Avg",
		"By Economy rate",
		"By Not out",
		"By # of innings",
		"By # of balls faced",
		"By Strike Rate",
		"By # of Boundaries hit",
		"Overs per wicket",
		"Overs bowled",
		"Runs per wicket",
	}
	return titles[st]
}
func getTeamName(plnamewt string) (string, string) {
	tokens := strings.Split(plnamewt, "/")
	//plname := tokens[0]
	//tname := tokens[1]
	return tokens[0], tokens[1]
}
func getTable(rs map[string]db.Summary, teamname string, st statType) string {
	title := st.string()
	keys := make([]string, 0, len(rs))
	var sortorder int
	//= 1 if no reverse for econ rate. 0 reverse
	var divBy100 int
	var head string
	if st == byRuns {
		head = "Runs,BBI"
	} else if st == byBattingAvg {
		head = "Avg"
		divBy100 = 1
	} else if st == byEconRate {
		sortorder = 1
		divBy100 = 1
		head = "RPO"
	} else if st == byDroppedCatches {
		head = "#"
	} else if st == byDismissal {
		head = "#"
	} else if st == byWickets {
		head = "#, BBM"
	} else if st == byNumberInnings {
		head = "#"
	} else if st == byNotout {
		head = "#"
	} else if st == byNumberofBallsFaced {
		head = "#"
	} else if st == byStrikeRate {
		head = "%%"
	} else if st == byOversPerWicket {
		head = "#"
		sortorder = 1
		divBy100 = 1
	} else if st == byRunsPerWicket {
		head = "#"
		sortorder = 1
		divBy100 = 1
	} else if st == byOversBowled {
		head = "#"
		divBy100 = 1
	} else if st == byBoundariesHit {
		head = "#"
	}
	for k := range rs {
		ignore := 0
		v := rs[k]
		if db.IsPlayerActive(v.PlayerID) == false {
			continue
		}
		_, tname := getTeamName(k)

		if tname != teamname {
			continue
		}
		var kn string
		if st == byRuns {
			kn = fmt.Sprintf("%04d#%s,%d", v.RunsScored, k, v.Highest)
		} else if st == byWickets {
			kn = fmt.Sprintf("%04d#%s,%d-%d", v.Wickets, k, v.BestWickets, v.BestWicketsRuns)
			if int(v.OversBowled) == 0 {
				ignore = 1
			}
		} else if st == byDismissal {
			kn = fmt.Sprintf("%04d#%s", v.Dismissal, k)
			if int(v.Dismissal) == 0 {
				ignore = 1
			}
		} else if st == byNotout {
			kn = fmt.Sprintf("%04d#%s", v.NotOut, k)
			if v.NotOut == 0 {
				ignore = 1
			}
		} else if st == byDroppedCatches {
			kn = fmt.Sprintf("%04d#%s", v.DroppedCatches, k)
		} else if st == byNumberInnings {
			kn = fmt.Sprintf("%04d#%s", v.InningsPlayed, k)
		} else if st == byBattingAvg {
			kn = fmt.Sprintf("%06d#%s", int(v.Average*100), k)
		} else if st == byEconRate {
			kn = fmt.Sprintf("%06d#%s", int(v.RunsPerOver*100), k)
			if int(v.OversBowled) == 0 {
				ignore = 1
			}

		} else if st == byNumberofBallsFaced {
			kn = fmt.Sprintf("%04d#%s", v.BallsFaced, k)
		} else if st == byStrikeRate {
			sr := float32(v.RunsScored) / float32(v.BallsFaced) * 100.0
			if v.BallsFaced == 0 {
				ignore = 1
			}
			kn = fmt.Sprintf("%04d#%s", int(sr), k)
		} else if st == byOversPerWicket {
			sr := v.OversBowled / float32(v.Wickets) * 100
			if v.Wickets == 0 {
				sr = 0
			}
			if v.Wickets == 0 {
				ignore = 1
			}

			kn = fmt.Sprintf("%04d#%s", int(sr), k)
		} else if st == byOversBowled {
			sr := v.OversBowled * 100
			kn = fmt.Sprintf("%04d#%s", int(sr), k)
			if int(sr) == 0 {
				ignore = 1
			}
		} else if st == byRunsPerWicket {
			sr := v.RunsPerWicket * 100
			kn = fmt.Sprintf("%04d#%s", int(sr), k)
			if int(sr) == 0 {
				ignore = 1
			}
		} else if st == byBoundariesHit {
			bh := v.FoursHit + v.SixesHit
			kn = fmt.Sprintf("%04d#%s", bh, k)
			if bh == 0 {
				ignore = 1
			}
		}

		if ignore == 0 {
			keys = append(keys, kn)
		}
	}
	if sortorder == 0 {
		sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	} else {
		sort.Sort(sort.StringSlice(keys))
	}
	var s Stat
	s.Data = keys
	return s.dumpHTMLTable(title, head, divBy100)
}

func getSummary(teamname, year string) {
	rs := mdb.GetSummary(year)
	fmt.Printf("%-15s,%-15s,%-6s,%-7s,%-7s,%7s,,%6s,%5s,%6s,%7s,%8s,%9s,%s,%s,%-15s\n",
		"player", "innings_played", "notout", "runs", "average", "highest",
		"Overs", "Runs", "maiden", "wickets", "RPO", "Dismissal", "dropped Catches", "BBM", "player")

	keys := make([]string, 0, len(rs))
	for k := range rs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		v := rs[name]
		plname, tname := getTeamName(name)
		if tname != teamname {
			continue
		}
		fmt.Printf("%-15s,%-15d,%s,%-7d,%7.2f,%7d,,%6.1f,%5s,%6s,%7s,%8.2f,%9s,%d,=\"%d-%d\",%-15s\n",
			plname, v.InningsPlayed, numberFormat(v.NotOut), v.RunsScored, v.Average, v.Highest, v.OversBowled,
			numberFormat(v.RunsConceded), numberFormat(v.Maiden), numberFormat(v.Wickets), v.RunsPerOver,
			numberFormat(v.Dismissal), v.DroppedCatches, v.BestWickets, v.BestWicketsRuns, plname)
	}

	rt := getTable(rs, teamname, byRuns)
	fmt.Println(keys)

	wt := getTable(rs, teamname, byWickets)
	dt := getTable(rs, teamname, byDismissal)
	// dct := getTable(rs, teamname, byDroppedCatches)
	bat := getTable(rs, teamname, byBattingAvg)
	ect := getTable(rs, teamname, byEconRate)
	not := getTable(rs, teamname, byNotout)
	nit := getTable(rs, teamname, byNumberInnings)
	bf := getTable(rs, teamname, byNumberofBallsFaced)
	bb := getTable(rs, teamname, byBoundariesHit)
	bs := getTable(rs, teamname, byStrikeRate)
	bo := getTable(rs, teamname, byOversPerWicket)
	rw := getTable(rs, teamname, byRunsPerWicket)
	bob := getTable(rs, teamname, byOversBowled)

	fo, err := os.Create("summary.html")
	if err != nil {
		panic(err)
	}
	defer fo.Close()
	w := bufio.NewWriter(fo)
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "<table>")

	fmt.Fprintf(w, "<tr>")

	fmt.Fprintf(w, "<td>\n")
	fmt.Fprintf(w, rt)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td>\n")
	fmt.Fprintf(w, bat)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td valign=\"top\">\n")
	fmt.Fprintf(w, not)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td>\n")
	fmt.Fprintf(w, nit)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td>\n")
	fmt.Fprintf(w, bf)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td valign=\"top\">\n")
	fmt.Fprintf(w, bb)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td valign=\"top\">\n")
	fmt.Fprintf(w, bs)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "</tr>")
	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, "</tr>")
	fmt.Fprintf(w, "</br>")
	fmt.Fprintf(w, "<tr>")

	fmt.Fprintf(w, "<td valign=\"top\">\n")
	fmt.Fprintf(w, wt)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td valign=\"top\">\n")
	fmt.Fprintf(w, ect)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td valign=\"top\">\n")
	fmt.Fprintf(w, bo)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td valign=\"top\">\n")
	fmt.Fprintf(w, rw)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td valign=\"top\">\n")
	fmt.Fprintf(w, bob)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td valign=\"top\">\n")
	fmt.Fprintf(w, dt)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "</tr>")

	// fmt.Fprintf(w, "<td>\n")
	// fmt.Fprintf(w, dct)
	// fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "</html>\n")
	w.Flush()
}

//Stat structure
type Stat struct {
	Data []string
}

func (s Stat) dumpHTMLTable(title string, rowHeading string, divBy100 int) string {
	var st string
	const bgcolor = "bgcolor=\"#d9d9d9\""
	st += fmt.Sprintf("<table>")
	st += fmt.Sprintf("<tr>\n")
	tokens := strings.Split(rowHeading, ",")
	st += fmt.Sprintf("<th colspan=%d %s>%s</th>", len(tokens)+1, bgcolor, title)
	st += fmt.Sprintf("</tr>\n")
	st += fmt.Sprintf("<tr>\n")

	st += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, "Name")
	for _, h := range tokens {
		st += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, h)
	}
	//st += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, rowHeading)
	st += fmt.Sprintf("</tr>\n")

	/*
		byRuns, byWickets have highest scores / BBM seperated by ,. like
		156#raj/phantom,69
	*/
	for i := 0; i < len(s.Data); i++ {
		tokens := strings.Split(s.Data[i], "#")
		dtkns := strings.Split(tokens[1], ",")
		ntokens := strings.Split(dtkns[0], "/")
		st += fmt.Sprintf("<tr>\n")
		r, _ := strconv.ParseFloat(tokens[0], 64)
		if divBy100 == 1 {
			r = r / 100.0
		}
		st += fmt.Sprintf("<td>%s</td>\n", ntokens[0])
		st += fmt.Sprintf("<td>%g</td>\n", r)
		if len(dtkns) > 1 {
			st += fmt.Sprintf("<td>%s</td>\n", dtkns[1])
		}
		st += fmt.Sprintf("</tr>\n")
	}
	st += fmt.Sprintf("</table>")
	return st
}

func getDetails() {
	rs := mdb.GetDetails()
	fmt.Println("Name,date,Runs,howout,dismissal,catchdropped,overs,maiden,runsconceded,wickets")

	//Key has the form playername/dateG
	var keys []string
	for k := range rs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := rs[k]
		tokens := strings.Split(v.Name, "/")
		if len(tokens) == 1 {
			fmt.Printf("zero")
		}
		if tokens[1] != "phantom" {
			continue
		}
		fmt.Printf("%s,%s,%0d,%s,%d,%d,%6.2f,%d,%d,%d\n",
			tokens[0], v.Date, v.RunsScored, v.HowOut, v.Dismissal, v.DroppedCatches, v.OversBowled, v.Maiden, v.RunsConceded, v.Wickets)
	}
}
func numberFormat(i int) string {
	if i == 0 {
		return ""
	}
	return strconv.Itoa(i)
}
func usage() {
	fmt.Println("--command=remove --date=yyyymmdd")
	fmt.Println("--command=upload --scorefile=yyyymmdd.csv")
	fmt.Println("--command=summary [--year=yyyy] [--team=phantom/tesla]")
	fmt.Println("--command=details")
	os.Exit(1)
}
func main() {
	var command, file, date, year, team string
	flag.StringVar(&command, "command", "", "a string")
	flag.StringVar(&file, "scorefile", "", "a string")
	flag.StringVar(&year, "year", "", "yyyy for the year. if empty year is assumed for the current year")
	flag.StringVar(&team, "team", "phantom", "")
	flag.StringVar(&date, "date", "", "date string in yyyymmdd format")
	flag.Parse()
	openDb()
	if command == "summary" {
		if len(year) > 4 {
			usage()
		} else if len(year) == 0 {
			year = strconv.Itoa(time.Now().Year())
		}
	}

	if command == "summary" {
		getSummary(team, year)
	} else if command == "disable" {
		mdb.DisablePlayers(strconv.Itoa(time.Now().Year()))
	} else if command == "remove" {
		if date == "" {
			usage()
		}
		mdb.RemoveGame(date)
	} else if command == "details" {
		getDetails()
	} else if command == "upload" {
		gm, ec := parser.ReadLine(file)
		fmt.Printf("date of file %s\n", gm.GameDate)
		if ec == 0 {
			mdb.UpdateGame(gm)
			gm.GenHTML(gm.GameDate)
		}
	} else {
		usage()
	}
}
