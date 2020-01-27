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
	byRuns           statType = 0
	byWickets        statType = 1
	byDismissal      statType = 2
	byDroppedCatches statType = 3
	byBattingAvg     statType = 4
	byEconRate       statType = 5
	byNotout         statType = 6
	byNumberInnings  statType = 7
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
	head := "Runs"
	if st == byBattingAvg {
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
		head = "#"
	} else if st == byNumberInnings {
		head = "#"
	} else if st == byNotout {
		head = "#"
	}
	for k := range rs {
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
			kn = fmt.Sprintf("%04d#%s", v.RunsScored, k)
		} else if st == byWickets {
			kn = fmt.Sprintf("%04d#%s", v.Wickets, k)
		} else if st == byDismissal {
			kn = fmt.Sprintf("%04d#%s", v.Dismissal, k)
		} else if st == byNotout {
			kn = fmt.Sprintf("%04d#%s", v.NotOut, k)
		} else if st == byDroppedCatches {
			kn = fmt.Sprintf("%04d#%s", v.DroppedCatches, k)
		} else if st == byNumberInnings {
			kn = fmt.Sprintf("%04d#%s", v.InningsPlayed, k)
		} else if st == byBattingAvg {
			kn = fmt.Sprintf("%06d#%s", int(v.Average*100), k)
		} else if st == byEconRate {
			kn = fmt.Sprintf("%06d#%s", int(v.RunsPerOver*100), k)
		}
		keys = append(keys, kn)
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
	dct := getTable(rs, teamname, byDroppedCatches)
	bat := getTable(rs, teamname, byBattingAvg)
	ect := getTable(rs, teamname, byEconRate)
	not := getTable(rs, teamname, byNotout)
	nit := getTable(rs, teamname, byNumberInnings)
	fo, err := os.Create("summary.html")
	if err != nil {
		panic(err)
	}
	defer fo.Close()
	w := bufio.NewWriter(fo)
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "<table>")

	fmt.Fprintf(w, "<td>\n")
	fmt.Fprintf(w, rt)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td>\n")
	fmt.Fprintf(w, bat)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td>\n")
	fmt.Fprintf(w, wt)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td>\n")
	fmt.Fprintf(w, ect)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td>\n")
	fmt.Fprintf(w, dt)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td>\n")
	fmt.Fprintf(w, dct)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td>\n")
	fmt.Fprintf(w, not)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td>\n")
	fmt.Fprintf(w, nit)
	fmt.Fprintf(w, "</td>\n")

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
	st += fmt.Sprintf("<th colspan=2 %s>%s</th>", bgcolor, title)
	st += fmt.Sprintf("</tr>\n")
	st += fmt.Sprintf("<tr>\n")

	st += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, "Name")
	st += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, rowHeading)
	st += fmt.Sprintf("</tr>\n")

	for i := 0; i < len(s.Data); i++ {
		tokens := strings.Split(s.Data[i], "#")
		ntokens := strings.Split(tokens[1], "/")
		st += fmt.Sprintf("<tr>\n")
		r, _ := strconv.ParseFloat(tokens[0], 64)
		if divBy100 == 1 {
			r = r / 100.0
		}
		st += fmt.Sprintf("<td>%s</td>\n", ntokens[0])
		st += fmt.Sprintf("<td>%g</td>\n", r)
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
	fmt.Println("--command=summary [--year=yyyy]")
	fmt.Println("--command=details")
	os.Exit(1)
}
func main() {
	var command, file, date, year string
	flag.StringVar(&command, "command", "", "a string")
	flag.StringVar(&file, "scorefile", "", "a string")
	flag.StringVar(&year, "year", "", "yyyy for the year. if empty year is assumed for the current year")
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
		getSummary("phantom", year)
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
		gm := parser.ReadLine(file)
		fmt.Printf("date of file %s\n", gm.GameDate)
		mdb.UpdateGame(gm)
		gm.GenHTML(gm.GameDate)
	} else {
		usage()
	}
}
