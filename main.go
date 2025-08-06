package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	//"strings"

	"new-gamebook/fightingfantasy"
	"new-gamebook/game"
	"new-gamebook/lonewolf"

	"github.com/BurntSushi/toml"
)

// NewGameSystem はシステム名に基づいて GameSystem を返す
func NewGameSystem(systemName, configDir string) (game.GameSystem, error) {
	switch systemName {
	case "lonewolf":
		return lonewolf.NewLoneWolfSystem(configDir + "/combat_result_table.toml"), nil
	case "fightingfantasy":
		return fightingfantasy.NewFightingFantasySystem(), nil
	default:
		return nil, fmt.Errorf("unknown system: %s", systemName)
	}
}

// NewGameState はゲーム状態を初期化
func NewGameState(config *game.GameConfig, configDir string) (*GameState, error) {
	system, err := NewGameSystem(config.System, configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create game system: %w", err)
	}
	if err := system.Initialize(config); err != nil {
		return nil, fmt.Errorf("failed to initialize system: %w", err)
	}

	gs := &game.GameState{ //GameStateを初期化
		Nodes:         make(map[string]game.Node),
		Reader:        bufio.NewReader(os.Stdin),
		System:        system, // System を設定
		CurrentNodeID: "1",
	}
	// ...（Player, Nodes の初期化コード）

	//Nodeマップ作成　Configに
	nodeMap := make(map[string]game.Node)
	for _, node := range config.Nodes {
		nodeMap[node.ID] = node
	}

	player := &lonewolf.Player{
		Stats:      make(map[string]int),
		Attributes: make(map[string]bool),
		Equipments: &lonewolf.Equipment{
			Head:          nil,
			Body:          nil,
			Currentweapon: 0,
			Weapon1:       nil,
			Weapon2:       nil,
			Shield:        false,
			Backpack:      []*lonewolf.Item{},
		},
		Gold: 0,
	}

	if stats, ok := config.Player["stats"].(map[string]interface{}); ok {
		for k, v := range stats {
			if val, ok := v.(int64); ok {
				player.Stats[k] = int(val)
			}
		}
	}
	if attributes, ok := config.Player["attributes"].(map[string]interface{}); ok {
		for k, v := range attributes {
			if val, ok := v.(bool); ok {
				player.Attributes[k] = val
			}
		}
	}

	gs.Nodes = nodeMap
	gs.Player = player
	return gs, nil
}

// GameConfig はゲーム全体のTOML設定を表す
type GameConfig struct {
	System string          `toml:"system"`
	Nodes  []lonewolf.Node `toml:"nodes"`
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

	gameState, err := NewGameState(&config, ".")
	if err != nil {
		log.Fatalf("Error initializing game state: %v", err)
	}

	//gameState.DisplayStatus()
	gameState.Run(gs * lonewolf.GameState)
}
