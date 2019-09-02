package parser

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Batting struct {
	Name       string
	RunsScored int
	HowOut     string
	FielerName string
	BowlerName string
}
type Bowling struct {
	Name         string
	OversBowled  float32
	RunsConceded int
	Maiden       int
	Wickets      int
}
type Innings struct {
	TeamName       string //name of the team batting
	Bat            map[int]Batting
	Bowl           map[int]Bowling
	Total          int
	OversPlayed    float32
	DroppedCatches []string
}
type Game struct {
	GameDate   string
	Innings1Id int
	Innings2Id int
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

	bg.Name = tokens[0]
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
func procBatting(line string) Batting {
	var bg Batting
	tokens := strings.Split(line, ",")
	if len(tokens[0]) == 0 {
		return bg
	}
	bg.Name = tokens[0]
	bg.HowOut = tokens[1]
	bg.FielerName = tokens[2]
	bg.BowlerName = tokens[3]
	runs, err := strconv.Atoi(tokens[4])
	//fmt.Printf("err %s\n", err)
	if err == nil {
		bg.RunsScored = runs
	}
	return bg
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
	return tokens[1]
}

//
func ReadLine(filename string) Game {
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
			batBowlCount += 1
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
			batBowlCount += 1
			continue
		}
		if strings.HasPrefix(line, "#wonby") == true {
			tokens := strings.Split(line, ",")
			game.WonBy = tokens[1]
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
			if batBowlCount > 2 {
				inn2.DroppedCatches = append(inn2.DroppedCatches, plname)
			} else {
				inn1.DroppedCatches = append(inn1.DroppedCatches, plname)
			}
			continue
		}

		if batBowlCount == 1 || batBowlCount == 3 {
			bat := procBatting(line)
			if len(bat.Name) == 0 {
				fmt.Printf(" *** ignoring batting line # %d [%s]\n", linecounter, line)
				continue
			}
			idx += 1
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
			idx += 1
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
	return game
}
