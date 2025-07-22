package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	//srings"

	"new-gamebook/fightingfantasy"
	"new-gamebook/game"
	"new-gamebook/lonewolf"

	"github.com/BurntSushi/toml"
)

// GameSystem はゲームシステムのインターフェース
type GameSystem interface {
	Initialize(config *game.GameConfig) error
	HandleNode(gs *game.GameState, node game.Node) error
	UpdatePlayer(gs *game.GameState, action string) error
}

// NewGameSystem はシステム名に基づいて適切なGameSystemを返す
func NewGameSystem(systemName, configDir string) (GameSystem, error) {
	switch systemName {
	case "lonewolf":
		return lonewolf.NewLoneWolfSystem(configDir + "/combat_result_table.toml"), nil
	case "fightingfantasy":
		return fightingfantasy.NewFightingFantasySystem(), nil
	default:
		return nil, fmt.Errorf("unknown system: %s", systemName)
	}
}

// NewGameState は新しいGameStateを初期化
func NewGameState(config *game.GameConfig, configDir string) (*game.GameState, error) {
	system, err := NewGameSystem(config.System, configDir)
	if err != nil {
		return nil, err
	}

	nodeMap := make(map[string]game.Node)
	for _, node := range config.Nodes {
		nodeMap[node.ID] = node
	}

	player := &game.Player{
		Stats:      make(map[string]int),
		Attributes: make(map[string]bool),
		Inventory:  []string{},
		Equipment:  make(map[string]string),
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
	if inventory, ok := config.Player["inventory"].([]interface{}); ok {
		for _, item := range inventory {
			if str, ok := item.(string); ok {
				player.Inventory = append(player.Inventory, str)
			}
		}
	}

	gs := &game.GameState{
		Player:        player,
		CurrentNodeID: config.Nodes[0].ID,
		Nodes:         nodeMap,
		Reader:        bufio.NewReader(os.Stdin),
		Config:        config,
	}

	if err := system.Initialize(config); err != nil {
		return nil, err
	}
	return gs, nil
}

func main() {
	tomlData, err := ioutil.ReadFile("testlw.toml")
	if err != nil {
		log.Fatalf("Error reading TOML file: %v", err)
	}

	var config game.GameConfig
	if _, err := toml.Decode(string(tomlData), &config); err != nil {
		log.Fatalf("Error decoding TOML: %v", err)
	}

	gameState, err := NewGameState(&config, ".")
	if err != nil {
		log.Fatalf("Error initializing game state: %v", err)
	}

	gameState.Run()
}
