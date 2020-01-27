package parser

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	mystruct "github.com/cricsum/pkg/struct"
)

//Batting data structure
type Batting struct {
	Name       string
	RunsScored int
	HowOut     string
	FielerName string
	BowlerName string
	BallsFaced int
	FoursHit   int
	SixesHit   int
}

//Bowling data structure
type Bowling struct {
	Name         string
	OversBowled  float32
	RunsConceded int
	Maiden       int
	Wickets      int
}

//Innings data structure
type Innings struct {
	TeamName       string //name of the team batting
	Bat            map[int]Batting
	Bowl           map[int]Bowling
	Total          int
	OversPlayed    float32
	DroppedCatches []string
}

//Game data structure
type Game struct {
	GameDate   string
	Innings1Id int
	Innings2Id int
	Team1Name  string
	Team2Name  string
	Team1      Innings
	Team2      Innings
	WonBy      string
}

func procBowling(line string) Bowling {
	var bg Bowling
	tokens := strings.Split(line, ",")
	if len(tokens[0]) == 0 {
		return bg
	}
	if len(tokens) < 5 {
		return bg
	}

	bg.Name = strings.Trim(tokens[0], " ")
	ob, err := strconv.ParseFloat(tokens[1], 64)
	if err == nil {
		bg.OversBowled = float32(ob)
	}
	in, err := strconv.Atoi(tokens[3])
	if err == nil {
		bg.RunsConceded = in
	}
	in, err = strconv.Atoi(tokens[2])
	if err == nil {
		bg.Maiden = in
	}
	in, err = strconv.Atoi(tokens[4])
	if err == nil {
		bg.Wickets = in
	}
	return bg

}
func procBatting(line, version string, linecounter int) (Batting, int) {
	var bg Batting
	tokens := strings.Split(line, ",")
	vf, _ := strconv.ParseFloat(version, 64)
	if len(tokens[0]) == 0 {
		fmt.Printf("batting line [%d] is empty\n", linecounter)
		return bg, 1
	}
	if len(tokens) < 5 && vf < 2 {
		fmt.Printf("batting line [%d] should have batsman name, howout, fielder, bowler name, runs scored\n", linecounter)
		return bg, 1
	}
	if len(tokens) < 8 && vf >= 2 {
		fmt.Printf("batting line [%d] should have batsman name, howout, fielder, bowler name, runs scored, balls faced, # of 4s, # of 6s\n", linecounter)
		return bg, 1
	}
	bg.Name = strings.Trim(tokens[0], " ")
	bg.HowOut = strings.Trim(tokens[1], " ")
	bg.FielerName = strings.Trim(tokens[2], " ")
	bg.BowlerName = strings.Trim(tokens[3], " ")
	runs, err := strconv.Atoi(tokens[4])
	//fmt.Fprintf(w,"err %s\n", err)
	if err == nil {
		bg.RunsScored = runs
	}
	if vf >= 2 {
		bg.BallsFaced, _ = strconv.Atoi(tokens[5])
		bg.FoursHit, _ = strconv.Atoi(tokens[6])
		bg.SixesHit, _ = strconv.Atoi(tokens[7])
	}
	if !validInnings(bg, version, linecounter) {
		return bg, 1
	}
	return bg, 0
}
func getTeamName(commentLine string) string {
	/*
		the file has the following format
		==
		#inn1/team1name,
		batting team
		#blg2/team2name
		bowling team
		#inn2/team2name
		#blg1/team1name
		==
		if no team is specified, it is assumed to be phantom
	*/
	tokens := strings.Split(commentLine, ",")
	tempstr := tokens[0]
	tokens = strings.Split(tempstr, "/")
	teamName := ""
	if len(tokens) > 1 {
		teamName = tokens[1]
	} else {
		teamName = "phantom"
	}
	return teamName
}
func getScore(scoreline string) (int, float32) {
	/*
		#score,95,overs ,22,
	*/
	tokens := strings.Split(scoreline, ",")
	ival, err := strconv.Atoi(tokens[1])
	var score int
	var overs float32
	if err == nil {
		score = ival
	}
	var ovs float64
	ovs, err = strconv.ParseFloat(tokens[3], 64)
	if err == nil {
		overs = float32(ovs)
	}
	return score, overs
}
func getDroppedPlayerName(droppedline string) string {
	/*
		#catchdropped,satishn,,,
	*/
	tokens := strings.Split(droppedline, ",")
	return strings.Trim(tokens[1], " ")
}

//ReadLine parses the score file in csv format and returns a game object
func ReadLine(filename string) (Game, int) {
	start := time.Now()
	log.Println("loading from ", filename)
	inFile, err := os.Open(filename)
	defer inFile.Close()
	if err != nil {
		log.Fatal(err)
		//TODO:return nil
	}
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)

	linecounter := 0
	batBowlCount := 0
	idx := 0
	var inn1 Innings
	var inn2 Innings
	inn1.Bat = make(map[int]Batting)
	inn2.Bat = make(map[int]Batting)
	inn1.Bowl = make(map[int]Bowling)
	inn2.Bowl = make(map[int]Bowling)
	game := Game{}

	version := "0"
	ec := 0
	for scanner.Scan() {
		line := scanner.Text()

		linecounter++
		/*
			if linecounter == 1 {
				continue
			}
		*/

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#inn") == true {
			batBowlCount++
			idx = 0
			if batBowlCount == 1 {
				inn1.TeamName = getTeamName(line)
			}
			if batBowlCount == 3 {
				inn2.TeamName = getTeamName(line)
			}
			continue
		}
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "#blg") == true {
			idx = 0
			batBowlCount++
			continue
		}
		if strings.HasPrefix(line, "#version") == true {
			tokens := strings.Split(line, ",")
			version = tokens[1]
		}
		if strings.HasPrefix(line, "#wonby") == true {
			tokens := strings.Split(line, ",")
			game.WonBy = tokens[1]
			if len(tokens) == 4 {
				game.Team1Name = tokens[2]
				game.Team2Name = tokens[3]
			} else {
				game.Team1Name = "team1"
				game.Team2Name = "team2"
			}
			fmt.Printf("%s Vs %s wonby %s\n", game.Team1Name, game.Team2Name, game.WonBy)
			continue
		}

		if strings.HasPrefix(line, "#score") == true {
			score, overs := getScore(line)
			if batBowlCount > 2 {
				inn2.Total = score
				inn2.OversPlayed = overs
			} else {
				inn1.Total = score
				inn1.OversPlayed = overs
			}
			continue
		}
		if strings.HasPrefix(line, "#catchdropped") == true {
			plname := getDroppedPlayerName(line)
			if len(plname) == 0 {
				fmt.Printf(" *** ignoring dropped line # %d [%s]\n", linecounter, line)
				continue
			}

			if batBowlCount > 2 {
				inn2.DroppedCatches = append(inn2.DroppedCatches, plname)
			} else {
				inn1.DroppedCatches = append(inn1.DroppedCatches, plname)
			}
			continue
		}

		if batBowlCount == 1 || batBowlCount == 3 {
			bat, e := procBatting(line, version, linecounter)
			ec += e
			if e > 0 {
				continue
			}
			// if len(bat.Name) == 0 {
			// 	fmt.Printf(" *** ignoring batting line # %d [%s]\n", linecounter, line)
			// 	continue
			// }
			idx++
			if batBowlCount == 1 {
				inn1.Bat[idx] = bat
			}
			if batBowlCount == 3 {
				inn2.Bat[idx] = bat
			}
		}
		if batBowlCount == 2 || batBowlCount == 4 {
			ball := procBowling(line)
			if len(ball.Name) == 0 {
				fmt.Printf(" *** ignoring bowling line # %d [%s]\n", linecounter, line)
				continue
			}
			idx++
			if batBowlCount == 2 {
				inn1.Bowl[idx] = ball
			}
			if batBowlCount == 4 {
				inn2.Bowl[idx] = ball
			}
		}
	}
	game.Team1 = inn1
	game.Team2 = inn2
	tokens := strings.Split(filename, ".")
	game.GameDate = tokens[0]
	log.Printf("readline %s ", time.Since(start))
	return game, ec
}

func validInnings(bat Batting, version string, linecounter int) bool {
	if len(bat.Name) == 0 {
		fmt.Printf(" *** ignoring batting line # %d. Batsman name can't be empty\n", linecounter)
		return false
	}
	if bat.RunsScored < ((bat.FoursHit * 4) + (bat.SixesHit * 6)) {
		fmt.Printf(" *** ignoring batting line # %d. Runs scored is less than boundaries scored\n", linecounter)
		return false
	}
	if bat.BallsFaced < (bat.FoursHit + bat.SixesHit) {
		fmt.Printf(" *** ignoring batting line # %d. Balls faced can't be < boundary balls\n", linecounter)
		return false
	}
	vf, _ := strconv.ParseFloat(version, 64)
	if vf >= 2 {
		if bat.RunsScored > 0 && bat.BallsFaced <= 0 {
			fmt.Printf(" *** ignoring batting line # %d. Runs can't be scored with out facing a ball\n", linecounter)
			return false
		}
	}
	//var ho mystruct.Howout //= "b" // = string(bat.HowOut)
	ho := mystruct.Howout(string(bat.HowOut))
	//ho = bat.HowOut
	er := ho.Valid()
	if er != nil {
		fmt.Printf(" *** ignoring batting line # %d. %s\n", linecounter, er)
		return false
	}

	return true
}

//GenHTML generate scorecard in html format
func (g Game) GenHTML(date string) {
	fo, err := os.Create(date + ".html")
	if err != nil {
		panic(err)
	}
	defer fo.Close()
	w := bufio.NewWriter(fo)
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "<body>")
	fmt.Fprintf(w, "<b>Date </b>%s&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;<b>Won by </b> %s ", g.GameDate, g.WonBy)

	i1bat, i1bwl, i1dc := inningsTable(g.Team1)
	i2bat, i2bwl, i2dc := inningsTable(g.Team2)
	team1Name := "Team1"
	team2Name := "Team2"

	fmt.Fprintf(w, "<table>")

	//header
	fmt.Fprintf(w, "<tr>\n")

	fmt.Fprintf(w, "<th bgcolor=\"#d9d9d9\">\n")
	fmt.Fprintf(w, "Innings of %s\n", team1Name)
	fmt.Fprintf(w, "</th>\n")

	fmt.Fprintf(w, "<th bgcolor=\"#d9d9d9\">\n")
	fmt.Fprintf(w, "Innings of %s\n", team2Name)
	fmt.Fprintf(w, "</th>\n")

	fmt.Fprintf(w, "</tr>\n")

	//batting
	fmt.Fprintf(w, "<tr>\n")

	fmt.Fprintf(w, "<td>\n")
	fmt.Fprintf(w, "%s\n", i1bat)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "<td>\n")
	fmt.Fprintf(w, "%s\n", i2bat)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "</tr>\n")

	//bowling
	fmt.Fprintf(w, "<tr>\n")

	fmt.Fprintf(w, "<td>\n")
	fmt.Fprintf(w, "%s\n", i1bwl)
	fmt.Fprintf(w, "</td>\n")
	fmt.Fprintf(w, "<td>\n")
	fmt.Fprintf(w, "%s\n", i2bwl)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "</tr>\n")

	//dropped catches
	fmt.Fprintf(w, "<tr>\n")

	fmt.Fprintf(w, "<td>\n")
	fmt.Fprintf(w, "%s\n", i1dc)
	fmt.Fprintf(w, "</td>\n")
	fmt.Fprintf(w, "<td>\n")
	fmt.Fprintf(w, "%s\n", i2dc)
	fmt.Fprintf(w, "</td>\n")

	fmt.Fprintf(w, "</tr>\n")

	fmt.Fprintf(w, "</table>")

	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")
	w.Flush()
}
func inningsTable(inn Innings) (string, string, string) {
	var battbl, bltbl, dctbl string
	const bgcolor = "bgcolor=\"#d9d9d9\""
	battbl += fmt.Sprintf("<table>")

	battbl += fmt.Sprintf("<tr>\n")
	battbl += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, "Total")
	battbl += fmt.Sprintf("<th %s colspan=\"4\" >%s %3.2f</th>\n", bgcolor, "Ov", inn.OversPlayed)
	battbl += fmt.Sprintf("<th %s colspan=\"4\" >%d %s</th>\n", bgcolor, inn.Total, "Runs")
	battbl += fmt.Sprintf("</tr>\n")
	battbl += fmt.Sprintf("<tr>\n")
	battbl += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, "Batsman")
	battbl += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, "how out")
	battbl += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, "fielder")
	battbl += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, "bowler")
	battbl += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, "runs")
	battbl += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, "B")
	battbl += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, "4s")
	battbl += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, "6s")
	battbl += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, "SR")
	battbl += fmt.Sprintf("</tr>\n")

	for i := 1; i <= len(inn.Bat); i++ {
		bt := inn.Bat[i]
		battbl += fmt.Sprintf("<tr>\n")
		battbl += fmt.Sprintf("<td>%s</td>\n", bt.Name)
		battbl += fmt.Sprintf("<td>%s</td>\n", bt.HowOut)
		battbl += fmt.Sprintf("<td>%s</td>\n", bt.FielerName)
		battbl += fmt.Sprintf("<td>%s</td>\n", bt.BowlerName)
		battbl += fmt.Sprintf("<td>%d</td>\n", bt.RunsScored)
		battbl += fmt.Sprintf("<td>%d</td>\n", bt.BallsFaced)
		battbl += fmt.Sprintf("<td>%d</td>\n", bt.FoursHit)
		battbl += fmt.Sprintf("<td>%d</td>\n", bt.SixesHit)
		sr := float32(bt.RunsScored) / float32(bt.BallsFaced) * 100.0
		if bt.BallsFaced == 0 {
			sr = 0
		}
		battbl += fmt.Sprintf("<td>%3.1f</td>\n", sr)
		battbl += fmt.Sprintf("</tr>\n")
	}
	battbl += fmt.Sprintf("</table>")
	//	bltbl += fmt.Sprintf("<br>")

	bltbl += fmt.Sprintf("<table>")
	bltbl += fmt.Sprintf("<tr>\n")
	bltbl += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, "Bowler")
	bltbl += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, "O")
	bltbl += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, "M")
	bltbl += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, "R")
	bltbl += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, "W")
	bltbl += fmt.Sprintf("<th %s>%s</th>\n", bgcolor, "Econ")
	bltbl += fmt.Sprintf("</tr>\n")

	for i := 1; i <= len(inn.Bowl); i++ {
		bl := inn.Bowl[i]
		bltbl += fmt.Sprintf("<tr>\n")
		bltbl += fmt.Sprintf("<td>%s</td>\n", bl.Name)
		bltbl += fmt.Sprintf("<td>%g</td>\n", bl.OversBowled)
		bltbl += fmt.Sprintf("<td>%d</td>\n", bl.Maiden)
		bltbl += fmt.Sprintf("<td>%d</td>\n", bl.RunsConceded)
		bltbl += fmt.Sprintf("<td>%d</td>\n", bl.Wickets)
		econ := float32(bl.RunsConceded) / bl.OversBowled
		bltbl += fmt.Sprintf("<td>%3.1f</td>\n", econ)
		bltbl += fmt.Sprintf("</tr>\n")
	}
	bltbl += fmt.Sprintf("</table>")
	//dctbl += fmt.Sprintf("<br>")

	dctbl += fmt.Sprintf("<b>Dropped Catches </b>\n ")
	dctbl += fmt.Sprintf("<br>")
	for _, dc := range inn.DroppedCatches {
		dctbl += fmt.Sprintf("%s\n", dc)
		dctbl += fmt.Sprintf("<br>")
	}
	dctbl += fmt.Sprintf("<br>")
	return battbl, bltbl, dctbl
}
