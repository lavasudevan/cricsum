package main

import (
	"flag"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/cricsum/pkg/db"
	"github.com/cricsum/pkg/parser"
)

func getSummary(teamname string) {
	rs := db.GetSummary()
	fmt.Printf("\n%-15s,%-15s,%-6s,%-7s,%-7s,%7s,,%6s,%5s,%6s,%7s,%8s,%9s,%-15s\n",
		"player", "innings_played", "notout", "runs", "average", "highest",
		"Overs", "Runs", "maiden", "wickets", "RPO", "Dismissal", "player")

	keys := make([]string, 0, len(rs))
	for k, _ := range rs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		v := rs[name]
		tokens := strings.Split(name, "/")
		plname := tokens[0]
		tname := tokens[1]
		if tname != teamname {
			continue
		}
		fmt.Printf("%-15s,%-15d,%s,%-7d,%7.2f,%7d,,%6.1f,%5s,%6s,%7s,%8.2f,%9s,%-15s\n",
			plname, v.InningsPlayed, numberFormat(v.NotOut), v.RunsScored, v.Average, v.Highest, v.OversBowled,
			numberFormat(v.RunsConceded), numberFormat(v.Maiden), numberFormat(v.Wickets), v.RunsPerOver,
			numberFormat(v.Dismissal), plname)
	}
}

func getDetails() {
	rs := db.GetDetails()
	fmt.Println("Name,date,Runs,howout,overs,maiden,runsconceded,wickets")

	var keys []string
	for k := range rs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := rs[k]
		tokens := strings.Split(v.Name, "/")
		if tokens[1] != "phantom" {
			continue
		}
		fmt.Printf("%s,%s,%0d,%s,%6.2f,%d,%d,%d\n",
			tokens[0], v.Date, v.RunsScored, v.HowOut, v.OversBowled, v.Maiden, v.RunsConceded, v.Wickets)
	}
}
func numberFormat(i int) string {
	if i == 0 {
		return ""
	} else {
		return strconv.Itoa(i)
	}
}

func main() {
	var command, file string
	flag.StringVar(&command, "command", "", "a string")
	flag.StringVar(&file, "scorefile", "", "a string")
	flag.Parse()
	if command == "summary" {
		getSummary("phantom")
	}
	if command == "details" {
		getDetails()
	}
	if command == "upload" {
		gm := parser.ReadLine(file)
		fmt.Printf("date of file %s\n", gm.GameDate)
		db.UpdateGame(gm)
	}

}
