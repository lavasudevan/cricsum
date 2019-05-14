package parser

import (
	"bufio"
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
	Bat         map[int]Batting
	Bowl        map[int]Bowling
	Total       int
	OversPlayed float32
}
type Game struct {
	GameDate string
	Team1    Innings
	Team2    Innings
}

func procBowling(line string) Bowling {
	var bg Bowling
	tokens := strings.Split(line, ",")
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
	hashCount := 0
	idx := 0
	var inn1 Innings
	var inn2 Innings
	inn1.Bat = make(map[int]Batting)
	inn2.Bat = make(map[int]Batting)
	inn1.Bowl = make(map[int]Bowling)
	inn2.Bowl = make(map[int]Bowling)

	for scanner.Scan() {
		line := scanner.Text()

		linecounter++
		/*
			if linecounter == 1 {
				continue
			}
		*/

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") == true {
			hashCount += 1
			idx = 0
			continue
		}
		if len(line) == 0 {
			continue
		}
		idx += 1
		if hashCount == 1 || hashCount == 3 {
			bat := procBatting(line)
			if hashCount == 1 {
				inn1.Bat[idx] = bat
			}
			if hashCount == 3 {
				inn2.Bat[idx] = bat
			}
		}
		if hashCount == 2 || hashCount == 4 {
			ball := procBowling(line)
			if hashCount == 2 {
				inn1.Bowl[idx] = ball
			}
			if hashCount == 4 {
				inn2.Bowl[idx] = ball
			}
		}

	}
	//	game := &Game{}
	game := Game{}
	game.Team1 = inn1
	game.Team2 = inn2
	tokens := strings.Split(filename, ".")
	game.GameDate = tokens[0]
	log.Printf("readline %s ", time.Since(start))
	return game
}
