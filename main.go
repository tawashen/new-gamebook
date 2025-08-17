package main

import (
	"io/ioutil"
	"log"

	//"strings"

	//"new-gamebook/fightingfantasy"
	"new-gamebook/game"
	"new-gamebook/lonewolf"

	"github.com/BurntSushi/toml"
)

// GameConfig はゲーム全体のTOML設定を表す
type GameConfig struct {
	System string `toml:"system"`
	//	Player Player
	//Nodes []Node `toml:"nodes"`
}

func main() {
	tomlData, err := ioutil.ReadFile("testlw.toml")
	if err != nil {
		log.Fatalf("Error reading TOML file: %v", err)
	}

	var config GameConfig
	if _, err := toml.Decode(string(tomlData), &config); err != nil {
		log.Fatalf("Error decoding TOML: %v", err)
	}

	var gs game.GameSystem

	switch config.System {
	case "lonewolf":
		gs = lonewolf.NewLoneWolfSystem("combat_results_table.toml")
	}

	//gameState.DisplayStatus()
	gs.Run()

}
