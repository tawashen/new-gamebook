package game

import (
	"bufio"
	"fmt"
	"new-gamebook/lonewolf"
)

// GameSystem はゲームシステムのインターフェース
type GameSystem interface {
	MakingPlayer(gs *GameState) error
	Initialize(config *GameConfig) error
	HandleNode(gs *GameState, node Node) error
	UpdatePlayer(gs *GameState, action string) error
}

// Node はゲームの各ステップ（ノード）を表す
type Node struct {
	ID       string    `toml:"id"`
	Type     string    `toml:"type"`
	Text     string    `toml:"text"`
	Choices  []Choice  `toml:"choices,omitempty"`
	Enemies  []*Enemy  `toml:"enemies,omitempty"`
	Outcomes []Outcome `toml:"outcomes,omitempty"`
}

// Choice は選択肢を表す
type Choice struct {
	Description        string            `toml:"description"`
	NextNodeID         string            `toml:"next_node_id"`
	RequiredDiscipline string            `toml:"required_discipline,omitempty"`
	RequiredItem       string            `toml:"required_item,omitempty"`
	Conditions         map[string]string `toml:"conditions,omitempty"`
}

// Enemy は戦闘の敵キャラクター
type Enemy struct {
	Name string `toml:"Name"`
	HP   int    `toml:"HP"`
	CS   int    `toml:"CS"`
}

// Outcome は遭遇戦の結果と次に進むノードを表す
type Outcome struct {
	Description  string `toml:"description,omitempty"`
	Condition    string `toml:"condition,omitempty"`
	ConditionInt []int  `toml:"condition_int,omitempty"`
	NextNodeID   string `toml:"next_node_id"`
}

/*
// こちらの方が良いのでは？
type Player struct {
	Stats      map[string]int
	Attributes map[string]bool
	Equipments *Equipment
	Gold int
}

type Equipment struct {
	Head     *Armor
	Body     *Armor
	Weapon1  *Weapon
	Weapon2  *Weapon
	Shield   bool
	Backpack []*Item
}

type Inventory interface {
	Get(gs *GameState)
	Use(gs *GameState) //装備品の場合は装備を行う。アイテムは自動使用だけど便宜上設定
	Drop(gs *GameState)
}

type Weapon struct {
	Name    string //WeaponSkill判定に使用予定
	Slot    string //Weapon1 Weapon2
	CSBonus int    //いるかなぁ？
}

type Armor struct {
	Name    string
	Slot    string //装備箇所
	HPBonus int
}

type Item struct {
	Name   string
	Slot   string //Backpack Porch?
	Effect string
}

*/

// GameConfig はゲーム全体のTOML設定を表す
type GameConfig struct {
	System string                 `toml:"system"`
	Player map[string]interface{} `toml:"player"`
	Nodes  []Node                 `toml:"nodes"`
}

// GameState はゲームの状態を保持
type GameState struct {
	Player        *lonewolf.Player
	CurrentNodeID string
	Nodes         map[string]Node
	Reader        *bufio.Reader
	System        GameSystem // System フィールドを追加
}

// display_status はプレイヤーの状態を表示
func (gs *GameState) DisplayStatus() {
	fmt.Println("--- ステータス ---")

	// gs.Player が nil でないことを確認
	if gs.Player == nil {
		fmt.Println("プレイヤーデータが初期化されていません。")
		fmt.Println("--- ステータス ---")
		return // プレイヤーが nil なら、これ以上処理しない
	}

	// Stats の表示
	fmt.Println("能力値:") // "Stats" を「能力値」に変更
	if gs.Player.Stats != nil {
		for stat, value := range gs.Player.Stats {
			fmt.Printf("  %s: %d\n", stat, value)
		}
	} else {
		fmt.Println("  能力値データがありません。")
	}

	// Attributes の表示
	fmt.Println("属性:") // "Attribute" を「属性」に変更
	if gs.Player.Attributes != nil {
		foundAttribute := false
		for attr, active := range gs.Player.Attributes {
			if active {
				fmt.Printf("  - %s\n", attr)
				foundAttribute = true
			}
		}
		if !foundAttribute {
			fmt.Println("  有効な属性がありません。")
		}
	} else {
		fmt.Println("  属性データがありません。")
	}

	// Inventory の表示
	fmt.Println("インベントリ:")
	if //gs.Player.Inventory != nil &&
	len(gs.Player.Inventory) > 0 {
		for _, item := range gs.Player.Inventory {
			fmt.Printf("  - %s\n", item)
		}
	} else {
		fmt.Println("  アイテムがありません。")
	}

	// Equipment の表示
	fmt.Println("装備:") // "Equipment" を「装備」に変更
	if gs.Player.Equipment != nil {
		if len(gs.Player.Equipment) == 0 {
			fmt.Println("  装備品がありません。")
		} else {
			for slot, item := range gs.Player.Equipment {
				fmt.Printf("  %s: %s\n", slot, item)
			}
		}
	} else {
		fmt.Println("  装備データがありません。")
	}
	fmt.Println("--- ステータス ---")
}

// Run はゲームループを開始
func (gs *GameState) Run() {
	gs.System.MakingPlayer(gs)
	for {
		node, exists := gs.Nodes[gs.CurrentNodeID]
		if !exists {
			fmt.Println("\nエラー: 存在しないノードIDに到達しました:", gs.CurrentNodeID)
			break
		}

		if err := gs.System.HandleNode(gs, node); err != nil { // gs.Config.System → gs.System
			fmt.Println("エラー:", err)
			break
		}

		if node.Type == "end" {
			fmt.Println("ゲーム終了。")
			break
		}
	}
}
